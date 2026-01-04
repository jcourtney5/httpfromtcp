package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/jcourtney5/httpfromtcp/internal/headers"
)

type StatusCode int

const (
	StatusCodeOK                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reasonPhrase := ""
	
	switch statusCode {
	case StatusCodeOK:
		reasonPhrase = "OK"
	case StatusCodeBadRequest:
		reasonPhrase = "Bad Request"
	case StatusCodeInternalServerError:
		reasonPhrase = "Internal Server Error"
	}

	_, err := w.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reasonPhrase)))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	defaultHeaders := headers.NewHeaders()
	defaultHeaders.Set("Content-Length", strconv.Itoa(contentLen))
	defaultHeaders.Set("Connection", "close")
	defaultHeaders.Set("Content-Type", "text/plain")
	return defaultHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		_, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}
