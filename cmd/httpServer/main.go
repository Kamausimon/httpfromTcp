package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Kamausimon/httpFromTcp/internal/headers"
	"github.com/Kamausimon/httpFromTcp/internal/request"
	"github.com/Kamausimon/httpFromTcp/internal/response"
	"github.com/Kamausimon/httpFromTcp/internal/server"
)

const port = 42069

func DefaultHandler(w io.Writer, req *request.Request) *server.HandlerError {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":

		return &server.HandlerError{
			StatusCode: 400,
			Error: fmt.Errorf(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`),
		}
	case "/myproblem":

		return &server.HandlerError{
			StatusCode: 500,
			Error: fmt.Errorf(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`),
		}
	default:

		_, err := io.WriteString(w, `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		if err != nil {
			return &server.HandlerError{
				StatusCode: 500,
				Error:      fmt.Errorf("failed to write response: %v", err),
			}
		}
		return nil
	}
}

func ProxyHandler(w io.Writer, req *request.Request) *server.HandlerError {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		return DefaultHandler(w, req)
	}

	proxyPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	proxyURL := "https://httpbin.org" + proxyPath

	resp, err := http.Get(proxyURL)
	if err != nil {
		return &server.HandlerError{
			StatusCode: 500,
			Error:      fmt.Errorf("proxy request failed: %v", err),
		}
	}

	defer resp.Body.Close()

	responseWriter := response.NewWriter(w)

	statusCode := response.StatusCode(resp.StatusCode)
	err = responseWriter.WriteStatusLine(statusCode)
	if err != nil {
		return &server.HandlerError{StatusCode: 500, Error: err}
	}

	chunkedHeaders := response.GetChunkedHeaders()
	chunkedHeaders.Override("Trailer", "X-Content-SHA256, X-Content-Length")
	err = responseWriter.WriteHeaders(chunkedHeaders)
	if err != nil {
		return &server.HandlerError{StatusCode: 500, Error: err}
	}

	var fullBody []byte // Track the complete response body
	totalLength := 0

	buffer := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]

			// Add to full body tracking
			fullBody = append(fullBody, chunk...)
			totalLength += n

			fmt.Printf("Read %d bytes\n", n)

			// Write chunk size in hex
			_, writeErr := fmt.Fprintf(w, "%x\r\n", n)
			if writeErr != nil {
				return &server.HandlerError{StatusCode: 500, Error: writeErr}
			}

			// Write chunk data
			_, writeErr = w.Write(chunk)
			if writeErr != nil {
				return &server.HandlerError{StatusCode: 500, Error: writeErr}
			}

			// Write chunk terminator
			_, writeErr = io.WriteString(w, "\r\n")
			if writeErr != nil {
				return &server.HandlerError{StatusCode: 500, Error: writeErr}
			}

			if flusher, ok := w.(interface{ Flush() error }); ok {
				flusher.Flush()
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return &server.HandlerError{StatusCode: 500, Error: err}
		}
	}
	// Write final 0-sized chunk (but not the final \r\n yet)
	_, err = responseWriter.WriteChunkedBodyDone()
	if err != nil {
		return &server.HandlerError{StatusCode: 500, Error: err}
	}

	trailers := headers.NewHeaders()
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Override("X-Content-SHA256", sha256)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = responseWriter.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")

	// Final \r\n to end the HTTP message
	_, err = io.WriteString(w, "\r\n")
	if err != nil {
		return &server.HandlerError{StatusCode: 500, Error: err}
	}

	return nil // Success
}

func main() {
	server, err := server.Serve(port, ProxyHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
