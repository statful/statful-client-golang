package statful

import (
	"io"
	"net"
	"time"
)

type UdpClient struct {
	Address string
	Timeout time.Duration
}

func (u *UdpClient) Send(reader io.Reader) error {
	conn, err := net.Dial("udp", u.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	wb, err := io.Copy(conn, reader)
	if err != nil {
		return err
	}
	debugLog("Flushed", wb, "bytes")

	return nil
}
