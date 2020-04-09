package statful

import (
	"bytes"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestStatfulSDK(t *testing.T) {
	metricsData := make(chan []byte, 1)

	statfulWithoutGlobalTags := Statful{
		Sender: &BufferedMetricsSender{
			FlushSize: 1000,
			Buf:       bytes.Buffer{},
			Client: &ChannelClient{
				data: metricsData,
			},
		},
		GlobalTags: Tags{},
	}
	statfulWithGlobalTags := Statful{
		Sender: &BufferedMetricsSender{
			FlushSize: 1000,
			Buf:       bytes.Buffer{},
			Client: &ChannelClient{
				data: metricsData,
			},
		},
		GlobalTags: Tags{"global": "tag"},
	}
	//cancelPeriodicFlush := statfulMetrics.StartFlushInterval(1 * time.Second)

	SetDebugLogger(t.Log)
	SetErrorLogger(t.Log)

	scenarios := []struct {
		description      string
		statful          Statful
		metricsProducer  func(s *Statful)
		totalFlushes     int
		totalMetricsSent int
		metricsSent      []*regexp.Regexp
	}{
		// counters only
		{
			description: "3 Counters with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Counter("potatoes", 2, Tags{})
				s.Counter("potatoes", 20, Tags{"foo": "bar"})
				s.Counter("potatoes", 200, Tags{"foo": "bar", "global": "tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 2\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes,foo=bar 20\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes(,(global=tag|foo=bar))+ 200\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		},
		{
			description: "3 counters with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Counter("potatoes", 2, Tags{})
				s.Counter("potatoes", 20, Tags{"foo": "bar"})
				s.Counter("potatoes", 200, Tags{"foo": "bar", "global": "tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,global=tag 2\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes(,(global=tag|foo=bar))+ 20\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes(,(global=tag|foo=bar))+ 200\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		},
		{
			description: "concurrent counters",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Counter("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     10,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ ((count|sum),)+10"),
			},
		},
		// gauges only
		{
			description: "3 gauges",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Gauge("potatoes", 1, Tags{})
				s.Gauge("turnips", 10, Tags{"foo": "bar"})
				s.Gauge("carrots", 100, Tags{"foo": "bar", "global": "tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+ last,10"),
			},
		},
		{
			description: "3 gauges with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Gauge("potatoes", 1, Tags{})
				s.Gauge("turnips", 10, Tags{"foo": "bar"})
				s.Gauge("carrots", 100, Tags{"foo": "bar", "global": "tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,global=tag 1\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("turnips(,(global=tag|foo=bar))+ 10\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+ last,10"),
			},
		},
		{
			description: "concurrent gauges",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Gauge("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     10,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ last,+10"),
			},
		},
		// histograms only
		{
			description: "3 histograms",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Histogram("potatoes", 1, Tags{})
				s.Histogram("turnips", 10, Tags{"foo": "bar"})
				s.Histogram("carrots", 100, Tags{"foo": "bar", "global": "tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		},
		{
			description: "3 histograms with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Histogram("potatoes", 1, Tags{})
				s.Histogram("turnips", 10, Tags{"foo": "bar"})
				s.Histogram("carrots", 100, Tags{"foo": "bar", "global":"tag"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,global=tag 1\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("turnips(,(global=tag|foo=bar))+ 10\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		},
		{
			description: "concurrent histograms",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Histogram("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     11,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ ((avg|count|p90),)+10"),
			},
		},
		// custom metrics only
		{
			description: "3 custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put("potatoes", 1, Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("turnips", 10, Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("carrots", 100, Tags{"foo": "bar", "global":"tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+"),
			},
		},
		{
			description: "3 custom metrics with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put("potatoes", 1, Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("turnips", 10, Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("carrots", 100, Tags{"foo": "bar", "global":"tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,global=tag 1\\.0+ [0-9]+"),
				regexp.MustCompile("turnips(,(global=tag|foo=bar))+ 10\\.0+ [0-9]+"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+ [0-9]+"),
			},
		},
		{
			description: "concurrent custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Put("potatoes", 1, Tags{"worker": workerId}, time.Now().Unix(), Aggregations{}, Freq10s)
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     8,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+"),
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			cancelPeriodicFlush := s.statful.StartFlushInterval(25 * time.Millisecond)
			go s.metricsProducer(&s.statful)

			totalMetricsSent := 0
			totalFlushes := 0

		metricsReceiver:
			for {
				select {
				case d := <-metricsData:
					totalFlushes++
					totalMetricsSent += len(regexp.MustCompile("\n").FindAllSubmatchIndex(d, -1))
					for _, r := range s.metricsSent {
						if !r.Match(d) {
							t.Error("flushed data not what was expected: \n\texpected: \"", r.String(), "\"\n\tactual", string(d))
						}
					}
				case <-time.After(500 * time.Millisecond):
					break metricsReceiver
				}
			}
			cancelPeriodicFlush()

			if s.totalMetricsSent != totalMetricsSent {
				t.Error("Different number of metrics sent: expected ", s.totalMetricsSent, "got", totalMetricsSent)
			}

			if s.totalFlushes != totalFlushes {
				t.Error("Different number of flushes: expected ", s.totalFlushes, "got", totalFlushes)
			}
		})
	}
}
