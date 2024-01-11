# Http-Server-Go

This is my implementation of the ["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview) from [Codecrafters.io](https://app.codecrafters.io).

[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the
protocol that powers the web. In this challenge, I build a HTTP/1.1 server
that is capable of serving multiple clients. I can only use networking for TCP and not HTTP packages. 

I refreshed my understanding about TCP [HTTP request syntax](https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html),
and more.

**Note**: To try to implement this yourself in your language of choice, head over to
[codecrafters.io](https://codecrafters.io) to try the challenge.

# Design

The entry point for the HTTP server implementation is in `app/server.go`.

## Types

### HTTP server

```go
type HTTP struct {
	directory string
	dirFlag   bool
	logger    *zap.Logger
	listener  net.Listener
}
```

* The server utilized [uber/zap](https://github.com/uber-go/zap) to have structured logging.
* Go provides a builtin [net](https://pkg.go.dev/net) package for handling TCP connections.

### Request and Response

```go
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
```


The HTTP type doesn't maintain any state for the responses. This allows us to handle multiple connections to our server concurrently. 

# Over All Thoughts On Challenge

This was a very fun challenge. Obviously my implementation isn't a full implementation. However I had fun writing this out.

I think this code can be useful for anyone else trying the challenge themselves and looking to compare approaches to get better. Feel free to open any PR's if you would want to add to this implementation.
