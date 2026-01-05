package response

import (
	"fmt"
	"io"

	"github.com/jcourtney5/httpfromtcp/internal/headers"
)

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

type Writer struct {
	writerState writerState
	writer      io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writerState: writerStateStatusLine,
		writer:      w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	// make sure we are in status line state
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("writer not in status line state %d", w.writerState)
	}
	_, err := w.writer.Write(getStatusLine(statusCode))
	w.writerState = writerStateHeaders // next state
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	// make sure we are in headers state
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("writer not in headers state %d", w.writerState)
	}
	for key, value := range headers {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	w.writerState = writerStateBody // next state
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	// make sure we are in body state
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("writer not in body state %d", w.writerState)
	}
	return w.writer.Write(p)
}
