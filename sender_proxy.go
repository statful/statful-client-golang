package statful

import (
	"strings"
)

type ProxyMetricsSender struct {
	Client Client
}

func (p *ProxyMetricsSender) Put(metrics []*Metric) error {
	var b strings.Builder
	for _, m := range metrics {
		b.WriteString(m.String())
	}
	p.Client.Send(strings.NewReader(b.String()))
	return nil
}

func (p *ProxyMetricsSender) Flush() {
	// do nothing
}
