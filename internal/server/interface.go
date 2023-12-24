package server

import (
	"net"
)

type BaseServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
}
