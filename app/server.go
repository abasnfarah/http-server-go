package main

import "github.com/codecrafters-io/http-server-starter-go/app/http"

func main() {
	httpServer := http.NewHTTPServer()

	httpServer.ServeRequests("0.0.0.0", "4221")
}
