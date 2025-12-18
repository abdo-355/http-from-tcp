package server

import (
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/abdo-355/http-from-tcp/internal/request"
	"github.com/abdo-355/http-from-tcp/internal/response"
	"github.com/stretchr/testify/assert"
)

type MockConn struct {
	*strings.Reader
	*strings.Builder
}

func (mc *MockConn) Write(p []byte) (n int, err error) {
	return mc.Builder.Write(p)
}

func (mc *MockConn) Close() error                       { return nil }
func (mc *MockConn) RemoteAddr() net.Addr               { return nil }
func (mc *MockConn) LocalAddr() net.Addr                { return nil }
func (mc *MockConn) SetDeadline(t time.Time) error      { return nil }
func (mc *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (mc *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestHandle_Success(t *testing.T) {
	reqString := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	conn := &MockConn{Reader: strings.NewReader(reqString), Builder: new(strings.Builder)}

	handler := func(w *response.Writer, req *request.Request) {
		w.WriteStatusLine("HTTP/1.1", http.StatusOK, "OK")
		body := []byte("Success")
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
	}

	srv := &Server{handler: handler}

	srv.handle(conn)

	assert.Contains(t, conn.Builder.String(), "HTTP/1.1 200 OK")
	assert.Contains(t, conn.Builder.String(), "Success")
}

func TestHandle_BadRequest(t *testing.T) {
	reqString := "this is not a valid http request"
	conn := &MockConn{Reader: strings.NewReader(reqString), Builder: new(strings.Builder)}

	srv := &Server{}

	srv.handle(conn)

	assert.Contains(t, conn.Builder.String(), "HTTP/1.1 400 Bad Request")
}

func TestHandle_HandlerWritesError(t *testing.T) {
	reqString := "GET /error HTTP/1.1\r\nHost: example.com\r\n\r\n"
	conn := &MockConn{Reader: strings.NewReader(reqString), Builder: new(strings.Builder)}

	handler := func(w *response.Writer, req *request.Request) {
		w.WriteStatusLine("HTTP/1.1", http.StatusInternalServerError, "Internal Server Error")
		body := []byte("Handler induced error")
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
	}

	srv := &Server{handler: handler}

	srv.handle(conn)

	assert.Contains(t, conn.Builder.String(), "HTTP/1.1 500 Internal Server Error")
	assert.Contains(t, conn.Builder.String(), "Handler induced error")
}

