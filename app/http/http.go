package http

import (
	"net"
	"os"
	"strings"

	"go.uber.org/zap"
)

type HTTP struct {
	directory string
	dirFlag   bool
	logger    *zap.Logger
	listener  net.Listener
}

func NewHTTPServer(directoryFlagPtr string) *HTTP {
	logger, _ := zap.NewProduction()
	logger.Info("Starting HTTP Server")
	if directoryFlagPtr == "" {
		return &HTTP{logger: logger, directory: directoryFlagPtr, dirFlag: false}
	}
	return &HTTP{logger: logger, directory: directoryFlagPtr, dirFlag: true}
}

func (h *HTTP) deserializeRequest(reqBuffer []byte, req *request) {
	h.logger.Info("Deserializing request...")
	requestLine := strings.Split(string(reqBuffer), "\r\n")
	startLineSections := strings.Split(requestLine[0], " ")
	req.HTTPMethod = startLineSections[0]
	req.Path = startLineSections[1]
	req.HTTPVersion = startLineSections[2]

	var j int
	for i, header := range requestLine[1:] {
		if header == "" {
			j = i + 1
			break
		}
		req.HTTPHeaders = append(req.HTTPHeaders, header)
	}

	for _, body := range requestLine[j:] {
		if body == "" {
			continue
		}
		body = strings.TrimRight(body, "\x00")
		body = strings.ReplaceAll(body, "\\n", "\n")
		body = strings.ReplaceAll(body, "\\r", "\r")
		body += "\r\n"
		req.Body = body
	}

	h.logger.Info("Deserialized Request: ", zap.Any("request", req))
}

func (h *HTTP) serializeResponse(res response) []byte {
	response := res.Status + "\r\n"

	for _, header := range res.HTTPHeaders {
		response += header + "\r\n"
	}

	response += "\r\n"
	if len(res.Body) > 0 {
		response += res.Body
		response += "\r\n"
	}
	return []byte(response)
}

func (h *HTTP) read(conn net.Conn, request *request) {
	reqBuffer := make([]byte, 1024)
	h.logger.Info("Reading request...")

	d, err := conn.Read(reqBuffer)
	if err != nil {
		h.logger.Error("Error reading from connection: " + err.Error())
		os.Exit(1)
	}
	h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))

	h.deserializeRequest(reqBuffer, request)
}

func (h *HTTP) write(conn net.Conn, response response) {
	resp := h.serializeResponse(response)
	d, err := conn.Write(resp)
	if err != nil {
		h.logger.Error("Error writing to connection: " + err.Error())
		os.Exit(1)
	}
	h.logger.Info("READ: Number of bytes recieved: ", zap.Int("bytes", d))
}

func (h *HTTP) handleConnection(conn net.Conn) {
	defer conn.Close()

	var request request
	var response response

	h.read(conn, &request)

	if request.HTTPMethod == "GET" {
		response = fetchGETResponse(request, h.dirFlag, h.directory)
	} else if request.HTTPMethod == "POST" {
		response = fetchPOSTResponse(request, h.dirFlag, h.directory)
	}

	h.write(conn, response)
}

func (h *HTTP) ServeRequests(ip string, port string) {
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		h.logger.Error("Failed to bind to port " + port + ": " + err.Error())
		os.Exit(1)
	}

	h.listener = l
	for {
		c, err := h.listener.Accept()
		if err != nil {
			h.logger.Error("Error accepting connection: " + err.Error())
			continue
		}

		h.logger.Info("Accepted connection", zap.String("remote", c.RemoteAddr().String()), zap.String("local", c.LocalAddr().String()))

		go h.handleConnection(c)
	}
}
