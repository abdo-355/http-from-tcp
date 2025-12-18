// Command httpserver starts an HTTP server that proxies certain requests to httpbin.org
// and serves static HTML responses for other endpoints.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
		handleHttpbinProxy(w, req)
		return
	}

	if target == "/video" && req.RequestLine.Method == "GET" {
		handleVideo(w)
		return
	}

	var status int
	var html string
	switch target {
	case "/yourproblem":
		status = http.StatusBadRequest
		html = `<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>`
	case "/myproblem":
		status = http.StatusInternalServerError
		html = `<html><head><title>500 Internal Server Error</title></head><body><h1>Internal Server Error</h1><p>Okay, you know what? This one is on me.</p></body></html>`
	default:
		status = http.StatusOK
		html = `<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>`
	}

	body := []byte(html)
	h := headers.NewHeaders()
	h.Set("content-type", "text/html")
	h.Set("content-length", strconv.Itoa(len(body)))
	h.Set("connection", "close")
	w.WriteStatusLine("HTTP/1.1", status, http.StatusText(status))
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handleHttpbinProxy(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	proxyRes, err := http.Get("https://httpbin.org" + target)
	if err != nil {
		sendInternalServerError(w, err)
		return
	}
	defer proxyRes.Body.Close()

	w.WriteStatusLine("HTTP/1.1", http.StatusOK, "OK")

	h := headers.NewHeaders()
	h.Set("content-type", "text/html")
	h.Set("transfer-encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders(h)

	buff := make([]byte, 1024)
	cl := 0
	hash := sha256.New()
	for {
		n, err := proxyRes.Body.Read(buff)
		if n > 0 {
			if _, err := w.WriteChunkedBody(buff[:n], hash); err != nil {
				sendInternalServerError(w, fmt.Errorf("error writing chunked body: %w", err))
				return
			}
		}

		cl += n

		if err == io.EOF {
			w.WriteChunkedBodyDone()
			break
		}

		if err != nil {
			sendInternalServerError(w, fmt.Errorf("error reading proxy response: %w", err))
			return
		}
	}

	hashValue := hash.Sum(nil)
	t := headers.NewHeaders()
	t.SetTrailer("X-Content-Sha256", hex.EncodeToString(hashValue))
	t.SetTrailer("X-Content-Length", strconv.Itoa(cl))
	if err := w.WriteTrailers(t); err != nil {
		sendInternalServerError(w, fmt.Errorf("error writing trailers: %w", err))
		return
	}
}

func handleVideo(w *response.Writer) {
	data, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		sendErrorResponse(w, http.StatusNotFound, "video not found", err)
		return
	}

	h := headers.NewHeaders()
	h.Set("content-type", "video/mp4")
	h.Set("content-length", strconv.Itoa(len(data)))
	h.Set("connection", "close")
	w.WriteStatusLine("HTTP/1.1", http.StatusOK, "OK")
	w.WriteHeaders(h)
	w.WriteBody(data)
}

func sendInternalServerError(w *response.Writer, err error) {
	log.Printf("Internal Server Error: %v", err)
	sendErrorResponse(w, http.StatusInternalServerError, "Internal Server Error", err)
}

func sendErrorResponse(w *response.Writer, statusCode int, message string, err error) {
	statusText := http.StatusText(statusCode)
	if statusText == "" {
		statusText = "Unknown"
	}

	errorBody := fmt.Sprintf("%d %s: %s - %s", statusCode, statusText, message, err.Error())

	w.WriteStatusLine("HTTP/1.1", statusCode, statusText)
	h := headers.NewHeaders()
	h.Set("content-type", "text/plain")
	h.Set("content-length", strconv.Itoa(len(errorBody)))
	h.Set("connection", "close")
	w.WriteHeaders(h)
	w.WriteBody([]byte(errorBody))
}
