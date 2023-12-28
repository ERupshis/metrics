package client

import (
	"context"

	"github.com/erupshis/metrics/internal/networkmsg"
)

// BaseClient implements client interface.
//
//go:generate mockgen -destination=../../../mocks/mock_BaseClient.go -package=mocks github.com/erupshis/metrics/internal/agent/client BaseClient
type BaseClient interface {
	Post(ctx context.Context, metrics []networkmsg.Metric) error
}
