package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/statful/statful-client-golang"
)

func main() {
	metrics := statful.Statful{
		Sender: &statful.BufferedMetricsSender{
			Client: &statful.ApiClient{
				Http:     &http.Client{},
				Url:      "https://api.statful.com",
				BasePath: "/tel/v2.0",
				Token:    "12345678-90ab-cdef-1234-567890abcdef",
			},
			FlushSize: 1000,
			Buf:       bytes.Buffer{},
		},
		GlobalTags: statful.Tags{"client": "golang"},
		DryRun:     true,
	}
	statful.SetDebugLogger(log.Println)
	statful.SetErrorLogger(log.Println)
	cancelFlushInterval := metrics.StartFlushInterval(10 * time.Second)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		metrics.CounterWithTags("http_requests_total", 1, statful.Tags{"status_code": "200", "uri": r.URL.String()})
		w.WriteHeader(http.StatusOK)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
	cancelFlushInterval()
}
