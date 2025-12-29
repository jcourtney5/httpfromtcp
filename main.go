package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const fileName = "messages.txt"

func main() {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open file %s, %v\n", fileName, err)
	}

	fmt.Printf("Reading data from %s\n", fileName)
	fmt.Println("=====================================")

	linesCh := getLinesChannel(f)
	for line := range linesCh {
		fmt.Println("read:", line)
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
