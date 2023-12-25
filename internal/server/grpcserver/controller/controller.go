package controller

import (
	"context"
	"io"

	"github.com/erupshis/metrics/internal/grpc/utils"
	"github.com/erupshis/metrics/internal/networkmsg"
	"github.com/erupshis/metrics/internal/server/memstorage"
	"github.com/erupshis/metrics/internal/server/memstorage/data"
	"github.com/erupshis/metrics/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Controller struct {
	pb.UnimplementedMetricsServer

	storage *memstorage.MemStorage
}

func New(memStorage *memstorage.MemStorage) *Controller {
	return &Controller{
		storage: memStorage,
	}
}

func (s *Controller) Updates(stream pb.Metrics_UpdatesServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&emptypb.Empty{})
		} else if err != nil {
			return err
		}

		s.storage.AddMetricMessageInStorage(utils.ConvertGrpcFormatToMetric(in.Metric))
	}
}

func (s *Controller) Update(ctx context.Context, in *pb.UpdateRequest) (*emptypb.Empty, error) {
	s.storage.AddMetricMessageInStorage(utils.ConvertGrpcFormatToMetric(in.Metric))
	return &emptypb.Empty{}, nil
}

func (s *Controller) Value(ctx context.Context, in *pb.ValueRequest) (*pb.ValueResponse, error) {
	metric := utils.ConvertGrpcFormatToMetric(in.Metric)
	if metric == nil {
		return nil, status.Errorf(codes.InvalidArgument, "couldn't convert incoming metric")
	}

	switch metric.MType {
	case data.GaugeType:
		value, err := s.storage.GetGauge(metric.ID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "metric not found: %v", err)
		}
		metric.Value = &value
	case data.CounterType:
		value, err := s.storage.GetCounter(metric.ID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "metric not found: %v", err)
		}
		metric.Delta = &value
	}

	return &pb.ValueResponse{
		Metric: utils.ConvertMetricToGrpcFormat(metric),
	}, nil
}

func (s *Controller) Values(_ *emptypb.Empty, stream pb.Metrics_ValuesServer) error {
	for key, val := range s.storage.GetAllGauges() {
		metric := networkmsg.CreateGaugeMetrics(key, val.(float64))
		err := stream.Send(&pb.ValuesResponse{
			Metric: utils.ConvertMetricToGrpcFormat(&metric),
		})

		if err != nil {
			return status.Errorf(codes.Unknown, "sending metric issues")
		}
	}

	for key, val := range s.storage.GetAllCounters() {
		metric := networkmsg.CreateCounterMetrics(key, val.(int64))
		err := stream.Send(&pb.ValuesResponse{
			Metric: utils.ConvertMetricToGrpcFormat(&metric),
		})

		if err != nil {
			return status.Errorf(codes.Unknown, "sending metric issues")
		}
	}

	return nil
}
