package client

import (
	"context"
	"fmt"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/go-resty/resty/v2"
)

type RestyClient struct {
	client *resty.Client
	log    logger.BaseLogger
	hash   *hasher.Hasher
}

func CreateResty(log logger.BaseLogger, hash *hasher.Hasher) BaseClient {
	return &RestyClient{resty.New(), log, hash}
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
		hashValue, err := c.hash.HashMsg(body, hashKey)
		if err != nil {
			return fmt.Errorf("resty postJSON request: hasher calculation: %w", err)
		}

		request.SetHeader(c.hash.GetHeader(), hashValue)
	}

	_, err = request.SetBody(compressedBody).Post(url)
	return err
}
