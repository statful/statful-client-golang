package statful

import (
	"bytes"
	"strings"
	"sync"
)

type BufferedMetricsSender struct {
	mu                 sync.Mutex

	FlushSize int
	Buf       bytes.Buffer
	Client    Client
}

func (bms *BufferedMetricsSender) Put(m *Metric) error {
	bms.mu.Lock()
	defer bms.mu.Unlock()

	// put the metric in the buffer
	bms.Buf.ReadFrom(strings.NewReader(m.String()))

	if bms.Buf.Len() >= bms.FlushSize {
		bms.Client.Send(&bms.Buf)
	}
	return nil
}

func (bms *BufferedMetricsSender) Flush() {
	bms.mu.Lock()
	defer bms.mu.Unlock()
	if bms.Buf.Len() > 0 {
		bms.Client.Send(&bms.Buf)
	}
}
