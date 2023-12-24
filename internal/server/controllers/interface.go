package controllers

import (
	"net"
)

type BaseController interface {
	Serve(lis net.Listener) error
	GracefulStop()
}
