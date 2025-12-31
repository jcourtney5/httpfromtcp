package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	str := string(bytes)

	lines := strings.Split(str, "\r\n")
	if len(lines) < 2 {
		return nil, errors.New("invalid request")
	}

	requestLine, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", line)
	}

	// get http method and validate that the method is all uppercase characters
	method := parts[0]
	for _, c := range method {
		if !unicode.IsUpper(c) {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	// get http target
	target := parts[1]

	// get http version and make sure it is "1.1"
	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", line)
	}
	httpPart := httpVersionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := httpVersionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, nil
}
