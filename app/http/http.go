package http

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
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

func parseHeaders(headers []string) string {
	for _, header := range headers {
		if strings.HasPrefix(header, "User-Agent") {
			return header[len("User-Agent: "):]
		}
	}
	return ""
}

func fetchResponse(request request, dirFlag bool, directory string) response {
	successful := []byte("HTTP/1.1 200 OK")
	unSuccessful := []byte("HTTP/1.1 404 Not Found")

	contentType := "text/plain"

	var responseStartLine []byte
	body := []byte("")
	userAgent := ""

	switch {
	case request.Path == "/":
		responseStartLine = successful

	case strings.HasPrefix(request.Path, "/echo"):
		responseStartLine = successful
		body = []byte(request.Path[len("/echo"):])
		if len(body) > 1 && body[0] == '/' {
			body = body[1:]
		} else {
			body = []byte("")
		}

	case strings.HasPrefix(request.Path, "/user-agent"):
		responseStartLine = successful
		userAgent = parseHeaders(request.HTTPHeaders)
		body = []byte(userAgent)

	case strings.HasPrefix(request.Path, "/files"):
		if dirFlag == false {
			responseStartLine = unSuccessful
		} else {
			filePath, err := filepath.Abs(directory + request.Path[len("/files"):])
			if err != nil {
				os.Exit(1)
			}
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				responseStartLine = unSuccessful
			} else {
				responseStartLine = successful
				fileContents, err := os.ReadFile(filePath)
				if err != nil {
					os.Exit(1)
				}
				body = fileContents
				contentType = "application/octet-stream"
			}
		}

	default:
		responseStartLine = unSuccessful

	}

	var contentLengthString string
	if len(body) > 0 {
		contentLengthString = fmt.Sprint(len(body))
	} else {
		contentLengthString = "0"
	}

	return response{
		Status:      string(responseStartLine),
		Body:        string(body),
		HTTPHeaders: []string{"Content-Type: " + contentType, "Content-Length: " + contentLengthString},
	}
}

type HTTP struct {
	logger    *zap.Logger
	Listener  net.Listener
	directory string
	dirFlag   bool
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
	requestLine := strings.Split(string(reqBuffer), "\r\n")
	startLineSections := strings.Split(requestLine[0], " ")
	req.HTTPMethod = startLineSections[0]
	req.Path = startLineSections[1]
	req.HTTPVersion = startLineSections[2]

	for _, header := range requestLine[1:] {
		if header == "" {
			break
		}
		req.HTTPHeaders = append(req.HTTPHeaders, header)
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
	response = fetchResponse(request, h.dirFlag, h.directory)
	h.write(conn, response)
}

func (h *HTTP) ServeRequests(ip string, port string) {
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		h.logger.Error("Failed to bind to port " + port + ": " + err.Error())
		os.Exit(1)
	}

	h.Listener = l
	for {
		c, err := h.Listener.Accept()
		if err != nil {
			h.logger.Error("Error accepting connection: " + err.Error())
			continue
		}

		h.logger.Info("Accepted connection", zap.String("remote", c.RemoteAddr().String()), zap.String("local", c.LocalAddr().String()))

		go h.handleConnection(c)
	}
}
