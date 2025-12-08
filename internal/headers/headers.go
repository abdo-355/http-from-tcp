// Package headers for all header related operations
package headers

import (
	"bytes"
	"fmt"
	"strings"
)

const CRLF = "\r\n"

type Headers struct {
	M map[string]string
}

func NewHeaders() Headers {
	return Headers{M: make(map[string]string)}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(CRLF))

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

	key = strings.ToLower(key)
	if InvalidHeaderFieldName(key) {
		return 0, false, fmt.Errorf("%s has invalid field name characters", key)
	}

	// first check if the key exists and if it does just concatenate the value
	if h.M[key] == "" {
		h.M[key] = strings.TrimSpace(parts[1])
	} else {
		h.M[key] = h.M[key] + ", " + strings.TrimSpace(parts[1])
	}

	return len(data[:idx]) + 2, false, nil
}

func InvalidHeaderFieldName(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]

		switch {
		case c >= 'a' && c <= 'z':
			continue
		case c >= 'A' && c <= 'Z':
			continue
		case c >= '0' && c <= '9':
			continue
		case c == '!' || c == '#' || c == '$' || c == '%' || c == '&':
			continue
		case c == '\'' || c == '*' || c == '+' || c == '-' || c == '.':
			continue
		case c == '^' || c == '_' || c == '`' || c == '|' || c == '~':
			continue
		default:
			// If we reach here, the character is invalid
			return true
		}
	}
	return false
}

func (h *Headers) Get(key string) string {
	// it should be case insensitive
	key = strings.ToLower(key)

	return h.M[key]
}

func (h *Headers) Len() int {
	return len(h.M)
}
