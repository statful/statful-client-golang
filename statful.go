package statful

import (
	"time"
)

const (
	MinFlushInterval = 50 * time.Millisecond
)

var (
	counterAggregations = Aggregations{AggCount: struct{}{}, AggSum: struct{}{}}
	gaugeAggregations   = Aggregations{AggLast: struct{}{}}
	histogramAggregations = Aggregations{AggAvg:   struct{}{}, AggCount: struct{}{}, AggP90:   struct{}{}}
)

type Statful struct {
	//
	Sender MetricsSender

	GlobalTags Tags
	DryRun     bool
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
func (s *Statful) Counter(name string, value float64) {
	s.CounterWithTags(name, value, Tags{})
}

// Creates a new counter and sends it using the MetricsSender
// The counter is created with the default aggregations count and sum
func (s *Statful) CounterWithTags(name string, value float64, tags Tags) {
	s.Put(&Metric{
		Name:  name,
		Value: value,
		Tags:  tags,
		Aggs:  counterAggregations,
		Freq:  Freq10s,
	})
}

// Creates a new gauge and sends it using the MetricsSender
// The gauge is created with the default aggregations last
func (s *Statful) Gauge(name string, value float64) {
	s.GaugeWithTags(name, value, Tags{})
}

// Creates a new gauge and sends it using the MetricsSender
// The gauge is created with the default aggregations last
func (s *Statful) GaugeWithTags(name string, value float64, tags Tags) {
	s.Put(&Metric{
		Name:  name,
		Value: value,
		Tags:  tags,
		Aggs:  gaugeAggregations,
		Freq:  Freq10s,
	})
}

// Creates a new histogram and sends it using the MetricsSender
// The histogram is created with the default aggregations avg, p90 and count
func (s *Statful) Histogram(name string, value float64) {
	s.HistogramWithTags(name, value, Tags{})
}

// Creates a new histogram and sends it using the MetricsSender
// The histogram is created with the default aggregations avg, p90 and count
func (s *Statful) HistogramWithTags(name string, value float64, tags Tags) {
	s.Put(&Metric{
		Name:  name,
		Value: value,
		Tags:  tags,
		Aggs:  histogramAggregations,
		Freq:  Freq10s,
	})
}

// Sends metric m using the MetricsSender.
func (s *Statful) Put(m *Metric) error {
	m.Tags.Merge(s.GlobalTags)
	if s.DryRun {
		debugLog(m.String())
	} else {
		s.Sender.Put(&Metric{
			Name:  m.Name,
			Value: m.Value,
			Tags:  m.Tags.Merge(s.GlobalTags),
			Aggs:  m.Aggs,
			Freq:  m.Freq,
		})
	}

	return nil
}
