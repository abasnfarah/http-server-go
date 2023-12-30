package main

import (
	"fmt"
	"strings"
	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

  c, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

  defer c.Close()

  request := make([]byte, 1024)
  d, err := c.Read(request)
  fmt.Printf("READ: Number of bytes recieved: %d\n", d)
  if err != nil {
    fmt.Println("Error reading from connection: ", err.Error())
    os.Exit(1)
  }
  fmt.Println("Received message: \r\n", string(request))

  requestLines := strings.Split(string(request), "\r\n")

  startLineSections := strings.Split(requestLines[0], " ")
  HTTPMethod := startLineSections[0]
  path := startLineSections[1]
  HTTPVersion := startLineSections[2]

  HTTPHeaders := requestLines[1:len(requestLines)-2]

  fmt.Println("HTTP Method: ", HTTPMethod)
  fmt.Println("Path: ", path)
  fmt.Println("HTTP Version: ", HTTPVersion)
  fmt.Println("HTTP Headers: ", HTTPHeaders)

  for _, header := range HTTPHeaders {
    fmt.Printf("Header: %s\n", header)
  }
  
  successful := []byte("HTTP/1.1 200 OK\r\n\r\n")
  unSuccessful := []byte("HTTP/1.1 404 Not Found\r\n\r\n")

  if path != "/" {
    d, err = c.Write(unSuccessful)
  }else {
    d, err = c.Write(successful)
  }

  if err != nil {
    fmt.Println("Error writing to connection: ", err.Error())
    os.Exit(1)
  }
  fmt.Printf("WRITE: Number of bytes sent: %d\n", d)

}
