package client

import "context"

// BaseClient implements client interface.
//
//go:generate mockgen -destination=../../../mocks/mock_BaseClient.go -package=mocks github.com/erupshis/metrics/internal/agent/client BaseClient
type BaseClient interface {
	PostJSON(ctx context.Context, url string, body []byte) error
}
