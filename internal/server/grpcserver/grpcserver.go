package grpcserver

import (
	"context"
	"net"

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
	s := grpc.NewServer(options...)
	pb.RegisterMetricsServer(s, controller)

	srv := &Server{
		s,
	}

	return srv
}

func (s *Server) Serve(lis net.Listener) error {
	return s.Server.Serve(lis)
}

func (s *Server) GracefulStop(_ context.Context) error {
	s.Server.GracefulStop()
	return nil
}
