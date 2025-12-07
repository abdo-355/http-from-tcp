package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	testCases := []struct {
		name            string
		initialHeaders  Headers
		data            []byte
		expectedHeaders Headers
		expectError     bool
	}{
		{
			name:            "Valid single header",
			initialHeaders:  NewHeaders(),
			data:            []byte("Host: localhost:42069\r\n\r\n"),
			expectedHeaders: Headers{"Host": "localhost:42069"},
			expectError:     false,
		},
		{
			name:            "Valid single header with extra whitespace",
			initialHeaders:  NewHeaders(),
			data:            []byte(" Host: localhost:42069 \r\n\r\n"),
			expectedHeaders: Headers{"Host": "localhost:42069"},
			expectError:     false,
		},
		{
			name:            "Valid 2 headers with existing headers",
			initialHeaders:  Headers{"host": "localhost:42069"},
			data:            []byte("User-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"),
			expectedHeaders: Headers{"host": "localhost:42069", "User-Agent": "curl/7.81.0", "Accept": "*/*"},
			expectError:     false,
		},
		{
			name:            "Valid done",
			initialHeaders:  NewHeaders(),
			data:            []byte("\r\n a bunch of other stuff"),
			expectedHeaders: Headers{},
			expectError:     false,
		},
		{
			name:            "Invalid spacing header",
			initialHeaders:  NewHeaders(),
			data:            []byte("       Host : localhost:42069       \r\n\r\n"),
			expectedHeaders: Headers{},
			expectError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			headers := tc.initialHeaders
			remaining := tc.data
			done := false
			var err error
			for !done {
				var n int
				n, done, err = headers.Parse(remaining)
				if err != nil {
					break
				}
				remaining = remaining[n:]
			}
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedHeaders, headers)
			}
		})
	}
}
