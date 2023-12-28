package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erupshis/metrics/internal/compressor"
	"github.com/erupshis/metrics/internal/hasher"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/go-resty/resty/v2"
)

// RestyClient object.
type RestyClient struct {
	client *resty.Client
	log    logger.BaseLogger
	hash   *hasher.Hasher
	IP     string
	host   string
}

// CreateResty creates resty http client. Receives logger and hasher in params.
func CreateResty(log logger.BaseLogger, hash *hasher.Hasher, IP string, host string) BaseClient {
	return &RestyClient{client: resty.New(), log: log, hash: hash, IP: IP}
}

// PostJSON sends data via http post request.
//
// Performs gzip compression and add hash sum for message validation if hashKey is set in hasher.
// Uses retryer to repeat call in case of connection error.
func (c *RestyClient) Post(context context.Context, metrics []networkmsg.Metric) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	compressedBody, err := compressor.GzipCompress(body)
	if err != nil {
		return fmt.Errorf("resty postJSON request: %w", err)
	}

	request := c.client.R().
		SetContext(context).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("X-Real-IP", c.IP)

	if c.hash.GetKey() != "" {
		hashValue, errHash := c.hash.HashMsg(body)
		if errHash != nil {
			return fmt.Errorf("resty postJSON request: hasher calculation: %w", errHash)
		}

		request.SetHeader(c.hash.GetHeader(), hashValue)
	}

	url := c.host
	if len(metrics) == 1 {
		url += "/update/"
	} else {
		url += "/updates/"
	}

	_, err = request.SetBody(compressedBody).Post(url)
	return err
}
