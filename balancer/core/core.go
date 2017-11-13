package core

import (
	"net"
	"bufio"
)

type Context struct {
	Hostname string
	Conn     BufferedConn
}

type ReadWriteCount struct {
	CountRead  uint
	CountWrite uint
}

func (rwc ReadWriteCount) IsZero() bool {
	return rwc.CountRead == 0 && rwc.CountWrite == 0
}

type BufferedConn struct {
	Reader *bufio.Reader
	net.Conn
}

func NewBufferedConn(c net.Conn) BufferedConn {
	return BufferedConn{
		bufio.NewReader(c),
		c,
	}
}

func (b BufferedConn) Peek(n int) ([]byte, error) {
	return b.Reader.Peek(n)
}

func (b BufferedConn) Read(p []byte) (int, error) {
	return b.Reader.Read(p)
}
