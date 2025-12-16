package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
}

func TestHeaderParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, 0, r.Headers.Len())

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nHost: example.com\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069, example.com", r.Headers.Get("host"))

	// Test: Case Insensitive Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHOST: localhost:42069\r\nUser-Agent: curl/7.81.0\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))

	// Test: Missing End of Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestBodyParse(t *testing.T) {
	// Test: Standard Body
	t.Run("Standard Body", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))
	})

	// Test: Body shorter than reported content length
	t.Run("Body shorter than reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			numBytesPerRead: 3,
		}
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	})

	// Test: Empty Body, 0 reported content length
	t.Run("Empty Body, 0 reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	// Test: Empty Body, no reported content length
	t.Run("Empty Body, no reported content length", func(t *testing.T) {
		reader := &chunkReader{
			data: "GET / HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))
	})

	// Test: No Content-Length but Body Exists
	t.Run("No Content-Length but Body Exists", func(t *testing.T) {
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"unexpected body",
			numBytesPerRead: 3,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		// Assuming no body parsed if no Content-Length
		assert.Equal(t, "", string(r.Body))
	})
}

func TestParseRequestLine(t *testing.T) {
	testCases := []struct {
		name        string
		data        []byte
		expectedRL  *RequestLine
		expectedN   int
		expectError bool
	}{
		{
			name: "Valid request line",
			data: []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			expectedRL: &RequestLine{
				Method:        "GET",
				RequestTarget: "/",
				HTTPVersion:   "1.1",
			},
			expectedN:   16,
			expectError: false,
		},
		{
			name:        "No CRLF",
			data:        []byte("GET / HTTP/1.1Host: example.com"),
			expectedRL:  nil,
			expectedN:   0,
			expectError: false,
		},
		{
			name:        "Invalid request line",
			data:        []byte("INVALID\r\nHost: example.com\r\n\r\n"),
			expectedRL:  nil,
			expectedN:   0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl, n, err := parseRequestLine(tc.data)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expectedRL != nil {
					assert.Equal(t, tc.expectedRL, rl)
				} else {
					assert.Nil(t, rl)
				}
				assert.Equal(t, tc.expectedN, n)
			}
		})
	}
}

func TestRequestLineFromString(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    *RequestLine
		expectError bool
	}{
		{
			name:  "Valid GET",
			input: "GET / HTTP/1.1",
			expected: &RequestLine{
				Method:        "GET",
				RequestTarget: "/",
				HTTPVersion:   "1.1",
			},
			expectError: false,
		},
		{
			name:        "Too few parts",
			input:       "GET /",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid method lowercase",
			input:       "get / HTTP/1.1",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid HTTP version",
			input:       "GET / HTTP/2.0",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Malformed version",
			input:       "GET / INVALID",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rl, err := requestLineFromString(tc.input)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, rl)
			}
		})
	}
}

func TestRequest_Parse(t *testing.T) {
	t.Run("Parse done state", func(t *testing.T) {
		r := &Request{state: Done}
		n, err := r.parse([]byte("some data"))
		assert.Equal(t, 0, n)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot parse data in done state")
	})

	t.Run("Parse initialized with incomplete request line", func(t *testing.T) {
		r := &Request{state: Initialized}
		n, err := r.parse([]byte("GET"))
		assert.Equal(t, 0, n)
		assert.NoError(t, err)
		assert.Equal(t, Initialized, r.state)
	})
}
