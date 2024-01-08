package main

import (
	"flag"
	"fmt"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func main() {
	directoryFlagPtr := flag.String("directory", "", "")
	flag.Parse()
	fmt.Println(*directoryFlagPtr)
	httpServer := http.NewHTTPServer(*directoryFlagPtr)

	httpServer.ServeRequests("0.0.0.0", "4221")
}
