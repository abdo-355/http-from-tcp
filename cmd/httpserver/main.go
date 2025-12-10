package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/abdo-355/http-from-tcp/internal/headers"
	"github.com/abdo-355/http-from-tcp/internal/request"
	"github.com/abdo-355/http-from-tcp/internal/response"
	"github.com/abdo-355/http-from-tcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() {
		err = server.Close()
		if err != nil {
			log.Fatalf("error closing the server: %v", err)
		}
	}()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	var status response.StatusCode
	var html string
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		status = response.StatusBadRequest
		html = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`
	case "/myproblem":
		status = response.StatusInternalServerError
		html = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`
	default:
		status = response.StatusOk
		html = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`
	}
	body := []byte(html)
	h := headers.NewHeaders()
	h.Set("content-type", "text/html")
	h.Set("content-length", strconv.Itoa(len(body)))
	h.Set("connection", "close")
	w.WriteStatusLine(status)
	w.WriteHeaders(h)
	w.WriteBody(body)
}
