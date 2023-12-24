package interceptors

import (
	"context"

	"github.com/erupshis/metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptorStream(logger logger.BaseLogger) func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.Info("Stream method %s called", info.FullMethod)

		err := handler(srv, ss)
		if err != nil {
			logger.Info("grpc stream: %v", err)
		}

		return err
	}
}

func LoggingInterceptorUnary(logger logger.BaseLogger) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger.Info("Unary method %s called", info.FullMethod)

		resp, err := handler(ctx, req)

		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				logger.Info("Unary method '%s' completed with error '%v', status: %s", info.FullMethod, err, st.Code().String())
			} else {
				logger.Info("Unary method '%s' completed with error '%v', status: unknown", info.FullMethod, err)
			}
		} else {
			logger.Info("Unary method '%s' completed, status: %s", info.FullMethod, codes.OK.String())
		}

		return resp, err
	}
}
