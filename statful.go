package statful

import (
	"time"
)

const (
	MinFlushInterval = 50 * time.Millisecond
)

var (
	counterAggregations   = Aggregations{AggCount: struct{}{}, AggSum: struct{}{}}
	gaugeAggregations     = Aggregations{AggLast: struct{}{}}
	histogramAggregations = Aggregations{AggAvg: struct{}{}, AggCount: struct{}{}, AggP90: struct{}{}}
)

type Statful struct {
	Sender MetricsSender
	GlobalTags Tags
}

// Starts a go routine that periodically flushes the metrics of MetricsSender
// Returns a function that stops the timer.
func (s *Statful) StartFlushInterval(interval time.Duration) func() {
	if interval < MinFlushInterval {
		interval = MinFlushInterval
	}
	ticker := time.NewTicker(interval)
	tickerDone := make(chan bool)
	go func(s *Statful, ticker *time.Ticker, tickerDone chan bool) {
		for {
			select {
			case <-ticker.C:
				s.Sender.Flush()
			case <-tickerDone:
				break
			}
		}
	}(s, ticker, tickerDone)

	return func() {
		ticker.Stop()
		tickerDone <- true
	}
}

// Creates a new counter and sends it using the MetricsSender
// The counter is created with the default aggregations count and sum
func (s *Statful) Counter(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), counterAggregations, Freq10s)
}

// Creates a new gauge and sends it using the MetricsSender
// The gauge is created with the default aggregations last
func (s *Statful) Gauge(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), gaugeAggregations, Freq10s)
}

// Creates a new histogram and sends it using the MetricsSender
// The histogram is created with the default aggregations avg, p90 and count
func (s *Statful) Histogram(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), histogramAggregations, Freq10s)
}

// Sends metric m using the MetricsSender.
func (s *Statful) Put(name string, value float64, tags Tags, timestamp int64, aggs Aggregations, freq AggregationFrequency) error {
	s.Sender.Put([]*Metric{
		{
			Name:  name,
			Value: value,
			Timestamp: timestamp,
			Tags:  tags.Merge(s.GlobalTags),
			Aggs:  aggs,
			Freq:  freq,
		},
	})

	return nil
}
