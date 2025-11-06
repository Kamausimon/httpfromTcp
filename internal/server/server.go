package server

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync/atomic"

	"github.com/Kamausimon/httpFromTcp/internal/request"
	"github.com/Kamausimon/httpFromTcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	Handler  Handler
}

type HandlerError struct {
	StatusCode int
	Error      error
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

type State int

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		Handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {

	s.closed.Store(true)

	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {

			if s.closed.Load() {
				return
			}

			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go s.handle(conn)
	}
}

func handleError(w io.Writer, handleErr *HandlerError) {
	if handleErr == nil {
		return
	}

	statusCode := response.StatusCode(handleErr.StatusCode)
	errBody := handleErr.Error.Error()

	err := response.NewWriter(w).WriteStatusLine(statusCode)
	if err != nil {
		fmt.Printf("error writing error status %s", err)
		return
	}

	errorHeaders := response.GetDefaultHeaders(len(errBody))

	err = response.NewWriter(w).WriteHeaders(errorHeaders)
	if err != nil {
		fmt.Printf("error writing error headers: %s", err)
	}

	_, err = io.WriteString(w, errBody)
	if err != nil {
		fmt.Printf("Error writing error body %v", err)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		parseError := &HandlerError{
			StatusCode: 400,
			Error:      fmt.Errorf("bad request %v", err),
		}
		handleError(conn, parseError)
		return
	}

	var buffer bytes.Buffer
	handleErr := s.Handler(&buffer, req)
	if handleErr != nil {
		handleError(conn, handleErr)
		return
	}
	responseBody := buffer.String()
	defaultHeaders := response.GetDefaultHeaders(len(responseBody))
	// Print request for debugging
	fmt.Printf("Method: %s, Target: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget)

	// Write status line
	err = response.NewWriter(conn).WriteStatusLine(response.StatusOK)
	if err != nil {
		fmt.Printf("Error writing status line: %v\n", err)
		return
	}

	// Write headers
	err = response.NewWriter(conn).WriteHeaders(defaultHeaders)
	if err != nil {
		fmt.Printf("Error writing headers: %v\n", err)
		return
	}

	// Write body
	_, err = io.WriteString(conn, responseBody)
	if err != nil {
		fmt.Printf("Error writing body: %v\n", err)
		return
	}
}
