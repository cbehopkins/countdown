package cntSlv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"sync"

	"github.com/cbehopkins/permutation"
)

type UmNetStruct struct {
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
	bob := UmNetStruct{Val: valArray, UseMult: useMult, PostResult: true}
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
		fv.MergeJson(message)
	}

	ps.netChannels <- conn
	ps.channelTokens <- true // Now we're done, add a token to allow another to start
	ps.coallateDone <- true
}
func (ps *permStruct) workerNetClose(fv *NumMap) {
	bob := UmNetStruct{PostResult: false}
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
			fv.MergeJson(message)
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

func (pstrct *permStruct) setupConns(fv *NumMap) (extraTokens int, allFail bool) {
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
					pstrct.netChannels <- conn
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
func (pstrct *permStruct) doneControl() {
	for i := 0; i < pstrct.numPermutations; i++ {
		<-pstrct.coallateDone
	}
	if pstrct.permuteMode == NetMap {
		// Send a message to all the channels to close them down
		// and collect the results
		//fmt.Println("all permutes finished, closing channels")
		pstrct.workerNetClose(pstrct.fv)
		//fmt.Println("Network close finished")
	}
	close(pstrct.coallateChan)
	close(pstrct.mapMergeChan)
	pstrct.mwg.Done()
}
func (pstrct *permStruct) Workers(proofList chan SolLst) {
	// Launch the thing what will actually do the work
	go pstrct.Work()
	pstrct.mwg.Add(2)
	if pstrct.permuteMode == ParMap {
		pstrct.mwg.Add(1)
		go pstrct.mergeFuncWorker()
	}

	// This will run until Work and workers it spawned are complete then Done mwg
	go pstrct.doneControl()
	// Thsi will if needed merge together the resuls and then Done mwg
	go pstrct.outputMerge(proofList)
	// wait for all then Done on mwg
	pstrct.Wait()
}
func (pstrct *permStruct) outputMerge(proofList chan SolLst) {
	for v := range pstrct.coallateChan {
		v.RemoveDuplicates()
		if proofList != nil {
			proofList <- v
		}
	}
	if proofList != nil {
		close(proofList)
	}
	pstrct.mwg.Done()
}
func (pstrct *permStruct) mergeFuncWorker() {
	mergeReport := false // Turn off reporting of new numbers for first run
	for v := range pstrct.mapMergeChan {
		pstrct.fv.Merge(v, mergeReport)
		mergeReport = true
	}
	pstrct.mwg.Done()
}
func (pstrct *permStruct) Wait() {
	pstrct.mwg.Wait()
}
func (pstrct *permStruct) NumWorkers(cnt int) {
	if pstrct.permuteMode == NetMap {
		extraTokens, allFail := pstrct.setupConns(pstrct.fv)
		cnt += extraTokens
		if allFail {
			pstrct.SetPM(LonMap)
		}
	}
	for i := 0; i < cnt; i++ {
		pstrct.channelTokens <- true
	}
}
func (pstrct *permStruct) Launch(bob NumCol) {
	if pstrct.permuteMode == ParMap {
		go pstrct.workerPar(bob, pstrct.fv)
	}
	if pstrct.permuteMode == LonMap {
		go pstrct.workerLone(bob, pstrct.fv)
	}
	if pstrct.permuteMode == NetMap {
		go pstrct.workerNetSend(bob, pstrct.fv)
	}
}
func (pstrct *permStruct) Work() {
	p := pstrct.p
	for result, err := p.Next(); err == nil; result, err = p.Next() {
		// To control the number of workers we run at once we need to grab a token
		// remember to return it later
		<-pstrct.channelTokens
		fmt.Printf("%3d permutation: left %3d, GoRs %3d\r", p.Index()-1, p.Left(), runtime.NumGoroutine())
		bob, ok := result.(NumCol)
		if !ok {
			log.Fatalf("Error Type conversion problem")
		}
		pstrct.Launch(bob)
	}
}
func (pstrct *permStruct) SetPM(val int) {
	pstrct.permuteMode = val
}
func RunPermute(arrayIn NumCol, foundValues *NumMap, proofList chan SolLst) {
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
	go RunPermute(arrayIn, foundValues, returnProofs)
	return returnProofs
}
