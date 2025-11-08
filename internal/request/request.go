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

	// pass entire request string to parseRequestLine (not just first line)
	rl, err := parseRequestLine(string(data))
	if err != nil {
		return nil, err
	}

	req := Request{
		RequestLine: rl,
	}

	return &req, nil
}

func parseRequestLine(rawRequest string) (RequestLine, error) {
	// Extract first line of the request (the request line)
	endOfRequestLine := strings.Index(rawRequest, "\r\n")
	if endOfRequestLine == -1 {
		return RequestLine{}, fmt.Errorf("Request line not found in request")
	}
	requestLine := rawRequest[:endOfRequestLine]

	// parse the extracted request line as before
	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		return RequestLine{}, fmt.Errorf("Invalid request line\nreceived:%s", requestLine)
	}

	if !isAllUpper(parts[0]) {
		return RequestLine{}, fmt.Errorf("Invalid request line\nreceived:%s", requestLine)
	}

	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) < 2 {
		return RequestLine{}, fmt.Errorf("Invalid HTTP version format\nreceived:%s", parts[2])
	}

	httpVersion := httpVersionParts[1]
	if httpVersion != "1.1" {
		return RequestLine{}, fmt.Errorf("unsupported HTTP version, expected http/1.1, received:%s", parts[2])
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
