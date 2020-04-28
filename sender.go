package statful

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Sender interface {
	Send(data io.Reader) error
	SendAggregated(data io.Reader, agg Aggregation, frequency AggregationFrequency) error
}

const (
	ep_metrics            = "/tel/v2.0/metrics"
	ep_metrics_aggregated = "/tel/v2.0/aggregation/:agg/frequency/:freq"
)

type HttpSender struct {
	Http          *http.Client
	Url           string
	BasePath      string
	Token         string
	NoCompression bool
}

func (h *HttpSender) Send(data io.Reader) error {
	p := h.Url + h.BasePath + ep_metrics

	return h.do(http.MethodPut, p, data)
}

func (h *HttpSender) SendAggregated(data io.Reader, agg Aggregation, freq AggregationFrequency) error {
	p := h.Url + h.BasePath + ep_metrics_aggregated
	p = strings.Replace(p, ":agg", string(agg), -1)
	p = strings.Replace(p, ":freq", strconv.Itoa(int(freq)), -1)

	return h.do(http.MethodPut, p, data)
}

func (h *HttpSender) do(method string, url string, data io.Reader) error {
	headers := http.Header{}

	if !h.NoCompression {
		compressed, err := gzipData(data)
		if err != nil {
			return err
		}
		data = compressed
		headers.Set("Content-Encoding", "gzip")
	}

	headers.Set("M-API-Token", h.Token)
	headers.Set("Content-Type", "text/plain")

	req, err := http.NewRequest(method, url, data)
	req.Header = headers
	if err != nil {
		return err
	}

	resp, err := h.Http.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("Http request failed with %v %v", resp.StatusCode, string(body)))
	}

	return nil
}

func gzipData(reader io.Reader) (io.Reader, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	defer gw.Close()

	if _, err := io.Copy(gw, reader); err != nil {
		return nil, err
	}

	if err := gw.Flush(); err != nil {
		return nil, err
	}

	if err := gw.Close(); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

type UdpSender struct {
	Address string
	Timeout time.Duration
}

func (u *UdpSender) Send(reader io.Reader) error {
	conn, err := net.Dial("udp", u.Address)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, reader)
	if err != nil {
		return err
	}

	return nil
}

func (u *UdpSender) SendAggregated(io.Reader, Aggregation, AggregationFrequency) error {
	return errors.New("UNSUPPORTED_OPERATION")
}
