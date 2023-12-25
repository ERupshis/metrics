package grpcserver

import (
	"github.com/erupshis/metrics/internal/server"
	"github.com/erupshis/metrics/internal/server/grpcserver/controller"
	"github.com/erupshis/metrics/pb"
	"google.golang.org/grpc"
)

var (
	_ server.BaseServer = (*Server)(nil)
)

type Server struct {
	*grpc.Server
}

func NewServer(controller *controller.Controller, options ...grpc.ServerOption) *Server {
	var opts []grpc.ServerOption
	for _, option := range options {
		opts = append(opts, option)
	}

	s := grpc.NewServer(opts...)
	pb.RegisterMetricsServer(s, controller)

	srv := &Server{
		s,
	}

	return srv
}
