package client

import (
	"context"
	"fmt"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/go-resty/resty/v2"
)

// RestyClient object.
type RestyClient struct {
	client *resty.Client
	log    logger.BaseLogger
	hash   *hasher.Hasher
}

// CreateResty creates resty http client. Receives logger and hasher in params.
func CreateResty(log logger.BaseLogger, hash *hasher.Hasher) BaseClient {
	return &RestyClient{resty.New(), log, hash}
}

// PostJSON sends data via http post request.
//
// Performs gzip compression and add hash sum for message validation if hashKey is set in hasher.
// Uses retryer to repeat call in case of connection error.
func (c *RestyClient) PostJSON(context context.Context, url string, body []byte) error {
	compressedBody, err := compressor.GzipCompress(body)
	if err != nil {
		return fmt.Errorf("resty postJSON request: %w", err)
	}

	request := c.client.R().
		SetContext(context).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip")

	if c.hash.GetKey() != "" {
		hashValue, errHash := c.hash.HashMsg(body)
		if errHash != nil {
			return fmt.Errorf("resty postJSON request: hasher calculation: %w", errHash)
		}

		request.SetHeader(c.hash.GetHeader(), hashValue)
	}

	_, err = request.SetBody(compressedBody).Post(url)
	return err
}
