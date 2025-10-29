package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("an error occurend when opening tcp connection:", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error accepting a connection on the listener:", err)
		}
		fmt.Println("connection accepted")
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		eightBytes := make([]byte, 8)
		currLine := ""
		for {
			_, err := f.Read(eightBytes)
			if errors.Is(err, io.EOF) {
				if currLine != "" {
					lines <- currLine
				}
				break
			}

			// this slice will only have either one or two elements
			splitStr := strings.Split(string(eightBytes), "\n")

			// add the first item to the current line
			currLine += splitStr[0]
			// then check if there's a second item which means there is a new line
			// which means to print the old line and reset the currLine with the second item
			if len(splitStr) == 2 {
				lines <- currLine
				currLine = splitStr[1]
			}
		}

		close(lines)
	}()

	return lines
}
