package statful

import (
	"io"
	"io/ioutil"
)

type ChannelClient struct {
	data chan<- []byte
}

func (c *ChannelClient) Send(data io.Reader) error {
	all, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	c.data <- all
	return nil
}
