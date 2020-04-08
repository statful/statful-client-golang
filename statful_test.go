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
		GlobalTags: Tags{"foo": "bar"},
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
			description: "3 counters",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Counter("potatoes", 1)
				s.Counter("turnips", 10)
				s.Counter("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("turnips 10\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("carrots 100\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		}, {
			description: "3 Counters with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.CounterWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.CounterWithTags("potatoes", 20, Tags{"origin": "india"})
				s.CounterWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,origin=portugal 2\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes,origin=india 20\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes,origin=uk 200\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		}, {
			description: "3 counters with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Counter("potatoes", 1)
				s.Counter("turnips", 10)
				s.Counter("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,foo=bar 1\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("carrots,foo=bar 100\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		}, {
			description: "3 Counters with tags and global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.CounterWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.CounterWithTags("potatoes", 20, Tags{"origin": "india"})
				s.CounterWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes(,(foo=bar|origin=portugal))+ 2\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=india))+ 20\\.0+ [0-9]+ ((count|sum),)+10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=uk))+ 200\\.0+ [0-9]+ ((count|sum),)+10"),
			},
		}, {
			description: "concurrent counters",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.CounterWithTags("potatoes", float64(1), Tags{"worker": workerId})
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
		}, {
			description: "concurrent counters with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.CounterWithTags("turnips", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     12,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("turnips(,(worker=\\d+|foo=bar))+ 1\\.?[0-9]+ [0-9]+ ((count|sum),)+10"),
			},
		},
		// gauges only
		{
			description: "3 gauges",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Gauge("potatoes", 1)
				s.Gauge("turnips", 10)
				s.Gauge("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("turnips 10\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("carrots 100\\.0+ [0-9]+ last,10"),
			},
		}, {
			description: "3 gauges with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.GaugeWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.GaugeWithTags("potatoes", 20, Tags{"origin": "india"})
				s.GaugeWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,origin=portugal 2\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("potatoes,origin=india 20\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("potatoes,origin=uk 200\\.0+ [0-9]+ last,10"),
			},
		}, {
			description: "3 gauges with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Gauge("potatoes", 1)
				s.Gauge("turnips", 10)
				s.Gauge("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,foo=bar 1\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("carrots,foo=bar 100\\.0+ [0-9]+ last,10"),
			},
		}, {
			description: "3 gauges with tags and global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.GaugeWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.GaugeWithTags("potatoes", 20, Tags{"origin": "india"})
				s.GaugeWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes(,(foo=bar|origin=portugal))+ 2\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=india))+ 20\\.0+ [0-9]+ last,10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=uk))+ 200\\.0+ [0-9]+ last,10"),
			},
		}, {
			description: "concurrent gauges",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.GaugeWithTags("potatoes", float64(1), Tags{"worker": workerId})
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
		}, {
			description: "concurrent gauges with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.GaugeWithTags("turnips", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     11,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("turnips(,(worker=\\d+|foo=bar))+ 1\\.?[0-9]+ [0-9]+ last,10"),
			},
		},
		// histograms only
		{
			description: "3 histograms",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Histogram("potatoes", 1)
				s.Histogram("turnips", 10)
				s.Histogram("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("turnips 10\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("carrots 100\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		}, {
			description: "3 histograms with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.HistogramWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.HistogramWithTags("potatoes", 20, Tags{"origin": "india"})
				s.HistogramWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,origin=portugal 2\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("potatoes,origin=india 20\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("potatoes,origin=uk 200\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		}, {
			description: "3 histograms with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Histogram("potatoes", 1)
				s.Histogram("turnips", 10)
				s.Histogram("carrots", 100)
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,foo=bar 1\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("carrots,foo=bar 100\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		}, {
			description: "3 histograms with tags and global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.HistogramWithTags("potatoes", 2, Tags{"origin": "portugal"})
				s.HistogramWithTags("potatoes", 20, Tags{"origin": "india"})
				s.HistogramWithTags("potatoes", 200, Tags{"origin": "uk"})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes(,(foo=bar|origin=portugal))+ 2\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=india))+ 20\\.0+ [0-9]+ ((avg|count|p90),)+10"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=uk))+ 200\\.0+ [0-9]+ ((avg|count|p90),)+10"),
			},
		}, {
			description: "concurrent histograms",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.HistogramWithTags("potatoes", float64(1), Tags{"worker": workerId})
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
		}, {
			description: "concurrent histograms with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.HistogramWithTags("turnips", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     12,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("turnips(,(worker=\\d+|foo=bar))+ 1\\.?[0-9]+ [0-9]+ ((avg|count|p90),)+10"),
			},
		},
		// custom metrics only
		{
			description: "3 custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 1,
					Tags:  Tags{},
				})
				s.Put(&Metric{
					Name:  "turnips",
					Value: 10,
					Tags:  Tags{},
				})
				s.Put(&Metric{
					Name:  "carrots",
					Value: 100,
					Tags:  Tags{},
				})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+ [0-9]+"),
				regexp.MustCompile("turnips 10\\.0+ [0-9]+"),
				regexp.MustCompile("carrots 100\\.0+ [0-9]+"),
			},
		}, {
			description: "3 custom metrics with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 2,
					Tags:  Tags{"origin": "portugal"},
				})
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 20,
					Tags:  Tags{"origin": "india"},
				})
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 200,
					Tags:  Tags{"origin": "uk"},
				})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,origin=portugal 2\\.0+ [0-9]+"),
				regexp.MustCompile("potatoes,origin=india 20\\.0+ [0-9]+"),
				regexp.MustCompile("potatoes,origin=uk 200\\.0+ [0-9]+"),
			},
		}, {
			description: "3 custom metrics with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 1,
					Tags:  Tags{},
				})
				s.Put(&Metric{
					Name:  "turnips",
					Value: 10,
					Tags:  Tags{},
				})
				s.Put(&Metric{
					Name:  "carrots",
					Value: 100,
					Tags:  Tags{},
				})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,foo=bar 1\\.0+ [0-9]+"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+ [0-9]+"),
				regexp.MustCompile("carrots,foo=bar 100\\.0+ [0-9]+"),
			},
		}, {
			description: "3 custom metrics with tags and global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 2,
					Tags:  Tags{"origin": "portugal"},
				})
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 20,
					Tags:  Tags{"origin": "india"},
					Aggs: Aggregations{AggAvg: struct{}{}, AggCount: struct{}{}},
					Freq: Freq30s,
				})
				s.Put(&Metric{
					Name:  "potatoes",
					Value: 200,
					Tags:  Tags{"origin": "uk"},
					Aggs: Aggregations{AggAvg: struct{}{}, AggCount: struct{}{}, AggP99: struct{}{}, AggSum: struct{}{}},
					Freq: Freq60s,
				})
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes(,(foo=bar|origin=portugal))+ 2\\.0+ [0-9]+"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=india))+ 20\\.0+ [0-9]+ ((avg|count),)+30"),
				regexp.MustCompile("potatoes(,(foo=bar|origin=uk))+ 200\\.0+ [0-9]+ ((avg|count|p99|sum),)+60"),
			},
		}, {
			description: "concurrent custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Put(&Metric{Name: "potatoes", Value: float64(1), Tags: Tags{"worker": workerId}})
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
		}, {
			description: "concurrent custom metrics with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Statful) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Statful, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Put(&Metric{Name: "turnips", Value: float64(1), Tags: Tags{"worker": workerId}})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
			},
			totalFlushes:     9,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("turnips(,(worker=\\d+|foo=bar))+ 1\\.?[0-9]+ [0-9]+"),
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
