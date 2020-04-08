package statful

import (
	"strings"
)

type ProxyMetricsSender struct {
	Client Client
}

func (p *ProxyMetricsSender) Put(m *Metric) error {
	p.Client.Send(strings.NewReader(m.String()))
	return nil
}

func (p *ProxyMetricsSender) Flush() {
	// do nothing
}
