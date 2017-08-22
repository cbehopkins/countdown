package cnt_slv

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

type perm_struct struct {
	p                *permutation.Permutator
	num_permutations int
	permute_mode     int
	fv               *NumMap
	channel_tokens   chan bool
	net_channels     chan net.Conn
	coallate_chan    chan SolLst
	coallate_done    chan bool
	map_merge_chan   chan *NumMap
	active_conns     sync.WaitGroup
	mwg              *sync.WaitGroup
}

func new_perm_struct(array_in NumCol, found_values *NumMap) *perm_struct {

	p, err := permutation.NewPerm(array_in, lessNumber)
	if err != nil {
		fmt.Println(err)
	}

	itm := new(perm_struct)
	itm.p = p
	itm.fv = found_values
	// Local copy as we may override it
	itm.permute_mode = found_values.PermuteMode
	itm.num_permutations = p.Left()
	itm.channel_tokens = make(chan bool, 512)
	if found_values.PermuteMode == NetMap {
		itm.net_channels = make(chan net.Conn, 512)
	}
	itm.coallate_chan = make(chan SolLst, 200)
	itm.coallate_done = make(chan bool, 8)
	itm.map_merge_chan = make(chan *NumMap)
	itm.mwg = new(sync.WaitGroup)
	return itm
}
func (ps *perm_struct) worker_par(it NumCol, fv *NumMap) {
	// This is the parallel worker function
	// It creates a new number map, populates it by working the incoming number set
	// then merges the number map back into the main numbermap
	// This is useful if we have more processes than we know what to do with

	//////////
	// Check if already solved

	if fv.Solved() {
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}

	//////////
	// Create the data structures needed to run this set of numbers
	var arthur *NumMap
	arthur = NewNumMap() //pass it the proof list so it can auto-check for validity at the en
	fv.const_lk.RLock()
	arthur.UseMult = fv.UseMult
	arthur.SelfTest = fv.SelfTest
	arthur.SeekShort = fv.SeekShort
	fv.const_lk.RUnlock()

	//////////
	// Run the compute
	work_n(it, arthur)
	arthur.LastNumMap()

	//////////
	// Now send the results
	//coallate_chan <- prfl
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start
	ps.map_merge_chan <- arthur
	ps.coallate_done <- true
}
func (ps *perm_struct) worker_lone(it NumCol, fv *NumMap) {
	if fv.Solved() {
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}
	ps.coallate_chan <- work_n(it, fv)
	ps.coallate_done <- true
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start

}
func (ps *perm_struct) worker_net_send(it NumCol, fv *NumMap) {
	fv.const_lk.RLock()
	use_mult := fv.UseMult
	fv.const_lk.RUnlock()
	if fv.Solved() {
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}

	val_array := make([]int, len(it))
	for i, j := range it {
		val_array[i] = j.Val
	}
	//////////
	// Take our array of numbers (val_array)
	// and turn them into an json request ready to send to the network
	bob := UmNetStruct{Val: val_array, UseMult: use_mult, PostResult: true}
	text, err := json.Marshal(bob)
	if err != nil {
		fmt.Printf("Json Marshall error in worker_net_send: %v\n", err)
		return
	}

	//////////
	// Now send to an open connection
	conn := <-ps.net_channels // Grab the connection for as little time as possible
	full_msg := string(text) + "\n"
	//fmt.Printf("Sending::%s", full_msg)
	//n, err:= fmt.Fprintf(conn, full_msg)
	n, err := conn.Write([]byte(full_msg))
	if err != nil {
		//fmt.Printf("Send Error %d in worker_net_send: %v\n", n, err)
		log.Fatal()
		return
	}
	if len(full_msg) != n {
		fmt.Printf("Send Error, Length is %d, sent %d\n", n, len(full_msg))
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

	ps.net_channels <- conn
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start
	ps.coallate_done <- true
}
func (ps *perm_struct) worker_net_close(fv *NumMap) {
	bob := UmNetStruct{PostResult: false}
	text, err := json.Marshal(bob)
	if err != nil {
		fmt.Printf("Json Marshall error in worker_net_close: %v\n", err)
		return // FIXME add error return
	}
	close(ps.net_channels)
	var par_merge sync.WaitGroup
	for conn := range ps.net_channels {
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
		par_merge.Add(1)
		go func() {
			fv.MergeJson(message)
			//err := fv.FastUnMarshalJson([]byte(message))
			//if (err !=nil) {
			//	fmt.Printf("Fast Unmarshall error %v\n", err)
			//	return // FIXME
			//}
			par_merge.Done()
		}()
	}
	for conn := range ps.net_channels {
		// Fixme this can probably be spawned
		conn.Close()
	}
	par_merge.Wait()
}

func (pstrct *perm_struct) setup_conns(fv *NumMap) (extra_tokens int, all_fail bool) {
	net_success := false

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
					net_success = true
					pstrct.net_channels <- conn
					extra_tokens++
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	if !net_success {
		fmt.Println("Failed to connect to any servers")
		all_fail = true
	}
	return
}

// This little go function waits for all the procs to have a done channel and then closes the channel
func (pstrct *perm_struct) done_control() {
	for i := 0; i < pstrct.num_permutations; i++ {
		<-pstrct.coallate_done
	}
	if pstrct.permute_mode == NetMap {
		// Send a message to all the channels to close them down
		// and collect the results
		//fmt.Println("all permutes finished, closing channels")
		pstrct.worker_net_close(pstrct.fv)
		//fmt.Println("Network close finished")
	}
	close(pstrct.coallate_chan)
	close(pstrct.map_merge_chan)
	pstrct.mwg.Done()
}
func (pstrct *perm_struct) Workers(proof_list chan SolLst) {
	// Launch the thing what will actually do the work
	go pstrct.Work()
	pstrct.mwg.Add(2)
	if pstrct.permute_mode == ParMap {
		pstrct.mwg.Add(1)
		go pstrct.merge_func_worker()
	}

	// This will run until Work and workers it spawned are complete then Done mwg
	go pstrct.done_control()
	// Thsi will if needed merge together the resuls and then Done mwg
	go pstrct.output_merge(proof_list)
	// wait for all then Done on mwg
	pstrct.Wait()
}
func (pstrct *perm_struct) output_merge(proof_list chan SolLst) {
	for v := range pstrct.coallate_chan {
		v.RemoveDuplicates()
		if proof_list != nil {
			proof_list <- v
		}
	}
	if proof_list != nil {
		close(proof_list)
	}
	pstrct.mwg.Done()
}
func (pstrct *perm_struct) merge_func_worker() {
	merge_report := false // Turn off reporting of new numbers for first run
	for v := range pstrct.map_merge_chan {
		pstrct.fv.Merge(v, merge_report)
		merge_report = true
	}
	pstrct.mwg.Done()
}
func (pstrct *perm_struct) Wait() {
	pstrct.mwg.Wait()
}
func (pstrct *perm_struct) NumWorkers(cnt int) {
	if pstrct.permute_mode == NetMap {
		extra_tokens, all_fail := pstrct.setup_conns(pstrct.fv)
		cnt += extra_tokens
		if all_fail {
			pstrct.SetPM(LonMap)
		}
	}
	for i := 0; i < cnt; i++ {
		pstrct.channel_tokens <- true
	}
}
func (pstrct *perm_struct) Launch(bob NumCol) {
	if pstrct.permute_mode == ParMap {
		go pstrct.worker_par(bob, pstrct.fv)
	}
	if pstrct.permute_mode == LonMap {
		go pstrct.worker_lone(bob, pstrct.fv)
	}
	if pstrct.permute_mode == NetMap {
		go pstrct.worker_net_send(bob, pstrct.fv)
	}
}
func (pstrct *perm_struct) Work() {
	p := pstrct.p
	for result, err := p.Next(); err == nil; result, err = p.Next() {
		// To control the number of workers we run at once we need to grab a token
		// remember to return it later
		<-pstrct.channel_tokens
		fmt.Printf("%3d permutation: left %3d, GoRs %3d\r", p.Index()-1, p.Left(), runtime.NumGoroutine())
		bob, ok := result.(NumCol)
		if !ok {
			log.Fatalf("Error Type conversion problem")
		}
		pstrct.Launch(bob)
	}
}
func (pstrct *perm_struct) SetPM(val int) {
	pstrct.permute_mode = val
}
func RunPermute(array_in NumCol, found_values *NumMap, proof_list chan SolLst) {
	// If your number of workers is limited by access to the centralmap
	// Then we have the ability to use several number maps and then merge them
	// No system I have access to have enough CPUs for this to be an issue
	// However the framework seems to be there
	// TBD make this a comannd line variable

	//fmt.Println("Start Permute")

	pstrct := new_perm_struct(array_in, found_values)
	required_tokens := 64
	pstrct.NumWorkers(required_tokens)
	pstrct.Workers(proof_list)

	found_values.LastNumMap()
}
func permuteN(array_in NumCol, found_values *NumMap) (proof_list chan SolLst) {
	return_proofs := make(chan SolLst, 16)
	go RunPermute(array_in, found_values, return_proofs)
	return return_proofs
}
