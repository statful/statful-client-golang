package statful

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	ep_metrics            = "/tel/v2.0/metrics"
	ep_metrics_aggregated = "/tel/v2.0/aggregation/{agg}/frequency/{freq}"
)

type ApiClient struct {
	Http          *http.Client
	Url           string
	BasePath      string
	Token         string
	NoCompression bool
}

func (a *ApiClient) Send(data io.Reader) error {
	if !a.NoCompression {
		compressed, err := gzipData(data)
		if err != nil {
			errorLog("Failed to compress data:", err)
			return err
		}
		data = compressed
	}

	req, err := http.NewRequest(http.MethodPut, a.Url+a.BasePath+ep_metrics, data)
	if err != nil {
		errorLog("Failed to build request:", err)
		return err
	}

	req.Header.Set("M-Api-Token", a.Token)
	req.Header.Set("Content-Type", "text/plain")
	if !a.NoCompression {
		req.Header.Set("Content-Encoding", "gzip")
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

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		debugLog("sent metrics")
	} else {
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

