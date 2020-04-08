package main

import (
	"bytes"
	"github.com/statful/client-golang"
	"log"
	"net/http"
)

func mainHttpServer() {
	metrics := statful.Statful{
		Sender: &statful.BufferedMetricsSender{
			Client: &statful.ApiClient{
				Http:     &http.Client{},
				Url:      "https://api.statful.com",
				BasePath: "/tel/v2.0/",
				Token:    "12345678-09ab-cdef-1234-567890abcdef",
			},
			FlushSize: 1000,
			Buf:       bytes.Buffer{},
		},
		GlobalTags: statful.Tags{"client": "golang"},
		DryRun:     false,
	}
	statful.SetDebugLogger(log.Println)
	statful.SetErrorLogger(log.Println)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		metrics.CounterWithTags("http_requests_total", 1, statful.Tags{"status_code": "200", "uri": r.URL.String()})
		w.WriteHeader(http.StatusOK)
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
