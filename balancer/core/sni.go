package core

import (
	"bytes"
	"crypto/tls"
	"io"
	"net"
	"sync"
	"time"
	"strings"
)

const MAX_HEADER_SIZE = 16385

var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, MAX_HEADER_SIZE)
	},
}

// Conn delegates all calls to net.Conn, but Read to reader
type Conn struct {
	reader   io.Reader
	net.Conn //delegate
}

func (c Conn) Read(b []byte) (n int, err error) {
	return c.reader.Read(b)
}

type bufferConn struct {
	io.Reader
}

type localAddr struct{}

func (l localAddr) String() string {
	return "127.0.0.1"
}

func (l localAddr) Network() string {
	return "tcp"
}

func newBufferConn(b []byte) *bufferConn {
	return &bufferConn{bytes.NewReader(b)}
}

func (c bufferConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (c bufferConn) Close() error {
	return nil
}

func (c bufferConn) LocalAddr() net.Addr {
	return localAddr{}
}

func (c bufferConn) RemoteAddr() net.Addr {
	return localAddr{}
}

func (c bufferConn) SetDeadline(t time.Time) error {
	return nil
}

func (c bufferConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c bufferConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// Sniff sniffs hostname from ClientHello message (if any),
// returns sni.Conn, filling it's Hostname field
func Sniff(conn net.Conn, readTimeout time.Duration) (net.Conn, string, error) {
	buf := pool.Get().([]byte)
	defer pool.Put(buf)

	conn.SetReadDeadline(time.Now().Add(readTimeout))
	i, err := conn.Read(buf)

	if err != nil {
		return nil, "", err
	}

	conn.SetReadDeadline(time.Time{}) // Reset read deadline

	hostname := extractHostname(buf[0:i])

	data := make([]byte, i)
	copy(data, buf) // Since we reuse buf between invocations, we have to make copy of data
	mreader := io.MultiReader(bytes.NewBuffer(data), conn)

	// Wrap connection so that it will Read from buffer first and remaining data
	// from initial conn
	return Conn{mreader, conn}, hostname, nil
}

func extractHostname(buf []byte) string {
	conn := tls.Server(newBufferConn(buf), &tls.Config{})
	defer conn.Close()
	conn.Handshake()
	return strings.TrimSpace(conn.ConnectionState().ServerName)
}
