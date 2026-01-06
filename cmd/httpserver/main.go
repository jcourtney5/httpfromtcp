package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jcourtney5/httpfromtcp/internal/request"
	"github.com/jcourtney5/httpfromtcp/internal/response"
	"github.com/jcourtney5/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerTest)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handlerTest(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerHttpBin(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	handler200(w, req)
	return
}

func handler400(w *response.Writer, _ *request.Request) {
	body := "<html>\n  <head>\n    <title>400 Bad Request</title>\n  </head>\n  <body>\n    <h1>Bad Request</h1>\n    <p>Your request honestly kinda sucked.</p>\n  </body>\n</html>"
	bodyBytes := []byte(body)
	w.WriteStatusLine(response.StatusCodeBadRequest)
	headers := response.GetDefaultHeaders(len(bodyBytes))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(bodyBytes)
}

func handler500(w *response.Writer, _ *request.Request) {
	body := "<html>\n  <head>\n    <title>500 Internal Server Error</title>\n  </head>\n  <body>\n    <h1>Internal Server Error</h1>\n    <p>Okay, you know what? This one is on me.</p>\n  </body>\n</html>"
	bodyBytes := []byte(body)
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	headers := response.GetDefaultHeaders(len(bodyBytes))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(bodyBytes)
}

func handler200(w *response.Writer, _ *request.Request) {
	body := "<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>"
	bodyBytes := []byte(body)
	w.WriteStatusLine(response.StatusCodeOK)
	headers := response.GetDefaultHeaders(len(bodyBytes))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(bodyBytes)
}

func handlerHttpBin(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := fmt.Sprintf("https://httpbin.org/%s", target)
	resp, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCodeOK)
	headers := response.GetDefaultHeaders(0)
	headers.Remove("Content-Type")
	headers.Override("Transfer-Encoding", "chunked")
	w.WriteHeaders(headers)

	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			fmt.Printf("Read %d bytes (chunk received)\n", n)
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err != nil {
			// io.EOF indicates the end of the stream
			if err == io.EOF {
				break
			}
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
}
