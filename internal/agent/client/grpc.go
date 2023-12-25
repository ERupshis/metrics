package client

import (
	"context"
	"fmt"

	"github.com/erupshis/metrics/internal/grpc/utils"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	_ BaseClient = (*Grpc)(nil)
)

type Grpc struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn

	IP string
}

func CreateGRPC(address string, IP string, options ...grpc.DialOption) (BaseClient, error) {
	conn, err := grpc.Dial(address, options...)
	if err != nil {
		return nil, fmt.Errorf("create connection to filestorage: %w", err)
	}

	client := pb.NewMetricsClient(conn)

	return &Grpc{
		client: client,
		conn:   conn,
		IP:     IP,
	}, nil
}

func (s *Grpc) Close() error {
	return s.conn.Close()
}

func (s *Grpc) Post(ctx context.Context, metrics []networkmsg.Metric) error {
	md := metadata.Pairs(
		"X-Real_IP", s.IP,
	)

	mdCtx := metadata.NewOutgoingContext(ctx, md)

	if len(metrics) == 1 {
		_, err := s.client.Update(mdCtx, &pb.UpdateRequest{Metric: utils.ConvertMetricToGrpcFormat(&metrics[0])})
		return err
	} else {
		stream, err := s.client.Updates(mdCtx)
		if err != nil {
			return err
		}

		defer func() {
			_ = stream.CloseSend()
		}()

		for _, metric := range metrics {
			metric := metric
			err = stream.Send(&pb.UpdatesRequest{Metric: utils.ConvertMetricToGrpcFormat(&metric)})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
