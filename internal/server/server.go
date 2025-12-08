package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"
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
	msg := []byte("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 12\r\n" +
		"\r\n" +
		"Hello World!")

	n, err := conn.Write(msg)
	if err != nil {
		log.Printf("Write failed: %v", err)
		return
	}
	if n != len(msg) {
		log.Printf("Partial write: %d/%d bytes", n, len(msg))
	}
}
