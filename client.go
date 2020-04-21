package statful

import (
	"io"
)

type Client interface {
	Send(data io.Reader) error
}

type FuncClient func (io.Reader) error

func (f FuncClient) Send(data io.Reader) error {
	return f(data)
}
