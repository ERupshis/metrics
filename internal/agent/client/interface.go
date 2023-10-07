package client

import "context"

type BaseClient interface {
	PostJSON(ctx context.Context, url string, body []byte) error
}
