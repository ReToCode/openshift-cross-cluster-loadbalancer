package balancer

import "net"

type Context struct {
	Hostname string
	Conn net.Conn
}

type ReadWriteCount struct {
	CountRead uint
	CountWrite uint
}

func (rw ReadWriteCount) IsZero() bool {
	return rw.CountRead == 0 && rw.CountWrite == 0
}