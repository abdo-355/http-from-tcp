package response

import (
	"fmt"
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
