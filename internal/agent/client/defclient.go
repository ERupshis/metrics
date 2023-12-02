package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/retryer"
)

// DefaultClient object.
type DefaultClient struct {
	client *http.Client
	log    logger.BaseLogger
	hash   *hasher.Hasher
}

// CreateDefault creates default http client. Receives logger and hasher in params.
func CreateDefault(log logger.BaseLogger, hash *hasher.Hasher) BaseClient {
	return &DefaultClient{client: &http.Client{}, log: log, hash: hash}
}

// PostJSON sends data via http post request.
//
// Performs gzip compression and add hash sum for message validation if hashKey is set in hasher.
// Uses retryer to repeat call in case of connection error.
func (c *DefaultClient) PostJSON(ctx context.Context, url string, body []byte) error {
	compressedBody, err := compressor.GzipCompress(body)
	if err != nil {
		return fmt.Errorf("defclient postJSON request: %w", err)
	}

	var hashValue string
	if c.hash.GetKey() != "" {
		hashValue, err = c.hash.HashMsg(compressedBody)
		if err != nil {
			return fmt.Errorf("defclient postJSON request: hasher calculation: %w", err)
		}
	}

	request := func(context context.Context) error {
		return c.makeRequest(context, http.MethodPost, url, compressedBody, hashValue)
	}

	err = retryer.RetryCallWithTimeout(ctx, c.log, nil, nil, request)
	if err != nil {
		err = fmt.Errorf("couldn't send post request")
	}
	return err
}

func (c *DefaultClient) makeRequest(ctx context.Context, method string, url string, data []byte, hashValue string) error {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	if hashValue != "" {
		req.Header.Set(c.hash.GetHeader(), hashValue)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			c.log.Info("close response body: %v", err)
		}
	}()

	return nil
}
