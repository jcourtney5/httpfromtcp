package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/jcourtney5/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialized    = iota
	requestStateParsingHeaders // 0
	requestStateDone           // 1
)

const bufferSize = 8
const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := &Request{
		Headers: headers.NewHeaders(),
		state:   requestStateInitialized,
	}

	for request.state != requestStateDone {
		// check if we need to grow the buffer
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		bytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", request.state, bytesRead)
				}
				break
			}
			return nil, err
		}
		readToIndex += bytesRead

		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		if bytesParsed > 0 {
			// remove parsed data from buf copying data left and decrementing readToIndex
			copy(buf, buf[bytesParsed:])
			readToIndex -= bytesParsed
		}
	}

	return request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}

	requestLineText := string(data[:idx])

	parts := strings.Split(requestLineText, " ")
	if len(parts) != 3 {
		return nil, 0, fmt.Errorf("poorly formatted request-line: %s", requestLineText)
	}

	// get http method and validate that the method is all uppercase characters
	method := parts[0]
	for _, c := range method {
		if !unicode.IsUpper(c) {
			return nil, 0, fmt.Errorf("invalid method: %s", method)
		}
	}

	// get http target
	target := parts[1]

	// get http version and make sure it is "1.1"
	httpVersionParts := strings.Split(parts[2], "/")
	if len(httpVersionParts) != 2 {
		return nil, 0, fmt.Errorf("malformed start-line: %s", requestLineText)
	}
	httpPart := httpVersionParts[0]
	if httpPart != "HTTP" {
		return nil, 0, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := httpVersionParts[1]
	if version != "1.1" {
		return nil, 0, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: target,
		HttpVersion:   version,
	}, idx + 2, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break // needs more data
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil // needs more data
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateDone
		}
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("Can't parse in done state")
	default:
		return 0, fmt.Errorf("Unknown state")
	}
}
