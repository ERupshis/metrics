package grpcserver

import (
	"github.com/erupshis/metrics/internal/grpc/interceptors/logging"
	"github.com/erupshis/metrics/internal/logger"
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

func NewServer(controller *controller.Controller, baseLogger logger.BaseLogger) *Server {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ChainUnaryInterceptor(logging.UnaryServer(baseLogger)))
	opts = append(opts, grpc.ChainStreamInterceptor(logging.StreamServer(baseLogger)))

	s := grpc.NewServer(opts...)
	pb.RegisterMetricsServer(s, controller)

	srv := &Server{
		s,
	}

	return srv
}
