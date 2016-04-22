package main

import (
"net"
"fmt"
"io"
"bufio"
//"strings"
"encoding/xml"
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
type UmNetStruct struct {
        XMLName   xml.Name `xml:"work"`
	UseMult	bool `xml:"mul"`
	Val []int  `xml:"int"`
}
func UnmarshallNet (input []byte) (result []int, use_mult bool){
 	bob := &UmNetStruct{}
 	//var bob []int `xml:"int"`
	bob.Val  = make([]int, 0,6)
	result = make([]int, 0,6)
	err := xml.Unmarshal(input, &bob)
        if err != nil {
                fmt.Printf("error: %v", err)
                return
        }
        //fmt.Printf("We've been given:\n%s\nand we turn this into:\n", input)
        //pretty.Println(bob)
        use_mult = bob.UseMult
        for _,j := range bob.Val {
                //fmt.Printf("Value of %d\n", j)
		result = append(result,j)
        }
	return
}
func HandleConnection (conn net.Conn) {

  // run loop forever (or until ctrl-c)
  for {
    // will listen for message to process ending in newline (\n)
    message, err := bufio.NewReader(conn).ReadString('\n')
    if (err != nil) {
        if (err == io.EOF) {
		fmt.Printf("Connection with client closed\n")
		return
	}
        fmt.Printf("Connection read error: %v\n", err)
	return
    }
    // output message received
    fmt.Print("Message Received:", string(message))
    int_array, use_mult := UnmarshallNet([]byte(message))
    var bob cnt_slv.NumCol
    var proof_list cnt_slv.SolLst
    found_values := cnt_slv.NewNumMap(&proof_list) //pass it the proof list so it can auto-check for validity at the end
    for _,j := range int_array {
    //  fmt.Printf("Adding Value of %d\n", j)
      bob.AddNum(j, found_values)
    }
    found_values.UseMult = use_mult
    cnt_slv.WorkN(bob, found_values)
    found_values.LastNumMap()
    //found_values.PrintProofs()
    byte_array, err := found_values.MarshalXml()
    if (err!=nil){
      fmt.Printf("Marshalling Error")
    } else {
	//fmt.Println(string(byte_array))
    }
    newmessage := string(byte_array)



    // sample process for string received
    //newmessage := strings.ToUpper(message)
    // send new string back to client
    _,err = conn.Write([]byte(newmessage + "\n"))
    if (err != nil) {
        fmt.Printf("Connection Write error: %v\n", err)
	return
    }
  }
}
