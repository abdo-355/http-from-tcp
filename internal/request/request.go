// Package request handles everything related to parsing the data from the tcp connection
package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/abdo-355/http-from-tcp/internal/headers"
)

type requestState int

const (
	Initialized requestState = iota
	Done
	ParsingHeaders
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers

	state requestState
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

const (
	CRLF       = "\r\n"
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	// to keep track of how much data we've read from the io.Reader into the buffer
	readToIndex := 0
	r := Request{state: Initialized}

	for r.state != Done {
		if readToIndex >= len(buf) {
			biggerBuf := make([]byte, len(buf)*2)
			copy(biggerBuf, buf)
			buf = biggerBuf
		}
		bytesRead, err := reader.Read(buf[readToIndex:])
		if err == io.EOF {
			if r.state == ParsingHeaders {
				return nil, fmt.Errorf("reached EOF before getting the full request")
			}
			break
		}
		readToIndex += bytesRead

		if err != nil {
			return nil, err
		}

		bytesParsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		// Remove the data that was parsed successfully from the buffer (this keeps our buffer small and memory efficient).
		copy(buf, buf[bytesParsed:])
		readToIndex -= bytesParsed
	}

	return &r, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(CRLF))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HTTPVersion:   versionParts[1],
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == Done {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}

	var bytesNum int
	switch r.state {
	case Initialized:
		rl, bn, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bn == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = ParsingHeaders
		bytesNum = bn
		r.Headers = headers.NewHeaders()
	case ParsingHeaders:
		bytesNum = 0
	default:
		return 0, fmt.Errorf("unexpected state")
	}

	for r.state == ParsingHeaders {
		reminingData := data[bytesNum:]
		n, finished, err := r.Headers.Parse(reminingData)
		if err != nil {
			return bytesNum, err
		}
		if n == 0 {
			return bytesNum, nil
		}
		bytesNum += n

		if finished {
			r.state = Done
		}

	}

	// Return total bytes parsed
	return bytesNum, nil
}
