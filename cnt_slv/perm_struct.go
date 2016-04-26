package cnt_slv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)
type UmNetStruct struct {
        UseMult bool  `json:"mul"`
        PostResult bool `json:"post,omitempty"`	// Postpone sending of the result
        Val     []int `json:"int"`
}

type perm_struct struct {
	channel_tokens chan bool
	net_channels   chan net.Conn
	coallate_chan  chan SolLst
	coallate_done  chan bool
	map_merge_chan chan *NumMap
}

func new_perm_struct(net_it bool) *perm_struct {
	itm := new(perm_struct)
	itm.channel_tokens = make(chan bool, 512)
	if net_it {
		itm.net_channels = make(chan net.Conn, 512)
	}
	itm.coallate_chan = make(chan SolLst, 200)
	itm.coallate_done = make(chan bool, 8)
	itm.map_merge_chan = make(chan *NumMap)
	return itm
}
func (ps *perm_struct) worker_par(it NumCol, fv *NumMap) {
	// This is the parallel worker function
	// It creates a new number map, populates it by working the incoming number set
	// then merges the number map back into the main numbermap
	// This is useful if we have more processes than we know what to do with

	//////////
	// Check if already solved
	fv.const_lk.RLock()
	if fv.Solved {
		fv.const_lk.RUnlock()
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}

	//////////
	// Create the data structures needed to run this set of numbers
	var arthur *NumMap
	var prfl SolLst
	arthur = NewNumMap(&prfl) //pass it the proof list so it can auto-check for validity at the en
	arthur.UseMult = fv.UseMult
	arthur.SelfTest = fv.SelfTest
	arthur.SeekShort = fv.SeekShort
	fv.const_lk.RUnlock()

	//////////
	// Run the compute
	prfl = work_n(it, arthur)
	arthur.LastNumMap()

	//////////
	// Now send the results
	//coallate_chan <- prfl
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start
	ps.map_merge_chan <- arthur
	ps.coallate_done <- true
}
func (ps *perm_struct) worker_lone(it NumCol, fv *NumMap) {
	fv.const_lk.RLock()
	if fv.Solved {
		fv.const_lk.RUnlock()
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}
	fv.const_lk.RUnlock()
	ps.coallate_chan <- work_n(it, fv)
	ps.coallate_done <- true
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start

}
func (ps *perm_struct) worker_net(it NumCol, fv *NumMap) {
	fv.const_lk.RLock()
	use_mult := fv.UseMult
	if fv.Solved {
		fv.const_lk.RUnlock()
		ps.coallate_done <- true
		ps.channel_tokens <- true
		return
	}
	fv.const_lk.RUnlock()
	val_array := make([]int, len(it))
	for i, j := range it {
		val_array[i] = j.Val
	}
	//////////
	// Take our array of numbers (val_array)
	// and turnt hem into an json request ready to send to the network
	bob := UmNetStruct{Val: val_array, UseMult: use_mult}
	text, err := json.Marshal(bob)

	//////////
	// Now send to an open connection
	conn := <-ps.net_channels // Grab the connection for as little time as possible
	fmt.Fprintf(conn, string(text)+"\n")

	//////////
	// listen for reply on open connection
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Read String error: %v\n", err)
	}
	ps.net_channels <- conn

	//////////
	// Take the message text we've got back and interpret it
	fv.MergeJson(message)
	// Not applicable for Net Mode
	//coallate_chan <- work_n(it, fv)

	ps.coallate_done <- true
	ps.channel_tokens <- true // Now we're done, add a token to allow another to start
}
func (pstrct *perm_struct) setup_conns() (extra_tokens int, all_fail bool) {
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
			for i := 0; i < 4; i++ { // Allow 4 connections per server
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
