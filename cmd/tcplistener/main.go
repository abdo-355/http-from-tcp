// Command tcplistener listens for TCP connections on port 8080,
// parses HTTP requests from them, and prints the request details.
package main

import (
	"fmt"
	"log"
	"net"

	"github.com/abdo-355/http-from-tcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("an error occurend when opening tcp connection:", err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting a connection on the listener:", err)
		}
		fmt.Println("connection accepted")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Request line:")
		fmt.Println("- Method:", req.RequestLine.Method)
		fmt.Println("- Target:", req.RequestLine.RequestTarget)
		fmt.Println("- Version:", req.RequestLine.HTTPVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers.M {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
	}
}
