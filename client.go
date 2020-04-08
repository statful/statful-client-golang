package statful

import (
	"io"
)

type Client interface {
	Send(data io.Reader) error
}

