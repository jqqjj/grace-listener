package graceful

import (
	"net"
	"sync"
)

type GraceConnection struct {
	net.Conn
	wg *sync.WaitGroup
}

func NewGraceConnection(wg *sync.WaitGroup, c net.Conn) *GraceConnection {
	wg.Add(1)
	return &GraceConnection{c, wg}
}

func (c *GraceConnection) Close() error {
	defer c.wg.Done()
	return c.Conn.Close()
}
