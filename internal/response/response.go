// Package response provides functionality for building HTTP responses.
package response

import (
	"bytes"
	"fmt"
	"hash"
	"strconv"

	"github.com/abdo-355/http-from-tcp/internal/headers"
)

type Writer struct {
	buffer *bytes.Buffer
	State  WriterState
}

func New() *Writer {
	return &Writer{buffer: new(bytes.Buffer)}
}

func (w *Writer) Bytes() []byte {
	return w.buffer.Bytes()
}

type WriterState int

const (
	WriteStatusLine WriterState = iota
	WriteHeaders
	WriteBody
	WriteTrailers
)

func (w *Writer) Write(data []byte) (n int, err error) {
	return w.buffer.Write(data)
}

func (w *Writer) WriteStatusLine(proto string, statusCode int, statusText string) {
	if w.State != WriteStatusLine {
		panic("invalid operations order. make sure this is run first")
	}
	fmt.Fprintf(w.buffer, "%s %d %s\r\n", proto, statusCode, statusText)
	w.State = WriteHeaders
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		M: map[string]string{
			"content-length": strconv.Itoa(contentLen),
			"connection":     "close",
			"content-type":   "text/plain",
		},
	}
}

func (w *Writer) WriteHeaders(headers headers.Headers) {
	if w.State != WriteHeaders {
		panic("invalid operations order. make sure this runs after writing the status line and before writing the body")
	}
	for k, v := range headers.M {
		fmt.Fprintf(w.buffer, "%s: %s\r\n", k, v)
	}

	w.buffer.WriteString("\r\n")
	w.State = WriteBody
}

func (w *Writer) WriteBody(b []byte) {
	if w.State != WriteBody {
		panic("invalid operations order. make sure this runs last")
	}

	w.buffer.Write(b)
}

func (w *Writer) WriteChunkedBody(p []byte, h hash.Hash) (int, error) {
	chunkSizeHex := fmt.Sprintf("%x", len(p))
	var bytesWritten int
	n, err := w.Write([]byte(chunkSizeHex + "\r\n"))
	if err != nil {
		return 0, err
	}
	bytesWritten += n

	n, err = w.Write(append(p, "\r\n"...))
	if err != nil {
		return 0, err
	}
	bytesWritten += n

	_, err = h.Write(p)

	return bytesWritten, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	w.State = WriteTrailers
	return w.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {

	if w.State != WriteTrailers {
		return fmt.Errorf("invalid operations order. make sure this runs last")
	}
	for k, v := range h.M {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))

		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	return nil
}
