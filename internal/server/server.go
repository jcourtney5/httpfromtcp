package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	request "github.com/jcourtney5/httpfromtcp/internal/request"
	"github.com/jcourtney5/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

type Server struct {
	handler  Handler
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &Server{
		handler:  handler,
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
	writer := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteStatusLine(response.StatusCodeBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		writer.WriteHeaders(response.GetDefaultHeaders(len(body)))
		writer.WriteBody(body)
		return
	}
	s.handler(writer, req)
}
