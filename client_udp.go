package statful

import (
	"errors"
	"io"
	"net"
	"time"
)

type UdpClient struct {
	Address string
	Timeout time.Duration
}

func (u *UdpClient) Put(reader io.Reader) error {
	conn, err := net.Dial("udp", u.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, reader)
	if err != nil {
		return err
	}

	return nil
}

func (u *UdpClient) PutAggregated(reader io.Reader) error {
	return errors.New("UNSUPPORTED_OPERATION")
}
