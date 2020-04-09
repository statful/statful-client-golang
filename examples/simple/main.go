package main

import (
	"log"
	"net/http"

	"github.com/statful/statful-client-golang"
)

func main() {
	metrics := statful.Statful{
		Sender: &statful.ProxyMetricsSender{
			Client: &statful.ApiClient{
				Http:     &http.Client{},
				Url:      "https://api.statful.com",
				BasePath: "/tel/v2.0",
				Token:    "12345678-90ab-cdef-1234-567890abcdef",
			},
		},
		GlobalTags: statful.Tags{"client": "golang"},
		DryRun:     true,
	}

	statful.SetDebugLogger(log.Println)
	statful.SetErrorLogger(log.Println)

	metrics.Put(&statful.Metric{
		Name:  "test.demo.metric",
		Value: 100,
		Tags:  statful.Tags{"env": "EU"},
	})
}
