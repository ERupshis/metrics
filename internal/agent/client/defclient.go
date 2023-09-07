package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/retryer"
)

type DefaultClient struct {
	client *http.Client
	log    logger.BaseLogger
}

func CreateDefault(log logger.BaseLogger) BaseClient {
	return &DefaultClient{client: &http.Client{}, log: log}
}

func (c *DefaultClient) PostJSON(ctx context.Context, url string, body []byte) error {
	compressedBody, err := compressor.GzipCompress(body)
	if err != nil {
		return fmt.Errorf("postJSON request: %w", err)
	}

	request := func(context context.Context) error {
		return c.makeRequest(context, http.MethodPost, url, compressedBody)
	}

	err = retryer.RetryCallWithTimeout(ctx, c.log, nil, nil, request)
	if err != nil {
		err = fmt.Errorf("couldn't send post request")
	}
	return err
}

func (c *DefaultClient) makeRequest(ctx context.Context, method string, url string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
