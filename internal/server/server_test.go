package server

import (
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/abdo-355/http-from-tcp/internal/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerError_Write(t *testing.T) {
	testCases := []struct {
		name        string
		status      response.StatusCode
		message     string
		expectError bool
	}{
		{
			name:        "Bad Request",
			status:      response.StatusBadRequest,
			message:     "Invalid request",
			expectError: false,
		},
		{
			name:        "Internal Server Error",
			status:      response.StatusInternalServerError,
			message:     "Something went wrong",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			hErr := &HandlerError{
				Status:  tc.status,
				Message: tc.message,
			}

			hErr.Write(buf)

			output := buf.String()
			lines := strings.Split(strings.TrimSpace(output), "\r\n")

			// Check status line
			require.Greater(t, len(lines), 0)
			switch tc.status {
			case response.StatusBadRequest:
				assert.True(t, strings.HasPrefix(lines[0], "HTTP/1.1 400 Bad Request"))
			case response.StatusInternalServerError:
				assert.True(t, strings.HasPrefix(lines[0], "HTTP/1.1 500 Internal Server Error"))
			}

			// Check headers are present
			foundContentLength := false
			foundConnection := false
			foundContentType := false
			for _, line := range lines[1:] {
				if strings.HasPrefix(line, "content-length:") {
					foundContentLength = true
					assert.Equal(t, "content-length: "+strconv.Itoa(len(tc.message)), line)
				} else if strings.HasPrefix(line, "connection: close") {
					foundConnection = true
				} else if strings.HasPrefix(line, "content-type: text/plain") {
					foundContentType = true
				} else if line == "" {
					// End of headers
					break
				}
			}
			assert.True(t, foundContentLength, "content-length header not found")
			assert.True(t, foundConnection, "connection header not found")
			assert.True(t, foundContentType, "content-type header not found")

			// Check body
			bodyStart := strings.Index(output, "\r\n\r\n") + 4
			body := output[bodyStart:]
			assert.Equal(t, tc.message, body)
		})
	}
}
