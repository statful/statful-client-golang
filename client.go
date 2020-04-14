package statful

import (
	"io"
)

type Client interface {
	Send(data io.Reader) error
}

type WriterClient struct {
	Writer io.Writer
}

func (c *WriterClient) Send(data io.Reader) error {
	_, err := io.Copy(c.Writer, data)
	return err
}