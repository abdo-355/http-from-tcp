package server

import (
	"bytes"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	Status  response.StatusCode
	Message string
}

func (he *HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.Status)
	msg := []byte(he.Message)

	hdrs := response.GetDefaultHeaders(len(msg))

	response.WriteHeaders(w, hdrs)
	w.Write(msg)
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

	buf := bytes.NewBuffer([]byte{})
	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOk)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
}
