package cntSlv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/cbehopkins/permutation"
)

const (
	// LonMap - one map for all workers
	LonMap = iota
	// ParMap One Map for each worker, then merge them at the end
	ParMap
	// NetMap try and use the Network
	NetMap
)

type umNetStruct struct {
	UseMult    bool  `json:"mul"`
	PostResult bool  `json:"post,omitempty"` // Postpone sending of the result
	Val        []int `json:"int"`
}

type permStruct struct {
	p            *permutation.Permutator
	permuteMode  int
	fv           *NumMap
	permuteChan  chan NumCol
	netChannels  chan net.Conn
	coallateChan chan SolLst
	mapMergeChan chan *NumMap
}

func newPermStruct(arrayIn NumCol, foundValues *NumMap) *permStruct {

	p, err := permutation.NewPerm(arrayIn, lessNumber)
	if err != nil {
		fmt.Println(err)
	}

	itm := new(permStruct)
	itm.p = p
	itm.fv = foundValues
	// Local copy as we may override it
	itm.permuteMode = foundValues.PermuteMode
	if foundValues.PermuteMode == NetMap {
		itm.netChannels = make(chan net.Conn, 512)
	}
	itm.permuteChan = make(chan NumCol)
	itm.coallateChan = make(chan SolLst, 200)
	itm.mapMergeChan = make(chan *NumMap)
	return itm
}
func (ps *permStruct) workerPar(it NumCol, fv *NumMap) {
	// This is the parallel worker function
	// It creates a new number map, populates it by working the incoming number set
	// then merges the number map back into the main numbermap
	// This is useful if we have congestion on the main map

	//////////
	// Check if already solved
	if fv.Solved() {
		return
	}

	//////////
	// Create the data structures needed to run this set of numbers
	var arthur *NumMap
	arthur = NewNumMap() //pass it the proof list so it can auto-check for validity at the en
	fv.constLk.RLock()
	arthur.UseMult = fv.UseMult
	arthur.SelfTest = fv.SelfTest
	arthur.SeekShort = fv.SeekShort
	fv.constLk.RUnlock()

	//////////
	// Run the compute
	workN(it, arthur)
	arthur.LastNumMap()

	//////////
	// Now send the results
	ps.mapMergeChan <- arthur
}

func (ps *permStruct) workerLone(it NumCol, fv *NumMap) {
	if !fv.Solved() {
		ps.coallateChan <- workN(it, fv)
	}
}
func (ps *permStruct) workerNetSend(it NumCol, fv *NumMap) {
	if fv.Solved() {
		return
	}
	fv.constLk.RLock()
	useMult := fv.UseMult
	fv.constLk.RUnlock()

	valArray := make([]int, len(it))
	for i, j := range it {
		valArray[i] = j.Val
	}
	//////////
	// Take our array of numbers (val_array)
	// and turn them into an json request ready to send to the network
	bob := umNetStruct{Val: valArray, UseMult: useMult, PostResult: true}
	text, err := json.Marshal(bob)
	if err != nil {
		fmt.Printf("Json Marshall error in worker_net_send: %v\n", err)
		return
	}

	//////////
	// Now send to an open connection
	conn := <-ps.netChannels // grab the connection for as little time as possible
	fullMsg := string(text) + "\n"
	//fmt.Printf("Sending::%s", full_msg)
	//n, err:= fmt.Fprintf(conn, full_msg)
	n, err := conn.Write([]byte(fullMsg))
	if err != nil {
		//fmt.Printf("Send Error %d in worker_net_send: %v\n", n, err)
		log.Fatal()
		return
	}
	if len(fullMsg) != n {
		fmt.Printf("Send Error, Length is %d, sent %d\n", n, len(fullMsg))
		log.Fatal()
	}
	//////////
	// listen for reply on open connection
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Read String error: %v\n", err)
		return
	}
	//fmt.Println("Received Message from server::")

	//////////
	// Take the message text we've got back and interpret it
	if len(message) > 3 {
		fv.MergeJSON(message)
	}

	ps.netChannels <- conn
}
func (ps *permStruct) workerNetClose(fv *NumMap) {
	bob := umNetStruct{PostResult: false}
	text, err := json.Marshal(bob)
	if err != nil {
		fmt.Printf("Json Marshall error in worker_net_close: %v\n", err)
		return // FIXME add error return
	}
	close(ps.netChannels)
	var parMerge sync.WaitGroup
	for conn := range ps.netChannels {
		// Send a message to each channel to close the connection
		//fmt.Printf("Sending Request for end::" + string(text)+"\n")
		fmt.Fprintf(conn, string(text)+"\n")
		//////////
		// listen for reply on open connection
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("Read String error: %v\n", err)
			return
		}
		//fmt.Println("Received Message from server::")

		//////////
		// Take the message text we've got back and interpret it
		parMerge.Add(1)
		go func() {
			fv.MergeJSON(message)
			//err := fv.FastUnMarshalJson([]byte(message))
			//if (err !=nil) {
			//	fmt.Printf("Fast Unmarshall error %v\n", err)
			//	return // FIXME
			//}
			parMerge.Done()
		}()
	}
	for conn := range ps.netChannels {
		// Fixme this can probably be spawned
		conn.Close()
	}
	parMerge.Wait()
}

