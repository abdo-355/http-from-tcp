package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	// read the file
	file, err := os.Open("./messages.txt")
	if err != nil {
		log.Fatal("error reading the file:", err)
	}

	lines := getLinesChannel(file)

	for line := range lines {
		fmt.Println("read:", line)
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
