package client

import "context"

type BaseClient interface {
	PostJSON(context context.Context, url string, body []byte) error
}
