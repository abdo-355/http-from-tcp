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
			data:            []byte("HOST: localhost:8080\r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{"host": "localhost:8080"}},
			expectError:     false,
		},
		{
			name:            "Valid single header with extra whitespace",
			initialHeaders:  NewHeaders(),
			data:            []byte(" HOST: localhost:8080 \r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{"host": "localhost:8080"}},
			expectError:     false,
		},
		{
			name:            "Valid 2 headers with existing headers",
			initialHeaders:  Headers{M: map[string]string{"host": "localhost:8080"}},
			data:            []byte("USER-AGENT: curl/7.81.0\r\nACCEPT: */*\r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{"host": "localhost:8080", "user-agent": "curl/7.81.0", "accept": "*/*"}},
			expectError:     false,
		},
		{
			name:            "Valid done",
			initialHeaders:  NewHeaders(),
			data:            []byte("\r\n a bunch of other stuff"),
			expectedHeaders: Headers{M: map[string]string{}},
			expectError:     false,
		},
		{
			name:            "Invalid spacing header",
			initialHeaders:  NewHeaders(),
			data:            []byte("       HOST : localhost:8080       \r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{}},
			expectError:     true,
		},
		{
			name:            "Valid header with mixed case key",
			initialHeaders:  NewHeaders(),
			data:            []byte("Content-Type: application/json\r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{"content-type": "application/json"}},
			expectError:     false,
		},
		{
			name:            "Invalid character in header key",
			initialHeaders:  NewHeaders(),
			data:            []byte("H©st: localhost:8080\r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{}},
			expectError:     true,
		},
		{
			name:            "Valid duplicate headers with initial matching header",
			initialHeaders:  Headers{M: map[string]string{"set-person": "initial-value"}},
			data:            []byte("Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n"),
			expectedHeaders: Headers{M: map[string]string{"set-person": "initial-value, lane-loves-go, prime-loves-zig, tj-loves-ocaml"}},
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

func TestInvalidHeaderFieldName(t *testing.T) {
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
			result := InvalidHeaderFieldName(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNewHeaders(t *testing.T) {
	h := NewHeaders()
	assert.NotNil(t, h.M)
	assert.Equal(t, 0, len(h.M))
}

func TestHeaders_Get(t *testing.T) {
	h := Headers{M: map[string]string{
		"host":         "example.com",
		"content-type": "application/json",
	}}

	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "Existing key lowercase",
			key:      "host",
			expected: "example.com",
		},
		{
			name:     "Existing key mixed case",
			key:      "Host",
			expected: "example.com",
		},
		{
			name:     "Non-existing key",
			key:      "nonexistent",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := h.Get(tc.key)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHeaders_Len(t *testing.T) {
	testCases := []struct {
		name     string
		headers  Headers
		expected int
	}{
		{
			name:     "Empty headers",
			headers:  NewHeaders(),
			expected: 0,
		},
		{
			name: "One header",
			headers: Headers{M: map[string]string{
				"host": "example.com",
			}},
			expected: 1,
		},
		{
			name: "Multiple headers",
			headers: Headers{M: map[string]string{
				"host":         "example.com",
				"content-type": "application/json",
				"user-agent":   "curl",
			}},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.headers.Len()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHeaders_Set(t *testing.T) {
	h := NewHeaders()

	h.Set("Host", "example.com")
	assert.Equal(t, "example.com", h.Get("host"))

	h.Set("host", "newhost.com") // Overwrite
	assert.Equal(t, "newhost.com", h.Get("host"))
}

func TestHeaders_SetTrailer(t *testing.T) {
	h := Headers{M: make(map[string]string)}

	h.SetTrailer("Checksum", "abc123")
	assert.Equal(t, "abc123", h.M["Checksum"]) // Trailers are case-sensitive
}
