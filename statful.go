package statful

import (
	"sync"
	"time"
)

const (
	MinFlushInterval = 50 * time.Millisecond
)

var (
	counterAggregations = Aggregations{AggCount: struct{}{}, AggSum: struct{}{}}
	gaugeAggregations   = Aggregations{AggLast: struct{}{}}
	timerAggregations   = Aggregations{AggAvg: struct{}{}, AggCount: struct{}{}, AggP90: struct{}{}}
)

type statful struct {
	sender bufferedMetricsSender

	ticker     *time.Ticker
	tickerDone chan bool

	globalTags Tags
}

type Options struct {
	DryRun        bool
	Tags          Tags
	FlushSize     int
	FlushInterval time.Duration

	Logger Logger
	Client Client
}

func New(o Options) *statful {
	statful := &statful{
		sender: bufferedMetricsSender{
			metricCount: 0,
			flushSize:   o.FlushSize,
			dryRun:      o.DryRun,
			mu:          sync.Mutex{},
			stdBuf:      make([]string, 0, o.FlushSize),
			aggBuf:      make(map[Aggregation]map[AggregationFrequency][]string),
			Client:      o.Client,
			Logger:      o.Logger,
		},
		globalTags: o.Tags,
	}

	if o.FlushInterval > 0 {
		statful.StartFlushInterval(o.FlushInterval)
	}

	return statful
}

// Starts a go routine that periodically flushes the metrics of MetricsSender
// Returns a function that stops the timer.
func (s *statful) StartFlushInterval(interval time.Duration) {
	if interval < MinFlushInterval {
		interval = MinFlushInterval
	}

	s.ticker = time.NewTicker(interval)
	s.tickerDone = make(chan bool)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.sender.Flush()
			case <-s.tickerDone:
				break
			}
		}
	}()
}

func (s *statful) StopFlushInterval() {
	s.ticker.Stop()
	s.tickerDone <- true
}

func (s *statful) Counter(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), counterAggregations, Freq10s)
}

func (s *statful) CounterAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	s.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (s *statful) Gauge(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), gaugeAggregations, Freq10s)
}

func (s *statful) GaugeAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	s.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (s *statful) Timer(name string, value float64, tags Tags) {
	s.Put(name, value, tags, time.Now().Unix(), timerAggregations, Freq10s)
}

func (s *statful) TimerAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	s.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (s *statful) Put(name string, value float64, tags Tags, timestamp int64, aggs Aggregations, freq AggregationFrequency) error {
	return s.sender.Send(name, value, tags.Merge(s.globalTags), timestamp, aggs, freq)
}

func (s *statful) PutAggregated(name string, value float64, tags Tags, timestamp int64, agg Aggregation, freq AggregationFrequency) error {
	return s.sender.SendAggregated(name, value, tags.Merge(s.globalTags), timestamp, agg, freq)
}

func (s *statful) Flush() {
	s.sender.Flush()
}
