// Command udpserver is a UDP client that reads input from stdin
// and sends it as UDP packets to localhost:42069.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal("failed to ResolveUDPAddr:", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("failed to dial UDP:", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("something went wrong reading the input:", err)
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Println("an error occured when writing to the connection", err)
		}
	}

}
