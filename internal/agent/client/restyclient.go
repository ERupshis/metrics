package client

import (
	"github.com/erupshis/metrics/internal/compressor"
	"github.com/go-resty/resty/v2"
)

type restyClient struct {
	client *resty.Client
}

func CreateResty() BaseClient {
	return &restyClient{resty.New()}
}

func (c *restyClient) PostJson(url string, body []byte) error {
	compressedBody, _ := compressor.GzipCompress(body)

	_, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedBody).
		Post(url)

	return err
}
