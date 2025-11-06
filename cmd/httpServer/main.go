package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kamausimon/httpFromTcp/internal/request"
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

func main() {
	server, err := server.Serve(port, DefaultHandler)
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
