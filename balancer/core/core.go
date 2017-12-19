package core

import (
	"bufio"
	"net"
)

// MaxTicks defines the length of the stats for the UI
const MaxTicks = 40

type Context struct {
	HTTPS    bool
	Hostname string
	Conn     BufferedConn
}

type HostStats struct {
	Healthy            bool   `json:"healthy"`
	TotalConnections   int64  `json:"totalConnections"`
	ActiveConnections  uint   `json:"activeConnections"`
	RefusedConnections uint64 `json:"refusedConnections"`
}

type RouterHostWithStats struct {
	ClusterKey string      `json:"clusterKey"`
	HostIP     string      `json:"hostIP"`
	HTTPPort   int         `json:"httpPort"`
	HTTPSPort  int         `json:"httpsPort"`
	Stats      []HostStats `json:"stats"`
}

type GlobalStats struct {
	Mutation           string                         `json:"mutation"`
	Hosts              map[string]RouterHostWithStats `json:"hosts"`
	Ticks              []string                       `json:"ticks"`
	OverallConnections []uint                         `json:"overallConnections"`
	UnhealthyHosts     []int                          `json:"unhealthyHosts"`
	HealthyHosts       []int                          `json:"healthyHosts"`
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
