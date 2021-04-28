package statful

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"
)

type funcSender func(...interface{}) error

func (f funcSender) Send(data io.Reader) error {
	if all, err := ioutil.ReadAll(data); err != nil {
		return err
	} else {
		return f(string(all))
	}
}

func (f funcSender) SendAggregated(data io.Reader, agg Aggregation, frequency AggregationFrequency) error {
	if all, err := ioutil.ReadAll(data); err != nil {
		return err
	} else {
		return f(string(all), agg, frequency)
	}
}

type fmtLogger func(...interface{}) (int, error)

func (f fmtLogger) Println(v ...interface{}) {
	f(v...)
}

func ExampleSimple() {
	metrics := New(Configuration{
		FlushSize: 10,
		Logger:    fmtLogger(fmt.Println),
		Tags:      Tags{"client": "golang"},
		DryRun:    true,
	})

	metrics.Put("test.demo.metric", 100, Tags{}, 0, Aggregations{}, Freq10s)
	metrics.Flush()
	// Output: Dry metric: test.demo.metric,client=golang 100.000000 0
}

func ExampleHttpServer() {
	client := New(Configuration{
		DryRun:        false,
		Tags:          Tags{"client": "golang"},
		FlushSize:     50,
		FlushInterval: 10 * time.Second,
		Logger:        log.New(os.Stderr, "", log.LstdFlags),
		Sender: &HttpSender{
			Http:  &http.Client{},
			Url:   "https://api.Sender.com",
			Token: "12345678-90ab-cdef-1234-567890abcdef",
		},
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		client.Counter("http_requests_total", 1, Tags{"status_code": "200", "uri": r.URL.String()})
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Addr: ":8080"}

	srv.Shutdown(context.TODO())
	client.StopFlushInterval()
}

type ChannelSender struct {
	data chan<- []byte
}

func (c *ChannelSender) Send(data io.Reader) error {
	all, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	c.data <- all
	return nil
}

func (c *ChannelSender) SendAggregated(data io.Reader, agg Aggregation, freq AggregationFrequency) error {
	all, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	c.data <- all
	return nil
}

func TestStatfulSDK(t *testing.T) {
	metricsData := make(chan []byte, 1)

	statfulWithoutGlobalTags := New(Configuration{
		FlushSize: 10,
		Logger:    log.New(os.Stderr, "", log.LstdFlags),
		Sender: &ChannelSender{
			data: metricsData,
		},
	})

	statfulWithGlobalTags := New(Configuration{
		FlushSize: 10,
		Logger:    log.New(os.Stderr, "", log.LstdFlags),
		Sender: &ChannelSender{
			data: metricsData,
		},
		Tags: Tags{"global": "tag"},
	})
	//cancelPeriodicFlush := statfulMetrics.StartFlushInterval(1 * time.Second)

	scenarios := []struct {
		description      string
		statful          *Client
		metricsProducer  func(s *Client)
		totalFlushes     int
		totalMetricsSent int
		metricsSent      []*regexp.Regexp
	}{
		// counters only
		{
			description: "3 Counters with tags",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				s.Counter("potatoes", 2, Tags{})
				s.Counter("potatoes", 20, Tags{"foo": "bar"})
				s.Counter("potatoes", 200, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				s.Counter("potatoes", 2, Tags{})
				s.Counter("potatoes", 20, Tags{"foo": "bar"})
				s.Counter("potatoes", 200, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Client, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Counter("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
				s.Flush()
			},
			totalFlushes:     20,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ ((count|sum),)+10"),
			},
		},
		// gauges only
		{
			description: "3 gauges",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				s.Gauge("potatoes", 1, Tags{})
				s.Gauge("turnips", 10, Tags{"foo": "bar"})
				s.Gauge("carrots", 100, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				s.Gauge("potatoes", 1, Tags{})
				s.Gauge("turnips", 10, Tags{"foo": "bar"})
				s.Gauge("carrots", 100, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Client, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Gauge("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
				s.Flush()
			},
			totalFlushes:     20,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ last,+10"),
			},
		},
		// histograms only
		{
			description: "3 histograms",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				s.Timer("potatoes", 1, Tags{})
				s.Timer("turnips", 10, Tags{"foo": "bar"})
				s.Timer("carrots", 100, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				s.Timer("potatoes", 1, Tags{})
				s.Timer("turnips", 10, Tags{"foo": "bar"})
				s.Timer("carrots", 100, Tags{"foo": "bar", "global": "tag"})
				s.Flush()
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
			metricsProducer: func(s *Client) {
				wg := sync.WaitGroup{}
				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Client, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Timer("potatoes", float64(1), Tags{"worker": workerId})
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
				s.Flush()
			},
			totalFlushes:     20,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+ ((avg|count|p90),)+10"),
			},
		},
		// custom metrics only
		{
			description: "3 custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				s.Put("potatoes", 1, Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("turnips", 10, Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("carrots", 100, Tags{"foo": "bar", "global": "tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Flush()
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
			metricsProducer: func(s *Client) {
				s.Put("potatoes", 1, Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("turnips", 10, Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Put("carrots", 100, Tags{"foo": "bar", "global": "tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Flush()
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
			metricsProducer: func(s *Client) {
				wg := sync.WaitGroup{}

				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Client, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.Put("potatoes", 1, Tags{"worker": workerId}, time.Now().Unix(), Aggregations{}, Freq10s)
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
				s.Flush()
			},
			totalFlushes:     20,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+ [0-9]+"),
			},
		},
		// PutWithUser metrics only
		{
			description: "3 custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				s.PutWithUser("potatoes", 1, "user", Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.PutWithUser("turnips", 10, "user", Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.PutWithUser("carrots", 100, "user", Tags{"foo": "bar", "global": "tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Flush()
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes 1\\.0+,user [0-9]+"),
				regexp.MustCompile("turnips,foo=bar 10\\.0+,user [0-9]+"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+,user [0-9]+"),
			},
		},
		{
			description: "3 custom metrics with global tags",
			statful:     statfulWithGlobalTags,
			metricsProducer: func(s *Client) {
				s.PutWithUser("potatoes", 1, "user", Tags{}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.PutWithUser("turnips", 10, "user", Tags{"foo": "bar"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.PutWithUser("carrots", 100, "user", Tags{"foo": "bar", "global": "tag"}, time.Now().Unix(), Aggregations{}, Freq10s)
				s.Flush()
			},
			totalFlushes:     1,
			totalMetricsSent: 3,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,global=tag 1\\.0+,user [0-9]+"),
				regexp.MustCompile("turnips(,(global=tag|foo=bar))+ 10\\.0+,user [0-9]+"),
				regexp.MustCompile("carrots(,(global=tag|foo=bar))+ 100\\.0+,user [0-9]+"),
			},
		},
		{
			description: "concurrent custom metrics",
			statful:     statfulWithoutGlobalTags,
			metricsProducer: func(s *Client) {
				wg := sync.WaitGroup{}

				for i := 0; i < 10; i++ {
					wg.Add(1)
					go func(metrics *Client, workerId string, wg *sync.WaitGroup) {
						for i := 0; i < 20; i++ {
							metrics.PutWithUser("potatoes", 1, "user", Tags{"worker": workerId}, time.Now().Unix(), Aggregations{}, Freq10s)
						}
						wg.Done()
					}(s, strconv.Itoa(i), &wg)
				}

				wg.Wait()
				s.Flush()
			},
			totalFlushes:     20,
			totalMetricsSent: 200,
			metricsSent: []*regexp.Regexp{
				regexp.MustCompile("potatoes,worker=\\d+ 1\\.?[0-9]+,user [0-9]+"),
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			go s.metricsProducer(s.statful)

			totalMetricsSent := 0
			totalFlushes := 0

		metricsReceiver:
			for {
				select {
				case d := <-metricsData:
					totalFlushes++
					totalMetricsSent += len(bytes.Split(d, []byte("\n")))

					for _, r := range s.metricsSent {
						if !r.Match(d) {
							t.Error("flushed data not what was expected: \n\texpected: \"", r.String(), "\"\n\tactual", string(d))
						}
					}
				case <-time.After(100 * time.Millisecond):
					break metricsReceiver
				}
			}

			if s.totalMetricsSent != totalMetricsSent {
				t.Error("Different number of metrics sent: expected ", s.totalMetricsSent, "got", totalMetricsSent)
			}

			if s.totalFlushes != totalFlushes {
				t.Error("Different number of flushes: expected ", s.totalFlushes, "got", totalFlushes)
			}
		})
	}
}
