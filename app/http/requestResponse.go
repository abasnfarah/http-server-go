package http

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	SUCCESSFUL_GET  = "HTTP/1.1 200 OK"
	SUCCESSFUL_POST = "HTTP/1.1 201 OK"
	FAILED_GET      = "HTTP/1.1 404 Not Found"
	FAILED_POST     = "HTTP/1.1 404 Not Found"
)

type request struct {
	Body        string
	HTTPMethod  string
	HTTPVersion string
	HTTPHeaders []string
	Path        string
}

type response struct {
	Body        string
	HTTPHeaders []string
	Status      string
}

func parseHeaders(headers []string) string {
	for _, header := range headers {
		if strings.HasPrefix(header, "User-Agent") {
			return header[len("User-Agent: "):]
		}
	}
	return ""
}

func createDefaultResponse(request request, responseStartLine, body, contentType string) response {
	return response{
		Body:        string(body),
		HTTPHeaders: []string{"Content-Type: " + contentType, "Content-Length: " + fmt.Sprint(len(body))},
		Status:      string(responseStartLine),
	}
}

func fetchGETResponse(request request, dirFlag bool, directory string) response {
	contentType := "text/plain"
	body := ""
	userAgent := ""

	var response response

	switch {
	case request.Path == "/":
		response = createDefaultResponse(request, SUCCESSFUL_GET, body, contentType)

	case strings.HasPrefix(request.Path, "/echo"):
		body = request.Path[len("/echo"):]
		if len(body) > 1 && body[0] == '/' {
			body = body[1:]
		} else {
			body = ""
		}

		response = createDefaultResponse(request, SUCCESSFUL_GET, body, contentType)

	case strings.HasPrefix(request.Path, "/user-agent"):
		userAgent = parseHeaders(request.HTTPHeaders)
		body = userAgent

		response = createDefaultResponse(request, SUCCESSFUL_GET, body, contentType)

	case strings.HasPrefix(request.Path, "/files"):
		var responseStartLine string

		if !dirFlag {
			responseStartLine = FAILED_GET
		} else {
			filePath, _ := filepath.Abs(directory + request.Path[len("/files"):])

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				responseStartLine = FAILED_GET

			} else {
				responseStartLine = SUCCESSFUL_GET
				fileContents, _ := os.ReadFile(filePath)
				body = string(fileContents)
				contentType = "application/octet-stream"
			}
		}

		response = createDefaultResponse(request, responseStartLine, body, contentType)

	default:
		response = createDefaultResponse(request, FAILED_GET, body, contentType)

	}

	return response
}

func fetchPOSTResponse(request request, dirFlag bool, directory string) response {
	contentType := "text/plain"
	body := ""

	var response response

	switch {
	case strings.HasPrefix(request.Path, "/files"):
		var responseStartLine string

		if !dirFlag {
			responseStartLine = FAILED_POST
		} else {
			filePath, _ := filepath.Abs(directory + request.Path[len("/files"):])

			fmt.Println("Writing to file: ", filePath)
			fmt.Println("Body: ", request.Body)
			fmt.Printf("Body: %q\n", request.Body)
			if err := os.WriteFile(filePath, []byte(request.Body), 0644); err != nil {
				os.Exit(1)
			}
			responseStartLine = SUCCESSFUL_POST
		}

		response = createDefaultResponse(request, responseStartLine, body, contentType)

	default:
		response = createDefaultResponse(request, FAILED_POST, body, contentType)

	}

	return response
}
