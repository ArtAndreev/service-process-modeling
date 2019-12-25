package client

type PathType int

const (
	PathRequest PathType = iota
	PathResponse
)

type Path struct {
	PathType PathType
	Next     string
}
