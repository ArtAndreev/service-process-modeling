package service

import (
	"time"
)

type Config struct {
	MaxClientConn int // max client connections service can handle

	RequestProcessTime  time.Duration // time need for processing every request
	ResponseProcessTime time.Duration // time need for processing response back to client

	Parallel int // number or parallel processing
}
