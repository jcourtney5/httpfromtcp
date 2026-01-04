package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	request "github.com/jcourtney5/httpfromtcp/internal/request"
	"github.com/jcourtney5/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	messageBytes := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	response.WriteHeaders(w, headers)
	w.Write(messageBytes)
}

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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	// write status line
	err = response.WriteStatusLine(conn, response.StatusCodeOK)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCodeInternalServerError,
			Message:    fmt.Sprintf("Failed to write status line: %v\n", err),
		}
		hErr.Write(conn)
		return
	}

	// write headers
	headers := response.GetDefaultHeaders(buf.Len())
	err = response.WriteHeaders(conn, headers)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCodeInternalServerError,
			Message:    fmt.Sprintf("Failed to write headers %v\n", err),
		}
		hErr.Write(conn)
		return
	}

	// write content
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusCodeInternalServerError,
			Message:    fmt.Sprintf("Failed to write content %v\n", err),
		}
		hErr.Write(conn)
		return
	}
}
