package core

import (
	"io"
	"net"

	"github.com/sirupsen/logrus"
)

const (
	bufferSize = 16 * 1024
)

func Proxy(to BufferedConn, from BufferedConn) <-chan bool {
	doneChan := make(chan bool)

	go func() {
		err := copyData(to, from)
		e, ok := err.(*net.OpError)
		if err != nil && (!ok || e.Err.Error() != "use of closed network connection") {
			logrus.Warn(err)
		}

		to.Close()
		from.Close()

		doneChan <- true
	}()

	return doneChan
}

func copyData(to io.Writer, from io.Reader) error {
	buf := make([]byte, bufferSize)
	var err error

	for {
		readN, readErr := from.Read(buf)
		if readN > 0 {
			writeN, writeErr := to.Write(buf[0:readN])

			if writeErr != nil {
				err = writeErr
				break
			}

			if readN != writeN {
				err = io.ErrShortWrite
				break
			}
		}

		if readErr == io.EOF {
			break
		}

		if readErr != nil {
			err = readErr
			break
		}
	}

	return err
}
