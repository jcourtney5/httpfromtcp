package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jcourtney5/httpfromtcp/internal/request"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to set tcp listener on port %s, %v\n", port, err)
	}
	defer listener.Close()
	log.Printf("Listening on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection %v\n", err)
		}
		defer conn.Close()
		fmt.Println("Accepted new connection")

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("Failed to read request line %v\n", err)
		}
		fmt.Println("Request line:")
		fmt.Println("- Method:", r.RequestLine.Method)
		fmt.Println("- Target:", r.RequestLine.RequestTarget)
		fmt.Println("- Version:", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for key, value := range r.Headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

		fmt.Println("connection has been closed")
	}
}
