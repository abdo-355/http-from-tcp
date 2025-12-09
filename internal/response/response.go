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

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
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
	return err
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

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers.M {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))

		if err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("\r\n"))

	return err
}
