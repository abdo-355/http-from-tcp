# HTTP-FROM-TCP

`http-from-tcp` is a project that provides a from-scratch implementation of an HTTP/1.1 server in Go. It is designed to be an educational tool, showing how the HTTP protocol functions directly on top of TCP sockets.

The project includes a working web server, a diagnostic listener tool, and a set of internal packages that together form a lightweight and reusable HTTP library.

## Features

- **HTTP/1.1 Server:** A functional server built from the ground up, capable of handling common HTTP requests.
- **Request Parsing:** A streaming parser that translates raw TCP data into a structured HTTP request object.
- **Response Writing:** A stateful writer for constructing and sending valid HTTP/1.1 responses to a client.
- **Chunked Transfer Encoding:** Supports sending and receiving data in chunks, which is essential for handling large or streaming bodies.
- **Request Proxying:** The server can forward requests to `httpbin.org` and stream the responses back to the client using chunked encoding.
- **Static File Serving:** The server can serve local files (e.g., a video) over HTTP.
- **Unit Tests:** The core logic is validated by a comprehensive suite of unit tests.

## Getting Started

### Prerequisites

- Go (version 1.18 or later)

### Building the Binaries

To compile the executable applications, run the `go build` command from the project root:

```bash
# Build the main web server
go build ./cmd/httpserver

# Build the TCP listener diagnostic tool
go build ./cmd/tcplistener
```

### Running the Applications

You can run the applications directly using `go run`. Note that the server listens on port **8080**.

```bash
# Run the main HTTP server (listens on port 8080)
go run ./cmd/httpserver/main.go

# Run the TCP listener tool (listens on port 8080)
go run ./cmd/tcplistener/main.go
```

**Note:** `httpserver` and `tcplistener` may use the same port, so they cannot be run at the same time.

### Running Tests

To run the full suite of unit tests, execute the following command from the root directory:

```bash
go test ./...
```

## Usage

### HTTP Server

The `httpserver` application is a web server that showcases the project's features. It listens on `http://localhost:8080` and provides several endpoints:

- `GET /`: Responds with a default 200 OK HTML page.
- `GET /httpbin/*`: Forwards the request to `https://httpbin.org` and streams the response back. For example, `/httpbin/get` will be proxied.
- `GET /video`: Serves a local video file (`./assets/vim.mp4`).
- `GET /yourproblem`: Responds with a sample 400 Bad Request error page.
- `GET /myproblem`: Responds with a sample 500 Internal Server Error page.

You can interact with the server using a tool like `curl`:

```bash
# Get the default success page
curl http://localhost:8080/

# Proxy a request to httpbin's get endpoint
curl http://localhost:8080/httpbin/get

# Download the sample video
curl http://localhost:8080/video -o vim.mp4
```

### TCP Listener

The `tcplistener` is a diagnostic tool for inspecting raw HTTP requests. It listens for a single TCP connection, parses the incoming request, prints its structure to the console, and then exits.

1. Run the listener:

    ```bash
    go run ./cmd/tcplistener/main.go
    ```

2. In a separate terminal, send an HTTP request to it:

    ```bash
    curl --data "hello" http://localhost:8080/some/path
    ```

3. The listener's terminal will display the parsed request details.

## Project Structure

The codebase is organized into two primary directories: `cmd` and `internal`.

```
.
├── cmd/
│   ├── httpserver/     # Main HTTP server application
│   ├── tcplistener/    # TCP listener diagnostic tool
│   └── udpserver/      # A simple UDP client utility
└── internal/
    ├── headers/        # HTTP header parsing logic
    ├── request/        # HTTP request parsing logic
    ├── response/       # HTTP response writing logic
    └── server/         # Core TCP server implementation
```

- **`cmd/`**: Contains the main entry points for the executable applications.
- **`internal/`**: Contains the core logic, structured as a set of internal packages.
  - `server`: A reusable TCP server that handles connection listening and management.
  - `request`: Logic for parsing an incoming byte stream into a structured HTTP request.
  - `response`: Logic for creating and sending a structured HTTP response back to a client.
  - `headers`: A helper package for parsing and handling HTTP headers.
