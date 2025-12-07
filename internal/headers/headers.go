// Package headers for all header related operations
package headers

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/abdo-355/http-from-tcp/internal/request"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(request.CRLF))

	if idx == -1 {
		return 0, false, nil
	}

	// if the CRLF is at the start this means we finished the headers part
	if idx == 0 {
		return 2, true, nil
	}

	cleanedStr := string(bytes.TrimSpace(data[:idx]))
	parts := strings.SplitN(cleanedStr, ":", 2)

	key := parts[0]

	// make sure there are no spaces at the end of the key
	if parts[0] != strings.TrimSpace(parts[0]) {
		return 0, false, fmt.Errorf("invalid header structure. found a space between the header key and the colon")
	}

	h[key] = strings.TrimSpace(parts[1])

	return len(data[:idx]) + 2, false, nil
}
