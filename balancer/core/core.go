package core

import "net"

type Context struct {
	Hostname string
	Conn     net.Conn
}

type ReadWriteCount struct {
	CountRead uint
	CountWrite uint
}

func (rwc ReadWriteCount) IsZero() bool {
	return rwc.CountRead == 0 && rwc.CountWrite == 0
}
