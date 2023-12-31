package http

import (
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
	HTTPMethod  string
	Path        string
	HTTPVersion string
	HTTPHeaders []string
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

func (h *HTTP) deserialize(request []byte) {
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

func (h *HTTP) read() {
	reqBuffer := make([]byte, 1024)
	h.logger.Info("Reading request...")

	d, err := h.Connection.Read(reqBuffer)
	if err != nil {
		h.logger.Error("Error reading from connection: " + err.Error())
		os.Exit(1)
	}
	h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))

	h.deserialize(reqBuffer)
}

func (h *HTTP) write() {
	successful := []byte("HTTP/1.1 200 OK\r\n\r\n")
	unSuccessful := []byte("HTTP/1.1 404 Not Found\r\n\r\n")

	if h.Request.Path != "/" {
		d, err := h.Connection.Write(unSuccessful)
		if err != nil {
			h.logger.Error("Error writing to connection: " + err.Error())
			os.Exit(1)
		}
		h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))
	} else {
		d, err := h.Connection.Write(successful)
		if err != nil {
			h.logger.Error("Error writing to connection: " + err.Error())
			os.Exit(1)
		}
		h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))
	}

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
