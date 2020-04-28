package statful

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

const (
	ep_metrics            = "/tel/v2.0/metrics"
	ep_metrics_aggregated = "/tel/v2.0/aggregation/:agg/frequency/:freq"
)

type ApiClient struct {
	Http          *http.Client
	Url           string
	BasePath      string
	Token         string
	NoCompression bool
}

func (a *ApiClient) Put(data io.Reader) error {
	p := a.Url + a.BasePath + ep_metrics

	return a.do(http.MethodPut, p, data)
}

func (a *ApiClient) PutAggregated(data io.Reader, agg Aggregation, freq AggregationFrequency) error {
	p := a.Url + a.BasePath + ep_metrics_aggregated
	p = strings.Replace(p, ":agg", string(agg), -1)
	p = strings.Replace(p, ":freq", strconv.Itoa(int(freq)), -1)

	return a.do(http.MethodPut, p, data)
}

func (a *ApiClient) do(method string, url string, data io.Reader) error {
	headers := http.Header{}

	if !a.NoCompression {
		compressed, err := gzipData(data)
		if err != nil {
			return err
		}
		data = compressed
		headers.Set("Content-Encoding", "gzip")
	}

	headers.Set("M-Api-Token", a.Token)
	headers.Set("Content-Type", "text/plain")

	req, err := http.NewRequest(method, url, data)
	req.Header = headers
	if err != nil {
		return err
	}

	resp, err := a.Http.Do(req)
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

