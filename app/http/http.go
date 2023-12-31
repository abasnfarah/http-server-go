package http

import (
	"fmt"
	"net"
	"os"
	"strings"

	"go.uber.org/zap"
)

type request struct {
	HTTPMethod  string
	Path        string
	HTTPVersion string
	HTTPHeaders []string
}

type response struct {
	Status      string
	HTTPHeaders []string
	Body        string
}

type HTTP struct {
	logger     *zap.Logger
	Listener   net.Listener
	Connection net.Conn
	Request    request
	Response   response
}

func NewHTTPServer() *HTTP {
	logger, _ := zap.NewProduction()
	logger.Info("Starting HTTP Server")
	return &HTTP{logger: logger}
}

func (h *HTTP) accept() {
	c, err := h.Listener.Accept()
	if err != nil {
		os.Exit(1)
	}

	h.Connection = c
}

func (h *HTTP) serializeRequest(request []byte) {
	requestLine := strings.Split(string(request), "\r\n")
	startLineSections := strings.Split(requestLine[0], " ")
	h.Request.HTTPMethod = startLineSections[0]
	h.Request.Path = startLineSections[1]
	h.Request.HTTPVersion = startLineSections[2]

	for _, header := range requestLine[1:] {
		if header == "" {
			break
		}
		h.Request.HTTPHeaders = append(h.Request.HTTPHeaders, header)
	}

	h.logger.Info("Deserialized Request: ", zap.Any("request", h.Request))
}

func (h *HTTP) serializeResponse() {
	successful := []byte("HTTP/1.1 200 OK")
	unSuccessful := []byte("HTTP/1.1 404 Not Found")
	var responseStartLine []byte
	body := []byte("")

	switch {
	case h.Request.Path == "/":
		responseStartLine = successful

	case strings.HasPrefix(h.Request.Path, "/echo"):
		responseStartLine = successful
		body = []byte(h.Request.Path[len("/echo"):])
		if len(body) > 1 && body[0] == '/' {
			body = body[1:]
		} else {
			body = []byte("")
		}

	default:
		responseStartLine = unSuccessful

	}

	h.Response.Status = string(responseStartLine)
	h.Response.Body = string(body)
	h.Response.HTTPHeaders = append(h.Response.HTTPHeaders, "Content-Type: text/plain")

	var contentLengthString string
	if len(body) > 0 {
		contentLengthString = fmt.Sprint(len(body))
	} else {
		contentLengthString = "0"
	}

	h.Response.HTTPHeaders = append(h.Response.HTTPHeaders, "Content-Length: "+contentLengthString)

	h.logger.Info("Serialized Response: ", zap.Any("response", h.Response))
}

func (h *HTTP) deserializeResponse() []byte {

	response := h.Response.Status + "\r\n"

	for _, header := range h.Response.HTTPHeaders {
		response += header + "\r\n"
	}

	response += "\r\n"
	if len(h.Response.Body) > 0 {
		response += h.Response.Body
	}
	return []byte(response)
}

func (h *HTTP) read() {
	reqBuffer := make([]byte, 1024)
	h.logger.Info("Reading request...")

	d, err := h.Connection.Read(reqBuffer)
	if err != nil {
		h.logger.Error("Error reading from connection: " + err.Error())
		os.Exit(1)
	}
	h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))

	h.serializeRequest(reqBuffer)
	h.serializeResponse()
}

func (h *HTTP) write() {
	response := h.deserializeResponse()
	d, err := h.Connection.Write(response)
	if err != nil {
		h.logger.Error("Error writing to connection: " + err.Error())
		os.Exit(1)
	}
	h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))

}

func (h *HTTP) ServeRequests(ip string, port string) {
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		h.logger.Error("Failed to bind to port " + port + err.Error())
		os.Exit(1)
	}

	h.Listener = l
	h.accept()

	defer h.Connection.Close()

	h.read()
	h.write()
}
