package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const fileName = "messages.txt"

func main() {

	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file %s, %v\n", fileName, err)
	}
	defer f.Close()

	fmt.Printf("Reading data from %s\n", fileName)
	fmt.Println("=====================================")

	buffer := make([]byte, 8)

	for {
		n, err := f.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("end of file")
			}
			fmt.Printf("Error occured while reading file %s\n", err.Error())
			break
		}

		fmt.Printf("read: %s\n", string(buffer[:n]))
	}

}
