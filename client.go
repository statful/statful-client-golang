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

type Client struct {
	buffer buffer

	ticker     *time.Ticker
	tickerDone chan bool

	globalTags Tags
}

type Configuration struct {
	DisableAutoFlush bool
	DryRun           bool
	Tags             Tags
	FlushSize        int
	FlushInterval    time.Duration

	Logger Logger
	Sender Sender
}

func New(cfg Configuration) *Client {
	statful := &Client{
		buffer: buffer{
			metricCount:      0,
			flushSize:        cfg.FlushSize,
			dryRun:           cfg.DryRun,
			disableAutoFlush: cfg.DisableAutoFlush,
			mu:               sync.Mutex{},
			stdBuf:           make([]string, 0, cfg.FlushSize),
			aggBuf:           make(map[Aggregation]map[AggregationFrequency][]string),
			Sender:           cfg.Sender,
			Logger:           cfg.Logger,
		},
		globalTags: cfg.Tags,
	}

	if cfg.FlushInterval > 0 && !cfg.DisableAutoFlush {
		statful.StartFlushInterval(cfg.FlushInterval)
	}

	return statful
}

// Starts a go routine that periodically flushes the metrics from buffer
// If AutoFlush is deactivated it just send metrics synchronously.
// Returns a function that stops the timer.
func (c *Client) StartFlushInterval(interval time.Duration) {
	if c.buffer.disableAutoFlush {
		return
	}

	if interval < MinFlushInterval {
		interval = MinFlushInterval
	}

	c.ticker = time.NewTicker(interval)
	c.tickerDone = make(chan bool)

	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.buffer.Flush()
			case <-c.tickerDone:
				break
			}
		}
	}()
}

func (c *Client) StopFlushInterval() {
	if c.ticker != nil {
		c.ticker.Stop()
		c.tickerDone <- true
	}
}

func (c *Client) Counter(name string, value float64, tags Tags) {
	c.Put(name, value, tags, time.Now().Unix(), counterAggregations, Freq10s)
}

func (c *Client) CounterAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	c.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (c *Client) Gauge(name string, value float64, tags Tags) {
	c.Put(name, value, tags, time.Now().Unix(), gaugeAggregations, Freq10s)
}

func (c *Client) GaugeAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	c.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (c *Client) Timer(name string, value float64, tags Tags) {
	c.Put(name, value, tags, time.Now().Unix(), timerAggregations, Freq10s)
}

func (c *Client) TimerAggregated(name string, value float64, tags Tags, aggregation Aggregation, frequency AggregationFrequency) {
	c.PutAggregated(name, value, tags, time.Now().Unix(), aggregation, frequency)
}

func (c *Client) Put(name string, value float64, tags Tags, timestamp int64, aggs Aggregations, freq AggregationFrequency) error {
	return c.buffer.Put(name, value, tags.Merge(c.globalTags), timestamp, aggs, freq)
}

func (c *Client) PutAggregated(name string, value float64, tags Tags, timestamp int64, agg Aggregation, freq AggregationFrequency) error {
	return c.buffer.PutAggregated(name, value, tags.Merge(c.globalTags), timestamp, agg, freq)
}

func (c *Client) Flush() {
	c.buffer.Flush()
}

// FlushError flushes the client buffer and returns a FlushErr error if any errors happen.
func (c *Client) FlushError() error {
	return c.buffer.FlushError()
}
