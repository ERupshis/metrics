package grpcserver

import (
	"github.com/erupshis/metrics/internal/grpc/interceptors"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/controllers"
	"github.com/erupshis/metrics/internal/server/grpcserver/controller"
	"github.com/erupshis/metrics/pb"
	"google.golang.org/grpc"
)

var (
	_ controllers.BaseController = (*Server)(nil)
)

type Server struct {
	*grpc.Server
}

func NewServer(controller *controller.Controller, baseLogger logger.BaseLogger) *Server {
	var opts []grpc.ServerOption
	opts = append(opts, grpc.UnaryInterceptor(interceptors.LoggingInterceptorUnary(baseLogger)))
	opts = append(opts, grpc.StreamInterceptor(interceptors.LoggingInterceptorStream(baseLogger)))

	s := grpc.NewServer(opts...)
	pb.RegisterMetricsServer(s, controller)

	srv := &Server{
		s,
	}

	return srv
}
