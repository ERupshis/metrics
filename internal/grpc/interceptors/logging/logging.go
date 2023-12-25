package logging

import (
	"context"

	"github.com/erupshis/metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StreamServer(logger logger.BaseLogger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		logger.Info("Stream method %s called", info.FullMethod)

		err := handler(srv, ss)
		if err != nil {
			logger.Info("grpc stream: %v", err)
		}

		return err
	}
}

func UnaryServer(logger logger.BaseLogger) grpc.UnaryServerInterceptor {
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

func StreamClient(logger logger.BaseLogger) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		logger.Info("Stream method %s called", method)

		s, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			logger.Info("Stream method %s result with err: %v", method, err)
		} else {
			logger.Info("Stream method %s successfully initiated")
		}

		return s, err
	}
}

func UnaryClient(logger logger.BaseLogger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		logger.Info("Unary method %s called", method)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			logger.Info("Unary method %s result with err: %v", method, err)
		} else {
			logger.Info("Unary method %s successfully completed")
		}

		return err
	}
}
