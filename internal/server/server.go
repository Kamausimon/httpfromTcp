package server

import (
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

type Handler func(w *response.Writer, req *request.Request)

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
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	s.Handler(w, req)

}
