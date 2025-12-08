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

func growBuffer(buf []byte, offset int) []byte {
	if offset >= len(buf) {
		biggerBuf := make([]byte, len(buf)*2)
		copy(biggerBuf, buf)
		return biggerBuf
	}
	return buf
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	bufferOffset := 0
	r := Request{state: Initialized}

	for r.state != Done {
		buf = growBuffer(buf, bufferOffset)
		bytesRead, err := reader.Read(buf[bufferOffset:])
		if err == io.EOF {
			if r.state == ParsingHeaders {
				return nil, fmt.Errorf("unexpected EOF while parsing headers")
			}
			break
		}
		bufferOffset += bytesRead

		if err != nil {
			return nil, err
		}

		bytesParsed, err := r.parse(buf[:bufferOffset])
		if err != nil {
			return nil, err
		}

		// Remove the data that was parsed successfully from the buffer (this keeps our buffer small and memory efficient).
		copy(buf, buf[bytesParsed:])
		bufferOffset -= bytesParsed
	}

	return &r, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	crlfIndex := bytes.Index(data, []byte(CRLF))
	if crlfIndex == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:crlfIndex])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, crlfIndex + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request-line: %s", str)
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
		return 0, fmt.Errorf("cannot parse data in done state")
	}

	bytesParsed := 0
	if r.state == Initialized {
		rl, bytesConsumed, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}

		if bytesConsumed == 0 {
			return 0, nil
		}

		r.RequestLine = *rl
		r.state = ParsingHeaders
		bytesParsed = bytesConsumed
		r.Headers = headers.NewHeaders()
	}

	for r.state == ParsingHeaders {
		remainingData := data[bytesParsed:]
		n, finished, err := r.Headers.Parse(remainingData)
		if err != nil {
			return bytesParsed, err
		}
		if n == 0 {
			return bytesParsed, nil
		}
		bytesParsed += n

		if finished {
			r.state = Done
		}

	}

	return bytesParsed, nil
}
