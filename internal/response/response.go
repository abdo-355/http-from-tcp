package response

import (
	"fmt"
	"hash"
	"io"
	"strconv"

	"github.com/abdo-355/http-from-tcp/internal/headers"
)

type StatusCode int

const (
	StatusOk                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

type Writer struct {
	Conn  io.Writer
	State WriterState
}

type WriterState int

const (
	WriteStatusLine WriterState = iota
	WriteHeaders
	WriteBody
	WriteTrailers
)

func (w *Writer) Write(data []byte) (n int, err error) {
	n, err = w.Conn.Write(data)
	return
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != WriteStatusLine {
		return fmt.Errorf("invalid operations order. make sure this is run first")
	}
	msg := ""
	switch statusCode {
	case StatusOk:
		msg = "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		msg = "HTTP/1.1 400 Bad Request\r\n"
	case StatusInternalServerError:
		msg = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		msg = fmt.Sprintf("HTTP/1.1 %d\r\n", statusCode)

	}

	_, err := w.Write([]byte(msg))
	if err != nil {
		return err
	}
	w.State = WriteHeaders
	return nil
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

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != WriteHeaders {
		return fmt.Errorf("invalid operations order. make sure this runs after writing the status line and before writing the body")
	}
	for k, v := range headers.M {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))

		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	w.State = WriteBody
	return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
	if w.State != WriteBody {
		return 0, fmt.Errorf("invalid operations order. make sure this runs last")
	}

	n, err := w.Write(b)

	if err != nil {
		return 0, err
	}

	return n, nil
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
