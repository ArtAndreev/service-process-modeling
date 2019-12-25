package client

type Config struct {
	RequestPath  []string // Full request path, for example [1, 2] services
	ResponsePath []string // Full response path, for example [1]

	Count int // Client count (number of requests)
}
