package response

import (
	"crypto/sha256"
	"net/http"
	"testing"

	"github.com/abdo-355/http-from-tcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	testCases := []struct {
		name         string
		statusCode   int
		statusText   string
		expectedData string
		doPanic      bool
		initialState WriterState
	}{
		{
			name:         "Status OK",
			statusCode:   http.StatusOK,
			statusText:   "OK",
			expectedData: "HTTP/1.1 200 OK\r\n",
			doPanic:      false,
			initialState: WriteStatusLine,
		},
		{
			name:         "Status Bad Request",
			statusCode:   http.StatusBadRequest,
			statusText:   "Bad Request",
			expectedData: "HTTP/1.1 400 Bad Request\r\n",
			doPanic:      false,
			initialState: WriteStatusLine,
		},
		{
			name:         "Status Internal Server Error",
			statusCode:   http.StatusInternalServerError,
			statusText:   "Internal Server Error",
			expectedData: "HTTP/1.1 500 Internal Server Error\r\n",
			doPanic:      false,
			initialState: WriteStatusLine,
		},
		{
			name:         "Custom status code",
			statusCode:   404,
			statusText:   "Not Found",
			expectedData: "HTTP/1.1 404 Not Found\r\n",
			doPanic:      false,
			initialState: WriteStatusLine,
		},
		{
			name:         "Invalid state",
			statusCode:   http.StatusOK,
			statusText:   "OK",
			expectedData: "",
			doPanic:      true,
			initialState: WriteHeaders,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := New()
			w.State = tc.initialState

			if tc.doPanic {
				assert.Panics(t, func() {
					w.WriteStatusLine("HTTP/1.1", tc.statusCode, tc.statusText)
				})
			} else {
				w.WriteStatusLine("HTTP/1.1", tc.statusCode, tc.statusText)
				assert.Equal(t, tc.expectedData, w.buffer.String())
				assert.Equal(t, WriteHeaders, w.State)
			}
		})
	}
}

func TestGetDefaultHeaders(t *testing.T) {
	testCases := []struct {
		name       string
		contentLen int
		expected   headers.Headers
	}{
		{
			name:       "Zero content length",
			contentLen: 0,
			expected: headers.Headers{
				M: map[string]string{
					"content-length": "0",
					"connection":     "close",
					"content-type":   "text/plain",
				},
			},
		},
		{
			name:       "Positive content length",
			contentLen: 13,
			expected: headers.Headers{
				M: map[string]string{
					"content-length": "13",
					"connection":     "close",
					"content-type":   "text/plain",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetDefaultHeaders(tc.contentLen)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestWriteHeaders(t *testing.T) {
	testCases := []struct {
		name         string
		headers      headers.Headers
		expectLines  []string
		doPanic      bool
		initialState WriterState
	}{
		{
			name: "Valid headers",
			headers: headers.Headers{
				M: map[string]string{
					"content-type": "application/json",
					"host":         "example.com",
				},
			},
			expectLines:  []string{"content-type: application/json\r\n", "host: example.com\r\n", "\r\n"},
			doPanic:      false,
			initialState: WriteHeaders,
		},
		{
			name: "Empty headers",
			headers: headers.Headers{
				M: map[string]string{},
			},
			expectLines:  []string{"\r\n"},
			doPanic:      false,
			initialState: WriteHeaders,
		},
		{
			name: "Invalid state",
			headers: headers.Headers{
				M: map[string]string{"test": "value"},
			},
			expectLines:  nil,
			doPanic:      true,
			initialState: WriteStatusLine,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := New()
			w.State = tc.initialState

			if tc.doPanic {
				assert.Panics(t, func() {
					w.WriteHeaders(tc.headers)
				})
			} else {
				w.WriteHeaders(tc.headers)
				// Since map order is random, check that all expected lines are present
				for _, expected := range tc.expectLines {
					assert.Contains(t, w.buffer.String(), expected)
				}
				assert.Equal(t, WriteBody, w.State)
			}
		})
	}
}

func TestWriteBody(t *testing.T) {
	testCases := []struct {
		name         string
		body         []byte
		expectedData string
		doPanic      bool
		initialState WriterState
	}{
		{
			name:         "Valid body",
			body:         []byte("hello world"),
			expectedData: "hello world",
			doPanic:      false,
			initialState: WriteBody,
		},
		{
			name:         "Empty body",
			body:         []byte(""),
			expectedData: "",
			doPanic:      false,
			initialState: WriteBody,
		},
		{
			name:         "Invalid state",
			body:         []byte("test"),
			expectedData: "",
			doPanic:      true,
			initialState: WriteStatusLine,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := New()
			w.State = tc.initialState
			if tc.doPanic {
				assert.Panics(t, func() {
					w.WriteBody(tc.body)
				})
			} else {
				w.WriteBody(tc.body)
				assert.Equal(t, tc.expectedData, w.buffer.String())
			}
		})
	}
}

func TestWriteChunkedBody(t *testing.T) {
	w := New()
	w.State = WriteBody // Assume it's in body writing state for chunked
	h := sha256.New()

	data := []byte("hello world")
	n, err := w.WriteChunkedBody(data, h)
	require.NoError(t, err)
	assert.Equal(t, 16, n) // "b\r\nhello world\r\n" is 16 bytes
	assert.Equal(t, "b\r\nhello world\r\n", w.buffer.String())

	// Check hash
	expectedHash := sha256.Sum256(data)
	actualHash := h.Sum(nil)
	assert.Equal(t, expectedHash[:], actualHash)
}

func TestWriteChunkedBodyDone(t *testing.T) {
	w := New()
	w.State = WriteBody

	n, err := w.WriteChunkedBodyDone()
	require.NoError(t, err)
	assert.Equal(t, 3, n) // "0\r\n"
	assert.Equal(t, "0\r\n", w.buffer.String())
	assert.Equal(t, WriteTrailers, w.State)
}

func TestWriteTrailers(t *testing.T) {
	testCases := []struct {
		name         string
		trailers     headers.Headers
		expectedData string
		expectError  bool
		initialState WriterState
	}{
		{
			name: "Valid trailers",
			trailers: headers.Headers{
				M: map[string]string{
					"checksum": "abc123",
				},
			},
			expectedData: "checksum: abc123\r\n\r\n",
			expectError:  false,
			initialState: WriteTrailers,
		},
		{
			name: "Empty trailers",
			trailers: headers.Headers{
				M: map[string]string{},
			},
			expectedData: "\r\n",
			expectError:  false,
			initialState: WriteTrailers,
		},
		{
			name: "Invalid state",
			trailers: headers.Headers{
				M: map[string]string{"test": "value"},
			},
			expectedData: "",
			expectError:  true,
			initialState: WriteBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := New()
			w.State = tc.initialState

			err := w.WriteTrailers(tc.trailers)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedData, w.buffer.String())
			}
		})
	}
}
