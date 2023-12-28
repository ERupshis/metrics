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
	info string
	port string
}

func NewServer(controller *controller.Controller, info string, options ...grpc.ServerOption) *Server {
	s := grpc.NewServer(options...)
	pb.RegisterMetricsServer(s, controller)

	srv := &Server{
		Server: s,
		info:   info,
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
func (s *Server) GetInfo() string {
	return s.info
}

func (s *Server) Host(host string) {
	s.port = host
}
func (s *Server) GetHost() string {
	return s.port
}
