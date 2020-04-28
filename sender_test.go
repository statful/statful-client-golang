package statful

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	apiUrl      = "https://api.Sender.com"
	apiBasePath = "/tel/v2.0"
	apiToken    = "12345678-90ab-cdef-1234-567890abcdef"
	udpAddr     = ":2013"
)

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

var successFullRoundTripper = func(t *testing.T, metrics []string) RoundTripFunc {
	return RoundTripFunc(func(req *http.Request) *http.Response {
		verifyRequest(t, req, metrics)
		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewBufferString("{\"code\":\"SUCCESS\"}")),
			Request:    req,
		}
	})
}

func verifyRequest(t *testing.T, req *http.Request, metrics []string) {
	if !strings.HasPrefix(req.URL.String(), apiUrl) {
		t.Errorf("url is different than configured: expected %v got %v", apiUrl, req.URL.String())
	}
	if req.Header.Get("m-api-token") != apiToken {
		t.Errorf("missing api token: expected %v got %v", apiToken, req.Header.Get("m-api-token"))
	}
	if req.Header.Get("content-type") != "text/plain" {
		t.Errorf("content not in expected format: expected %v got %v", "text/plain", req.Header.Get("content-type"))
	}

	body := req.Body
	defer body.Close()

	if req.Header.Get("content-encoding") == "gzip" {
		r, err := gzip.NewReader(body)
		if err != nil {
			t.Errorf("Failed to unzip data: %v", err)
		}
		body = r
	}

	payload, err := ioutil.ReadAll(body)
	if err != nil {
		t.Errorf("Failed to read payload: %v", err)
	}

	for idx, ml := range strings.Split(string(payload), "\n") {
		if ml != metrics[idx] {
			t.Errorf("different metric lines: expected \"%v\" got \"%v\"", metrics[idx], ml)
		}
	}
}

func TestApiClient_PutMetrics_Success(t *testing.T) {
	scenarios := []struct {
		description  string
		url          string
		basePath     string
		apiToken     string
		roundTripper func(*testing.T, []string) RoundTripFunc
		metrics      []string
	}{
		{
			description: "HttpSender send single metric",
			url:         apiUrl,
			basePath:    apiBasePath,
			apiToken:    apiToken,
			metrics: []string{
				"test.demo.metric,Sender=golang,env=test 100 1585161000",
			},
			roundTripper: successFullRoundTripper,
		}, {
			description: "HttpSender send multiple metrics",
			url:         apiUrl,
			basePath:    apiBasePath,
			apiToken:    apiToken,
			metrics: []string{
				"test.demo.metric 50 1585161006",
				"test.demo.metric,Sender=golang,env=test 100 1585161000",
				"test.demo.metric 200 1585161001 count,10",
				"test.demo.metric,Sender=golang,env=test 300 1585161010 count,10",
				"test.demo.metric 400 1585161011 avg,p90,10",
				"test.demo.metric,Sender=golang,env=test 500 1585161100 avg,p90,10",
				"test.demo.metric,Sender=golang,env=test 600 1585161101",
				"test.demo.metric,Sender=golang,env=test 700 1585161110",
				"test.demo.metric,Sender=golang,env=test 800 1585161111",
				"test.demo.metric,Sender=golang,env=test 900 1585161002",
			},
			roundTripper: successFullRoundTripper,
		},
	}

	for _, s := range scenarios {
		api := HttpSender{
			Url:      s.url,
			BasePath: s.basePath,
			Token:    s.apiToken,
			Http: &http.Client{
				Transport: s.roundTripper(t, s.metrics),
				Timeout:   2 * time.Second,
			},
		}

		err := api.Send(bytes.NewBufferString(strings.Join(s.metrics, "\n")))
		if err != nil {
			t.Errorf("Failed to put metrics: %v", err)
		}

		api = HttpSender{
			Url:           s.url,
			BasePath:      s.basePath,
			Token:         s.apiToken,
			NoCompression: true,
			Http: &http.Client{
				Transport: s.roundTripper(t, s.metrics),
				Timeout:   2 * time.Second,
			},
		}

		err = api.Send(bytes.NewBufferString(strings.Join(s.metrics, "\n")))
		if err != nil {
			t.Errorf("Failed to put metrics: %v", err)
		}
	}
}

func TestApiClient_PutMetrics_FailToCompressData(t *testing.T) {}

func TestApiClient_PutMetrics_FailToCreateRequest(t *testing.T) {}

func TestApiClient_PutMetrics_FailToPerformRequest(t *testing.T) {}

func TestApiClient_PutMetrics_FailToReadBody(t *testing.T) {}

func TestApiClient_PutMetrics_FailStatusCodeNot200(t *testing.T) {}

func getUdpPacket(t *testing.T, addr string, request func()) []byte {
	// listen for udp packets
	resAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		t.Fatal("Failed to resolve udp address:", err)
	}
	listener, err := net.ListenUDP("udp", resAddr)
	if err != nil {
		t.Fatal("Failed to listen for udp packets:", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			t.Fatal("Failed to close udp listener:", err)
		}
	}()

	// execute request
	request()

	//read all messages until 0 bytes are read
	message := make([]byte, 1024*32)
	var bufLen int
	for {
		_ = listener.SetReadDeadline(time.Now().Add(time.Millisecond))
		n, _, err := listener.ReadFrom(message[bufLen:])
		if n == 0 {
			break
		} else {
			bufLen += n
		}
		if err != nil {
			t.Fatal("Failed to read udp packets", err, n)
		}
	}

	return message[0:bufLen]
}

func TestUdpClient_PutMetrics(t *testing.T) {
	scenarios := []struct {
		description string
		metrics     []string
	}{
		{
			description: "UdpSender send single metric",
			metrics: []string{
				"test.demo.metric,Sender=golang,env=test 100 1585161000",
			},
		}, {
			description: "UdpSender send multiple metrics",
			metrics: []string{
				"test.demo.metric 50 1585161006",
				"test.demo.metric,Sender=golang,env=test 100 1585161000",
				"test.demo.metric 200 1585161001 count,10",
				"test.demo.metric,Sender=golang,env=test 300 1585161010 count,10",
				"test.demo.metric 400 1585161011 avg,p90,10",
				"test.demo.metric,Sender=golang,env=test 500 1585161100 avg,p90,10",
				"test.demo.metric,Sender=golang,env=test 600 1585161101",
				"test.demo.metric,Sender=golang,env=test 700 1585161110",
				"test.demo.metric,Sender=golang,env=test 800 1585161111",
				"test.demo.metric,Sender=golang,env=test 900 1585161002",
			},
		},
	}

	for _, s := range scenarios {
		t.Run(s.description, func(t *testing.T) {
			udp := UdpSender{
				Address: udpAddr,
				Timeout: 2 * time.Second,
			}

			udpServerPacket := getUdpPacket(t, udpAddr, func() {
				err := udp.Send(bytes.NewBufferString(strings.Join(s.metrics, "\n")))
				if err != nil {
					t.Fatal("Failed to put metrics:", err)
				}
			})

			for idx, ml := range strings.Split(string(udpServerPacket), "\n") {
				if ml != s.metrics[idx] {
					t.Errorf("different metric lines: expected \"%v\" got \"%v\"", s.metrics[idx], ml)
				}
			}
		})
	}
}
