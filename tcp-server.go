package main

import (
"net"
"fmt"
"io"
"bufio"
"log"
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
  if (err != nil) {
        fmt.Printf("Listen error: %v\n", err)
  }

  for {
    // accept connection on port
    conn, err := ln.Accept()
    if (err != nil) {
          fmt.Printf("Accept error: %v\n", err)
    } else {
      go HandleConnection(conn)
    }
  }
}
// FIXME tidy this up
type UmNetStruct struct {
	UseMult	bool `json:"mul"`
	PostResult bool `json:"post"`
	Val []int  `json:"int"`
}
func UnmarshallNet (input []byte) (result []int, use_mult bool, post_result bool){
 	bob := &UmNetStruct{}
	bob.Val  = make([]int, 0,6)
	result = make([]int, 0,6)
	err := json.Unmarshal(input, &bob)
        if err != nil {
                fmt.Printf("error: %v", err)
                return
        }
        use_mult = bob.UseMult
	post_result = bob.PostResult
        for _,j := range bob.Val {
                //fmt.Printf("Value of %d\n", j)
		result = append(result,j)
        }
	return
}
func HandleConnection (conn net.Conn) {
  var bob cnt_slv.NumCol
  var proof_list cnt_slv.SolLst
  var found_values *cnt_slv.NumMap

  bob = cnt_slv.NumCol{}
  proof_list = cnt_slv.SolLst{}
  found_values = cnt_slv.NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end
  null_ba, err := found_values.MarshalJson()
  if (err!=nil){
    fmt.Printf("Marshalling Error")
  } else {
       //fmt.Println(string(byte_array))
  }
  null_resp := string(null_ba)

  // run loop forever (or until ctrl-c)
  for {
    // will listen for message to process ending in newline (\n)
    message, err := bufio.NewReader(conn).ReadString('\n')
    if (err != nil) {
        if (err == io.EOF) {
		fmt.Printf("Connection with client closed\n")
		found_values.LastNumMap()
		bob = cnt_slv.NumCol{}
		proof_list = cnt_slv.SolLst{}
		found_values = &cnt_slv.NumMap{}
		//runtime.GC()
		debug.FreeOSMemory()
		return
	}
        fmt.Printf("Connection read error: %v\n", err)
	return
    }
    // output message received
    //fmt.Print("Message Received:", string(message))
    int_array := []int{}
    int_array, use_mult,post_result := UnmarshallNet([]byte(message))

    found_values.UseMult = use_mult
    for _,j := range int_array {
        //fmt.Printf("Adding Value of %d\n", j)
	if (j==0) {
		log.Fatal("0 as input")
	}
      bob.AddNum(j, found_values)
    }
    if (len(int_array) >1) {
      cnt_slv.WorkN(bob, found_values)
    }
    
    // If we are not postponing the return of the result
    // Then send a message back with the results of the work so far
    if !post_result {
      found_values.LastNumMap()
      byte_array, err := found_values.MarshalJson()
      if (err!=nil){
        fmt.Printf("Marshalling Error")
      } else {
	//fmt.Println(string(byte_array))
      }
      newmessage := string(byte_array)
      // send new string back to client
      _,err = conn.Write([]byte(newmessage + "\n"))
      if (err != nil) {
          fmt.Printf("Connection Write error: %v\n", err)
	  return
      }

      // Now set us up for the next run
      proof_list = cnt_slv.SolLst{}
      found_values = cnt_slv.NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end          
    } else {
	//fmt.Println("Sending Null response:", null_resp)
	_,err = conn.Write([]byte(null_resp + "\n"))
	if (err != nil) {
          fmt.Printf("Connection Write error: %v\n", err)
          return
        }
    }
    bob = cnt_slv.NumCol{}
  }
}

