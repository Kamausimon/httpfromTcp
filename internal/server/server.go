package server

import (
	"fmt"
	"net"
	"sync/atomic"

	"github.com/Kamausimon/httpFromTcp/internal/request"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

type State int

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
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

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
		return
	}

	fmt.Printf("Method: %s, Target: %s, Version: %s\n",
		req.RequestLine.Method,
		req.RequestLine.RequestTarget,
		req.RequestLine.HttpVersion)

	fmt.Printf("Headers: %v\n", req.Headers)

	if len(req.Body) > 0 {
		fmt.Printf("Body: %s\n", string(req.Body))
	} else {
		fmt.Println("Body: (empty)")
	}

	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 13\r\n" +
		"\r\n" +
		"Hello World!"

	conn.Write([]byte(response))
}
