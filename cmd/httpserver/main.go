// Command httpserver starts an HTTP server that proxies certain requests to httpbin.org
// and serves static HTML responses for other endpoints.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/abdo-355/http-from-tcp/internal/headers"
	"github.com/abdo-355/http-from-tcp/internal/request"
	"github.com/abdo-355/http-from-tcp/internal/response"
	"github.com/abdo-355/http-from-tcp/internal/server"
)

const port = 8080

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
	target := req.RequestLine.RequestTarget
	if strings.HasPrefix(target, "/httpbin/") {
		target = strings.TrimPrefix(target, "/httpbin")

		res, err := http.Get("https://httpbin.org" + target)
		if err != nil {
			log.Fatal("error requesting the target:", err.Error())
		}
		defer res.Body.Close()

		w.WriteStatusLine(response.StatusOk)

		h := headers.NewHeaders()
		h.Set("content-type", "text/html")
		h.Set("transfer-encoding", "chunked")
		h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
		w.WriteHeaders(h)

		buff := make([]byte, 1024)
		cl := 0
		hash := sha256.New()
		for {
			n, err := res.Body.Read(buff)
			if n > 0 {
				_, err = w.WriteChunkedBody(buff[:n], hash)
				if err != nil {
					log.Fatal("error writing chunked body:", err)
				}
			}

			cl += n

			if err == io.EOF {
				w.WriteChunkedBodyDone()
				break
			}

			if err != nil {
				log.Fatal(err.Error())
			}
		}

		hashValue := hash.Sum(nil)

		t := headers.NewHeaders()

		t.SetTrailer("X-Content-Sha256", hex.EncodeToString(hashValue))
		t.SetTrailer("X-Content-Length", strconv.Itoa(cl))
		w.WriteTrailers(t)

		return
	}

	if target == "/video" && req.RequestLine.Method == "GET" {
		data, err := os.ReadFile("./assets/vim.mp4")
		if err != nil {
			log.Fatal("error reading the file:", err)
		}

		h := headers.NewHeaders()
		h.Set("content-type", "video/mp4")
		h.Set("content-length", strconv.Itoa(len(data)))
		h.Set("connection", "close")
		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(h)
		w.WriteBody(data)

		return
	}

	var status response.StatusCode
	var html string
	switch target {
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
