package main

import (
	"github.com/statful/client-golang"
	"net/http"
)

func mainSimple() {
	metrics := statful.Statful{
		Sender:     &statful.ProxyMetricsSender{
			Client: &statful.ApiClient{
				Http:          &http.Client{},
				Url:           "https://api.statful.com",
				BasePath:      "/tel/v2.0",
				Token:         "12345678-90ab-cdef-1234-567890abcdef",
			},
		},
		GlobalTags: statful.Tags{"client": "golang"},
		DryRun:     false,
	}

	metrics.Put(&statful.Metric{
		Name:  "test.demo.metric",
		Value: 100,
		Tags:  statful.Tags{"env": "EU"},
	})
}
