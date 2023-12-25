package server

import (
	"context"
	"net"
)

type BaseServer interface {
	Serve(lis net.Listener) error
	GracefulStop(ctx context.Context) error
}
