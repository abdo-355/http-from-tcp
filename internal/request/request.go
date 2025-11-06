package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// split the message to different lines and make sure it has
	// the minimum amount of lines for a valid http/1.1 message which might look like this
	// GET / HTTP/1.1\r\n
	// Host: example.com\r\n
	// \r\n
	dataLines := strings.Split(string(data), "\r\n")
	if len(dataLines) < 3 {
		return nil, fmt.Errorf("Invalid http/1.1 message\nrecieved:%s", string(data))
	}

	requestLineStr := dataLines[0]

	rl, err := parseRequestLine(requestLineStr)
	if err != nil {
		return nil, err
	}

	req := Request{
		RequestLine: rl,
	}

	return &req, nil
}

func parseRequestLine(d string) (RequestLine, error) {
	// make sure it's a valid request line
	parts := strings.Split(d, " ")
	if len(parts) < 3 {
		return RequestLine{}, fmt.Errorf("Invalid request line\nreceived:%s", d)
	}

	// the first part of the request line must be upper case
	if !isAllUpper(parts[0]) {
		return RequestLine{}, fmt.Errorf("Invalid request line\nreceived:%s", d)
	}

	// make sure it's http/1.1
	httpVersion := strings.Split(parts[2], "/")[1]
	if httpVersion != "1.1" {
		return RequestLine{}, fmt.Errorf("unsupported http version make sure it's http/1.1\version received:%s", parts[2])
	}

	return RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: parts[1],
		Method:        parts[0],
	}, nil
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
