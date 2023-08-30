package client

import (
	"bytes"
	"net/http"

	"github.com/erupshis/metrics/internal/compressor"
)

type DefaultClient struct {
	client *http.Client
}

func CreateDefault() BaseClient {
	return &DefaultClient{&http.Client{}}
}

func (c *DefaultClient) PostJSON(url string, body []byte) error {
	compressedBody, _ := compressor.GzipCompress(body)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	_, err = c.client.Do(req)
	return err
}