func (ps *permStruct) setupConns(fv *NumMap) (extraTokens int, allFail bool) {
	netSuccess := false

	cwd, _ := os.Getwd()
	file, err := os.Open("servers.cfg")
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Network mode disabled, couldn't find servers.cfg in:, " + cwd)
		} else {
			log.Fatal(err)
		}
		allFail = true
		return
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var server string
		server = scanner.Text()
		fmt.Printf("Trying to connect to server at %s\n", server)
		for i := 0; i < 3; i++ { // Allow 4 connections per server
			// connect to a socket
			conn, err := net.Dial("tcp", server)
			if err != nil {
				fmt.Printf("Dial error: %v\n", err)
			} else {
				netSuccess = true
				ps.netChannels <- conn
				extraTokens++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if !netSuccess {
		fmt.Println("Failed to connect to any servers")
		allFail = true
	}
	return
}

// This little go function waits for all the procs to have a done channel and then closes the channel
func (ps *permStruct) doneControl() {
	if ps.permuteMode == NetMap {
		// Send a message to all the channels to close them down
		// and collect the results
		//fmt.Println("all permutes finished, closing channels")
		ps.workerNetClose(ps.fv)
		//fmt.Println("Network close finished")
	}
	close(ps.coallateChan)
	close(ps.mapMergeChan)
}
func (ps *permStruct) Workers(proofList chan SolLst, numWorkers int) {
	go ps.Work() // The thing that generates Permutations to work
	coallateWg := ps.Launch(numWorkers)
	var mwg sync.WaitGroup
	// one thing -  outputMerge - to wait for
	mwg.Add(1)
	if ps.permuteMode == ParMap {
		mwg.Add(1)
		go ps.mergeFuncWorker(&mwg)
	}

	// Thsi will if needed merge together the resuls and then Done mwg
	go ps.outputMerge(proofList, &mwg)
	coallateWg.Wait()
	// This will run until Work and workers it spawned are complete then Done mwg
	ps.doneControl()
	// wait for all then Done on mwg
	mwg.Wait()
}
func (ps *permStruct) outputMerge(proofList chan SolLst, mwg *sync.WaitGroup) {
	for v := range ps.coallateChan {
		v.RemoveDuplicates()
		if proofList != nil {
			proofList <- v
		}
	}
	if proofList != nil {
		close(proofList)
	}
	mwg.Done()
}
func (ps *permStruct) mergeFuncWorker(mwg *sync.WaitGroup) {
	mergeReport := false // Turn off reporting of new numbers for first run
	for v := range ps.mapMergeChan {
		ps.fv.Merge(v, mergeReport)
		mergeReport = true
	}
	mwg.Done()
}

// NumNetWorkers deal with network needing more workers
func (ps *permStruct) NumNetWorkers(cnt int) int {
	if ps.permuteMode == NetMap {
		extraTokens, allFail := ps.setupConns(ps.fv)
		cnt += extraTokens
		if allFail {
			ps.SetPM(LonMap)
		}
	}
	return cnt
}

// Launch a worker
// i.e. spawn the thing that will do the calc
func (ps *permStruct) Launch(cnt int) *sync.WaitGroup {
	var coallateWg sync.WaitGroup
	type workerFunc func(NumCol, *NumMap)
	var wf workerFunc
	switch ps.permuteMode {
	case ParMap:
		wf = ps.workerPar
	case LonMap:
		wf = ps.workerLone
	case NetMap:
		wf = ps.workerNetSend
	default:
		log.Fatal("Unknown Permute Mode")
	}

	runner := func() {
		for bob := range ps.permuteChan {
			wf(bob, ps.fv)
		}
		coallateWg.Done()
	}

	coallateWg.Add(cnt)
	for i := 0; i < cnt; i++ {
		go runner()
	}
	return &coallateWg
}

// Work the permutation struct
// That is get permulations and send them on the
// toWork Chan
func (ps *permStruct) Work() {
	p := ps.p
	for result, err := p.Next(); err == nil; result, err = p.Next() {
		bob, ok := result.(NumCol)
		if !ok {
			log.Fatalf("Error Type conversion problem")
		}
		ps.permuteChan <- bob
	}
	close(ps.permuteChan)
}

// SetPM set the permute mode
func (ps *permStruct) SetPM(val int) {
	ps.permuteMode = val
}

// runPermute runs a permutation across a supplied set of numbers
func runPermute(arrayIn NumCol, foundValues *NumMap, proofList chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to have enough CPUs for this to be an issue
	// However the framework seems to be there

	pstrct := newPermStruct(arrayIn, foundValues)
	numWorkers := 8
	numWorkers = pstrct.NumNetWorkers(numWorkers)
	pstrct.Workers(proofList, numWorkers)
	foundValues.LastNumMap()
}
func permuteN(arrayIn NumCol, foundValues *NumMap) chan SolLst {
	returnProofs := make(chan SolLst, 16)
	go runPermute(arrayIn, foundValues, returnProofs)
	return returnProofs
}
