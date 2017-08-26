package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/cbehopkins/countdown/cnt_slv"
	"net"
	"os"
)

type UmNetStruct struct {
	Val []int `json:"int"`
}

func main() {

	// connect to this socket
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		fmt.Printf("Dial error: %v\n", err)
	}

	for {
		// read in input from stdin
		reader := bufio.NewReader(os.Stdin)
		waitingInput := true
		valArray := make([]int, 0, 6)
		cnt := 0
		for waitingInput {
			fmt.Print("Enter a Number:")
			text, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Text error: %v\n", err)
				return
			}
			if text == "\n" {
				fmt.Println("Blank input, Done")
				waitingInput = false
			} else {
				var i int
				_, erri := fmt.Sscanf(text, "%d\n", &i)
				if erri != nil {
					fmt.Printf("Txt error: %v\n", erri)
				} else {
					fmt.Println("Adding Integer: ", i)
					valArray = append(valArray, i)
					cnt++
					if cnt == 6 {
						waitingInput = false
					}
				}
			}
		}
		//////////
		// Take our array of numbers (val_array)
		// and turnt hem into an json request ready to send to the network
		bob := UmNetStruct{Val: valArray}
		text, err := json.Marshal(bob)

		//////////
		// Now send to an open connection
		fmt.Fprintf(conn, string(text)+"\n")

		//////////
		// listen for reply on open connection
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Printf("Read String error: %v\n", err)
		}

		//////////
		// Take the message text we've got back and interpret it
		cntSlv.ImportJson(message) // Import prints the proofs for us - useful for test but not much else
	}
}
