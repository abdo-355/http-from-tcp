// Package server implements an HTTP server that handles TCP connections and parses requests.
package server

import (
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/abdo-355/http-from-tcp/internal/request"
	"github.com/abdo-355/http-from-tcp/internal/response"
)

type Server struct {
	handler  Handler
	Listener net.Listener
	state    atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

func (he *HandlerError) Write(w io.Writer) {
	writer := response.Writer{
		Conn: w,
	}
	writer.WriteStatusLine(he.Status)
	msg := []byte(he.Message)

	hdrs := response.GetDefaultHeaders(len(msg))

	writer.WriteHeaders(hdrs)
	writer.WriteBody(msg)
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	srv := Server{
		Listener: listener,
		handler:  handler,
	}

	srv.state.Store(true)

	go srv.listen()

	return &srv, nil
}

func (s *Server) Close() error {
	s.state.Store(false)
	err := s.Listener.Close()
	return err
}

func (s *Server) listen() {
	for s.state.Load() {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Fatal("error accepting a connection on the listener:", err)
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			Status:  response.StatusBadRequest,
			Message: err.Error(),
		}

		hErr.Write(conn)
		return
	}

	w := response.Writer{
		Conn: conn,
	}

	s.handler(&w, req)
}
