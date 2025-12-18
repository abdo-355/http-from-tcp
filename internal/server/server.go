// Package server implements an HTTP server that handles TCP connections and parses requests.
package server

import (
	"log/slog"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/abdo-355/http-from-tcp/internal/httpwriter"
	"github.com/abdo-355/http-from-tcp/internal/request"
	"github.com/abdo-355/http-from-tcp/internal/response"
)

type Server struct {
	handler  Handler
	Listener net.Listener
	state    atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

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
			// If the server is closing, the error is expected.
			if !s.state.Load() {
				return
			}
			slog.Error("error accepting a connection on the listener:", "err", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		slog.Warn("error parsing request", "err", err, "remote_addr", conn.RemoteAddr())
		httpwriter.SendError(conn, 400)
		return
	}

	res := response.New()
	s.handler(res, req)

	if err := httpwriter.Write(conn, res); err != nil {
		slog.Error("error writing response", "err", err)
	}
}
