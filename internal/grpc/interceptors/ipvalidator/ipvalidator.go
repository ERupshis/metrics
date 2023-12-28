package ipvalidator

import (
	"context"
	"fmt"
	"net"

	"github.com/erupshis/metrics/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ValidatorIP struct {
	trustedSubnet *net.IPNet
	IP            string
}

func Create(trustedSubnet *net.IPNet, IP string) *ValidatorIP {
	return &ValidatorIP{
		trustedSubnet: trustedSubnet,
		IP:            IP,
	}
}

func (ip *ValidatorIP) StreamServer(logger logger.BaseLogger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if ip.trustedSubnet == nil {
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			logger.Info("Couldn't extract metadata from context")
		}

		ips := md.Get("X-Real-Ip")
		if len(ips) != 1 {
			logger.Info("Missing X-Real-Ip in metadata")
			return status.Errorf(codes.InvalidArgument, "missing X-Real-Ip in metadata")
		}

		ipNet, err := resolveIP(ips[0])
		if err != nil {
			return status.Errorf(codes.Unavailable, "parse ip error: %v", err)
		}

		if !ip.trustedSubnet.Contains(ipNet) {
			return status.Errorf(codes.Unavailable, "ip resolving failed")
		}

		return handler(srv, ss)
	}
}

func (ip *ValidatorIP) UnaryServer(logger logger.BaseLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if ip.trustedSubnet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Info("Couldn't extract metadata from context")
		}

		ips := md.Get("X-Real-Ip")
		if len(ips) != 1 {
			logger.Info("Missing X-Real-Ip in metadata")
			return nil, status.Errorf(codes.InvalidArgument, "missing X-Real-Ip in metadata")
		}

		ipNet, err := resolveIP(ips[0])
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, "parse ip error: %v", err)
		}

		if !ip.trustedSubnet.Contains(ipNet) {
			return nil, status.Errorf(codes.Unavailable, "ip resolving error: %v", err)
		}

		return handler(ctx, req)
	}
}

func resolveIP(ipStr string) (net.IP, error) {
	if ipStr == "" {
		return nil, fmt.Errorf("missing X-Real-IP header in request")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("failed parse ip from http header")
	}

	return ip, nil
}
