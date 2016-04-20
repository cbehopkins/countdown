package main

import (
"net"
"fmt"
"io"
"bufio"
"strings"
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
    // sample process for string received
    newmessage := strings.ToUpper(message)
    // send new string back to client
    _,err = conn.Write([]byte(newmessage + "\n"))
    if (err != nil) {
        fmt.Printf("Connection Write error: %v\n", err)
	return
    }
  }
}
