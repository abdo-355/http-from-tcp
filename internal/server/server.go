package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/abdo-355/http-from-tcp/internal/response"
)

type Server struct {
	Listener net.Listener
	state    atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("an error occurend when opening tcp connection:", err)
	}

	srv := Server{
		Listener: listener,
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

	h := response.GetDefaultHeaders(0)

	response.WriteStatusLine(conn, response.StatusOk)
	response.WriteHeaders(conn, h)
}
