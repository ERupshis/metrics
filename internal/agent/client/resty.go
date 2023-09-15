package client

import (
	"context"
	"fmt"
	"github.com/erupshis/metrics/internal/hasher"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/go-resty/resty/v2"
)

type RestyClient struct {
	client *resty.Client
}

func CreateResty() BaseClient {
	return &RestyClient{resty.New()}
}

func (c *RestyClient) PostJSON(context context.Context, url string, body []byte, hashKey string) error {
	compressedBody, err := compressor.GzipCompress(body)
	if err != nil {
		return fmt.Errorf("resty postJSON request: %w", err)
	}

	request := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip")

	if hashKey != "" {
		hashValue, err := hasher.HashMsg(hasher.SHA256, body, hashKey)
		if err != nil {
			return fmt.Errorf("resty postJSON request: hash calculation: %w", err)
		}

		request.SetHeader("HashSHA256", hashValue)
	}

	_, err = request.SetBody(compressedBody).Post(url)
	return err
}
