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
	// FastMap - actually slower but uses arrays for much lower memory usage
	FastMap
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
	p               *permutation.Permutator
	numPermutations int
	permuteMode     int
	fv              *NumMap
	channelTokens   chan bool
	netChannels     chan net.Conn
	coallateChan    chan SolLst
	coallateDone    chan bool
	mapMergeChan    chan *NumMap
	activeConns     sync.WaitGroup
	mwg             *sync.WaitGroup
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
	itm.numPermutations = p.Left()
	itm.channelTokens = make(chan bool, 512)
	if foundValues.PermuteMode == NetMap {
		itm.netChannels = make(chan net.Conn, 512)
	}
	itm.coallateChan = make(chan SolLst, 200)
	itm.coallateDone = make(chan bool, 8)
	itm.mapMergeChan = make(chan *NumMap)
	itm.mwg = new(sync.WaitGroup)
	return itm
}
func (ps *permStruct) workerPar(it NumCol, fv *NumMap) {
	// This is the parallel worker function
	// It creates a new number map, populates it by working the incoming number set
	// then merges the number map back into the main numbermap
	// This is useful if we have more processes than we know what to do with

	//////////
	// Check if already solved

	if fv.Solved() {
		ps.coallateDone <- true
		ps.channelTokens <- true
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
	workN(it, arthur, false)
	arthur.LastNumMap()

	//////////
	// Now send the results
	//coallate_chan <- prfl
	ps.channelTokens <- true // Now we're done, add a token to allow another to start
	ps.mapMergeChan <- arthur
	ps.coallateDone <- true
}
func (pr Proofs) addProofsNm(fv *NumMap) {
	for i, v := range pr {
		proofTxt := v.String()
		if proofTxt != "" {
			numP := parseString(proofTxt)
			if numP == nil {
				log.Fatal("Failed to parse", proofTxt)
			}
			if numP.Val == i {
				fv.Add(i, numP)
			} else {
				log.Fatal("Proved wrong value", i, numP.Val, proofTxt, numP)
			}
		}
	}
}

// createPl is used to create a new proof list from a NumCol
// i.e. fudge between the two formats
func (it NumCol) createPl() *proofLst {
	inP := newProofLst(0)
	for _, v := range it.Values() {
		inP.Init(v)
	}
	return inP
}
func (ps *permStruct) workerFast(it NumCol, fv *NumMap) {
	if !fv.Solved() {

		inP := it.createPl()
		// Get a data structure to put the result into
		proofs := getProofs()
		// Populate it
		proofs.wrkFast(*inP)
		// Now convert the result into something we can use
		proofs.addProofsNm(fv)

	}
	ps.coallateDone <- true
	ps.channelTokens <- true
}
func (ps *permStruct) workerLone(it NumCol, fv *NumMap) {
	if !fv.Solved() {
		ps.coallateChan <- workN(it, fv, false)
	}
	ps.coallateDone <- true
	ps.channelTokens <- true // Now we're done, add a token to allow another to start
}
func (ps *permStruct) workerNetSend(it NumCol, fv *NumMap) {
	fv.constLk.RLock()
	useMult := fv.UseMult
	fv.constLk.RUnlock()
	if fv.Solved() {
		ps.coallateDone <- true
		ps.channelTokens <- true
		return
	}

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
	conn := <-ps.netChannels // Grab the connection for as little time as possible
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
	ps.channelTokens <- true // Now we're done, add a token to allow another to start
	ps.coallateDone <- true
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
	} else {

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
	}
	if !netSuccess {
		fmt.Println("Failed to connect to any servers")
		allFail = true
	}
	return
}

// This little go function waits for all the procs to have a done channel and then closes the channel
func (ps *permStruct) doneControl() {
	for i := 0; i < ps.numPermutations; i++ {
		<-ps.coallateDone
	}
	if ps.permuteMode == NetMap {
		// Send a message to all the channels to close them down
		// and collect the results
		//fmt.Println("all permutes finished, closing channels")
		ps.workerNetClose(ps.fv)
		//fmt.Println("Network close finished")
	}
	close(ps.coallateChan)
	close(ps.mapMergeChan)
	ps.mwg.Done()
}
func (ps *permStruct) Workers(proofList chan SolLst) {
	// Launch the thing what will actually do the work
	go ps.Work()
	ps.mwg.Add(2)
	if ps.permuteMode == ParMap {
		ps.mwg.Add(1)
		go ps.mergeFuncWorker()
	}

	// This will run until Work and workers it spawned are complete then Done mwg
	go ps.doneControl()
	// Thsi will if needed merge together the resuls and then Done mwg
	go ps.outputMerge(proofList)
	// wait for all then Done on mwg
	ps.Wait()
}
func (ps *permStruct) outputMerge(proofList chan SolLst) {
	for v := range ps.coallateChan {
		v.RemoveDuplicates()
		if proofList != nil {
			proofList <- v
		}
	}
	if proofList != nil {
		close(proofList)
	}
	ps.mwg.Done()
}
func (ps *permStruct) mergeFuncWorker() {
	mergeReport := false // Turn off reporting of new numbers for first run
	for v := range ps.mapMergeChan {
		ps.fv.Merge(v, mergeReport)
		mergeReport = true
	}
	ps.mwg.Done()
}
func (ps *permStruct) Wait() {
	ps.mwg.Wait()
}
func (ps *permStruct) NumWorkers(cnt int) {
	if ps.permuteMode == NetMap {
		extraTokens, allFail := ps.setupConns(ps.fv)
		cnt += extraTokens
		if allFail {
			ps.SetPM(FastMap)
		}
	}
	for i := 0; i < cnt; i++ {
		ps.channelTokens <- true
	}
}
func (ps *permStruct) Launch(bob NumCol) {

	switch ps.permuteMode {
	case ParMap:
		go ps.workerPar(bob, ps.fv)
	case LonMap:
		go ps.workerLone(bob, ps.fv)
	case NetMap:
		go ps.workerNetSend(bob, ps.fv)
	case FastMap:
		go ps.workerFast(bob, ps.fv)
	default:
		log.Fatal("Unknown Permute Mode")
	}

}
func (ps *permStruct) Work() {
	p := ps.p
	for result, err := p.Next(); err == nil; result, err = p.Next() {
		// To control the number of workers we run at once we need to grab a token
		// remember to return it later
		<-ps.channelTokens
		//fmt.Printf("%3d permutation: left %3d, GoRs %3d\r", p.Index()-1, p.Left(), runtime.NumGoroutine())
		bob, ok := result.(NumCol)
		if !ok {
			log.Fatalf("Error Type conversion problem")
		}
		ps.Launch(bob)
	}
}
func (ps *permStruct) SetPM(val int) {
	ps.permuteMode = val
}
func runPermute(arrayIn NumCol, foundValues *NumMap, proofList chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to have enough CPUs for this to be an issue
	// However the framework seems to be there
	// TBD make this a comannd line variable

	pstrct := newPermStruct(arrayIn, foundValues)
	requiredTokens := 64
	pstrct.NumWorkers(requiredTokens)
	pstrct.Workers(proofList)
	foundValues.LastNumMap()
}
func permuteN(arrayIn NumCol, foundValues *NumMap) (proofList chan SolLst) {
	returnProofs := make(chan SolLst, 16)
	go runPermute(arrayIn, foundValues, returnProofs)
	return returnProofs
}
