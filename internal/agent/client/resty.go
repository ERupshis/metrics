package client

import (
	"context"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/go-resty/resty/v2"
)

type RestyClient struct {
	client *resty.Client
}

func CreateResty() BaseClient {
	return &RestyClient{resty.New()}
}

func (c *RestyClient) PostJSON(context context.Context, url string, body []byte) error {
	compressedBody, _ := compressor.GzipCompress(body)

	_, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedBody).
		Post(url)

	return err
}
