package client

type Client struct {
	name     string
	pathType PathType

	paths []Path
	next  int
}

func New(name string, paths []Path) *Client {
	return &Client{
		name:     name,
		pathType: PathRequest,

		paths: paths,
	}
}

func (c *Client) GetRequestName() string {
	return c.name
}

func (c *Client) GetPathType() PathType {
	return c.pathType
}

func (c *Client) GetNextPath() (Path, bool) {
	if c.next < len(c.paths) {
		next := c.paths[c.next]
		c.next++
		c.pathType = next.PathType

		return next, true
	}

	return Path{}, false
}
