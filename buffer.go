package statful

import (
	"strings"
	"sync"
)

type buffer struct {
	metricCount int
	flushSize   int
	dryRun      bool

	mu  sync.Mutex

	stdBuf []string
	aggBuf map[Aggregation]map[AggregationFrequency][]string

	Sender Sender
	Logger Logger
}

func (s *buffer) Put(name string, value float64, tags Tags, timestamp int64, aggregations Aggregations, frequency AggregationFrequency) error {
	// put the metric in the buffer
	s.mu.Lock()
	s.stdBuf = append(s.stdBuf, MetricToString(name, value, tags, timestamp, aggregations, frequency))
	s.metricCount++

	if s.metricCount >= s.flushSize {
		stdBuf, aggBuf := s.drainBuffers()
		go s.flushBuffers(stdBuf, aggBuf)
	}
	s.mu.Unlock()

	return nil
}

func (s *buffer) PutAggregated(name string, value float64, tags Tags, timestamp int64, aggregation Aggregation, frequency AggregationFrequency) error {
	// put the metric in the buffer
	s.mu.Lock()
	s.aggBuf[aggregation][frequency] = append(s.aggBuf[aggregation][frequency], MetricToString(name, value, tags, timestamp, Aggregations{}, 0))
	s.metricCount++

	if s.metricCount >= s.flushSize {
		stdBuf, aggBuf := s.drainBuffers()
		go s.flushBuffers(stdBuf, aggBuf)
	}
	s.mu.Unlock()

	return nil
}

func (s *buffer) Flush() {
	s.mu.Lock()
	stdBuf, aggBuf := s.drainBuffers()
	s.mu.Unlock()

	s.flushBuffers(stdBuf, aggBuf)
}

func (s *buffer) drainBuffers() ([]string, map[Aggregation]map[AggregationFrequency][]string) {
	var stdBuf []string
	var aggBuf map[Aggregation]map[AggregationFrequency][]string

	if s.metricCount > 0 {
		stdBuf = s.stdBuf
		s.stdBuf = make([]string, 0, s.flushSize)

		aggBuf = s.aggBuf
		s.aggBuf = make(map[Aggregation]map[AggregationFrequency][]string)

		s.metricCount = 0
	}

	return stdBuf, aggBuf
}

func (s *buffer) flushBuffers(stdBuf []string, aggBuf map[Aggregation]map[AggregationFrequency][]string) {
	if len(stdBuf) > 0 {
		if s.dryRun {
			for _, m := range stdBuf {
				s.Logger.Println("Dry metric:", m)
			}
		} else {
			err := s.Sender.Send(strings.NewReader(strings.Join(stdBuf, "\n")))
			if err != nil {
				s.Logger.Println("Failed to send metrics", err)
			}
		}
	}

	for agg, freqs := range aggBuf {
		for freq, buf := range freqs {
			if s.dryRun {
				s.Logger.Println("Dry aggregated metric:", buf, agg, freq)
			} else {
				s.Sender.SendAggregated(strings.NewReader(strings.Join(buf, "\n")), agg, freq)
			}
		}
	}

}