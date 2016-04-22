package main

import  (
"net"
"fmt"
"bufio"
"os"
"encoding/xml"
"github.com/cbehopkins/countdown/cnt_slv"
)
type UmNetStruct struct {
//        XMLName   xml.Name `xml:"s,omitempty"`
	XMLName   xml.Name `xml:"work"`
        Val []int  `xml:"int"`                                                                                                                                                                                                       
}

func main() {

  // connect to this socket
  conn, err := net.Dial("tcp", "127.0.0.1:8081")
  if (err != nil) {
	fmt.Printf("Dial error: %v\n", err)
  }
	
  for { 
    // read in input from stdin
    reader := bufio.NewReader(os.Stdin)
    waiting_input := true
    val_array := make([]int, 0,6)
    cnt:=0
    for waiting_input {
      fmt.Print("Enter a Number:")
      text, err := reader.ReadString('\n')
      if (err != nil) {
          fmt.Printf("Text error: %v\n", err)
          return
      }
      if text=="\n"{
	fmt.Println("Blank input, Done")
	 waiting_input = false
      } else {
	var i int
        _, erri := fmt.Sscanf(text,"%d\n", &i)
	if (erri != nil) {
          fmt.Printf("Txt error: %v\n", erri)
        } else {
		fmt.Println("Adding Integer: ", i)
		val_array = append(val_array, i)
                cnt++
                if cnt==6 {waiting_input = false}
        }
      }
    }
    //fmt.Println("Marshalling")
    bob := UmNetStruct{Val:val_array}
    text,err := xml.Marshal(bob)
    //fmt.Println("It's", string(text))
    // send to socket
    fmt.Fprintf(conn, string(text) + "\n")
    //fmt.Println("Sent")
    // listen for reply
    message, err := bufio.NewReader(conn).ReadString('\n')
    if (err != nil) {
        fmt.Printf("Read String error: %v\n", err)
    }
    //fmt.Print("Message from server: "+message)
    cnt_slv.ImportXml(message)
  }
}
