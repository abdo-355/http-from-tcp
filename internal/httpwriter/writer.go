package httpwriter

import (
	"fmt"
	"io"
	"net/http"

	"github.com/abdo-355/http-from-tcp/internal/response"
)

func SendError(w io.Writer, code int) error {
	res := response.New()
	statusText := http.StatusText(code)
	if statusText == "" {
		statusText = "Unknown"
	}
	res.WriteStatusLine("HTTP/1.1", code, statusText)

	// Create a minimal body for the error
	errorBody := fmt.Sprintf("%d %s", code, statusText)
	headers := response.GetDefaultHeaders(len(errorBody))
	res.WriteHeaders(headers)
	res.WriteBody([]byte(errorBody))

	_, err := w.Write(res.Bytes())
	return err
}

func Write(w io.Writer, res *response.Writer) error {
	_, err := w.Write(res.Bytes())
	return err
}
