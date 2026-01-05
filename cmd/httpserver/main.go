package main

import (
	"log"
	"os"
	"os/signal"
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
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	} else if req.RequestLine.RequestTarget == "/myproblem" {
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

func handler200(w *response.Writer, req *request.Request) {
	body := "<html>\n  <head>\n    <title>200 OK</title>\n  </head>\n  <body>\n    <h1>Success!</h1>\n    <p>Your request was an absolute banger.</p>\n  </body>\n</html>"
	bodyBytes := []byte(body)
	w.WriteStatusLine(response.StatusCodeOK)
	headers := response.GetDefaultHeaders(len(bodyBytes))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(bodyBytes)
}
