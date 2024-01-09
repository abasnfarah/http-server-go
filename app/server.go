package main

import (
	"flag"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func main() {
	directoryFlagPtr := flag.String("directory", "", "")
	flag.Parse()

	httpServer := http.NewHTTPServer(*directoryFlagPtr)
	httpServer.ServeRequests("0.0.0.0", "4221")
}
