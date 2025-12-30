package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		linesCh := getLinesChannel(conn)
		for line := range linesCh {
			fmt.Println(line)
		}

		fmt.Println("connection has been closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer f.Close()
		defer close(lines)

		buffer := make([]byte, 8)
		var lineBuilder strings.Builder

		for {
			n, err := f.Read(buffer)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("Error occured while reading file %s\n", err.Error())
				break
			}

			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			if len(parts) == 1 {
				lineBuilder.WriteString(parts[0])
			} else if len(parts) > 1 {
				for index, part := range parts {
					lineBuilder.WriteString(part)
					if index != len(parts)-1 {
						lines <- lineBuilder.String()
						lineBuilder.Reset()
					}
				}
			}
		}

		if lineBuilder.Len() > 0 {
			lines <- lineBuilder.String()
			lineBuilder.Reset()
		}
	}()

	return lines
}
