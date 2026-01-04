package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jcourtney5/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
	}

	// start listening in a go routing
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				// ignore errors if closed
				return
			}
			log.Printf("Failed to accept connection %v\n", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	content := ""
	statusCode := response.StatusCodeOK

	// write status line
	err := response.WriteStatusLine(conn, statusCode)
	if err != nil {
		log.Printf("Failed to write status line: %v\n", err)
	}

	// write headers
	headers := response.GetDefaultHeaders(len(content))
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Printf("Failed to write headers %v\n", err)
	}

	// write content
	_, err = conn.Write([]byte(content))
	if err != nil {
		log.Printf("Failed to write content %v\n", err)
	}
}
