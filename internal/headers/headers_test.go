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
			data:            []byte("HOST: localhost:42069\r\n\r\n"),
			expectedHeaders: Headers{"host": "localhost:42069"},
			expectError:     false,
		},
		{
			name:            "Valid single header with extra whitespace",
			initialHeaders:  NewHeaders(),
			data:            []byte(" HOST: localhost:42069 \r\n\r\n"),
			expectedHeaders: Headers{"host": "localhost:42069"},
			expectError:     false,
		},
		{
			name:            "Valid 2 headers with existing headers",
			initialHeaders:  Headers{"host": "localhost:42069"},
			data:            []byte("USER-AGENT: curl/7.81.0\r\nACCEPT: */*\r\n\r\n"),
			expectedHeaders: Headers{"host": "localhost:42069", "user-agent": "curl/7.81.0", "accept": "*/*"},
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
			data:            []byte("       HOST : localhost:42069       \r\n\r\n"),
			expectedHeaders: Headers{},
			expectError:     true,
		},
		{
			name:            "Valid header with mixed case key",
			initialHeaders:  NewHeaders(),
			data:            []byte("Content-Type: application/json\r\n\r\n"),
			expectedHeaders: Headers{"content-type": "application/json"},
			expectError:     false,
		},
		{
			name:            "Invalid character in header key",
			initialHeaders:  NewHeaders(),
			data:            []byte("H©st: localhost:42069\r\n\r\n"),
			expectedHeaders: Headers{},
			expectError:     true,
		},
		{
			name:            "Valid duplicate headers with initial matching header",
			initialHeaders:  Headers{"set-person": "initial-value"},
			data:            []byte("Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n"),
			expectedHeaders: Headers{"set-person": "initial-value, lane-loves-go, prime-loves-zig, tj-loves-ocaml"},
			expectError:     false,
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

func TestValidHeaderFieldName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid lowercase header",
			input:    "host",
			expected: false,
		},
		{
			name:     "Valid header with hyphen",
			input:    "user-agent",
			expected: false,
		},
		{
			name:     "Valid header with allowed special chars",
			input:    "x-custom-header!",
			expected: false,
		},
		{
			name:     "Invalid header with copyright symbol",
			input:    "H©st",
			expected: true,
		},
		{
			name:     "Invalid header with space",
			input:    "host header",
			expected: true,
		},
		{
			name:     "Invalid header with emoji",
			input:    "host😀",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ValidHeaderFieldName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
