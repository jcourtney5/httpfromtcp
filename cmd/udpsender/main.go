package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const address = "localhost:42069"

func main() {
	updAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving UDP address %s, %v\n", address, err)
		os.Exit(1)
	}

	updConn, err := net.DialUDP("udp", nil, updAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to Dial UDP Addr %s, %v\n", address, err)
		os.Exit(1)
	}
	defer updConn.Close()

	fmt.Printf("Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n", address)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read input: %v\n", err)
			os.Exit(1)
		}

		_, err = updConn.Write([]byte(msg))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write input: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Message sent: %s", msg)
	}
}
