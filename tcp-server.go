package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"
	//"strings"
	"encoding/json"
	//"github.com/tonnerre/golang-pretty"
	"github.com/cbehopkins/countdown/cnt_slv"
)

func main() {

	fmt.Println("Launching server...")

	// listen on all interfaces
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Printf("Listen error: %v\n", err)
	}

	for {
		// accept connection on port
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
		} else {
			go HandleConnection(conn)
		}
	}
}

// FIXME tidy this up
type UmNetStruct struct {
	UseMult    bool  `json:"mul"`
	PostResult bool  `json:"post"`
	Val        []int `json:"int"`
}

func UnmarshallNet(input []byte) (result []int, useMult bool, postResult bool) {
	bob := &UmNetStruct{}
	bob.Val = make([]int, 0, 6)
	result = make([]int, 0, 6)
	err := json.Unmarshal(input, &bob)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	useMult = bob.UseMult
	postResult = bob.PostResult
	for _, j := range bob.Val {
		//fmt.Printf("Value of %d\n", j)
		result = append(result, j)
	}
	return
}
func HandleConnection(conn net.Conn) {
	var bob cntSlv.NumCol
	var proofList cntSlv.SolLst
	var foundValues *cntSlv.NumMap

	bob = cntSlv.NumCol{}
	proofList = cntSlv.SolLst{}
	foundValues = cntSlv.NewNumMap(&proofList) //pass it the proof list so it can auto-check for validity at the end
	nullBa, err := foundValues.MarshalJson()
	if err != nil {
		fmt.Printf("Marshalling Error")
	} else {
		//fmt.Println(string(byte_array))
	}
	nullResp := string(nullBa)

	// run loop forever (or until ctrl-c)
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Connection with client closed\n")
				foundValues.LastNumMap()
				bob = cntSlv.NumCol{}
				proofList = cntSlv.SolLst{}
				foundValues = &cntSlv.NumMap{}
				//runtime.GC()
				debug.FreeOSMemory()
				return
			}
			fmt.Printf("Connection read error: %v\n", err)
			return
		}
		// output message received
		//fmt.Print("Message Received:", string(message))
		intArray := []int{}
		intArray, useMult, postResult := UnmarshallNet([]byte(message))

		foundValues.UseMult = useMult
		for _, j := range intArray {
			//fmt.Printf("Adding Value of %d\n", j)
			if j == 0 {
				log.Fatal("0 as input")
			}
			bob.AddNum(j, foundValues)
		}
		if len(intArray) > 1 {
			cntSlv.WorkN(bob, foundValues)
		}

		// If we are not postponing the return of the result
		// Then send a message back with the results of the work so far
		if !postResult {
			foundValues.LastNumMap()
			byteArray, err := foundValues.MarshalJson()
			if err != nil {
				fmt.Printf("Marshalling Error")
			} else {
				//fmt.Println(string(byte_array))
			}
			newmessage := string(byteArray)
			// send new string back to client
			_, err = conn.Write([]byte(newmessage + "\n"))
			if err != nil {
				fmt.Printf("Connection Write error: %v\n", err)
				return
			}

			// Now set us up for the next run
			proofList = cntSlv.SolLst{}
			foundValues = cntSlv.NewNumMap(&proofList) //pass it the proof list so it can auto-check for validity at the end
		} else {
			//fmt.Println("Sending Null response:", null_resp)
			_, err = conn.Write([]byte(nullResp + "\n"))
			if err != nil {
				fmt.Printf("Connection Write error: %v\n", err)
				return
			}
		}
		bob = cntSlv.NumCol{}
	}
}
