package controller

import (
	"context"

	"github.com/erupshis/metrics/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Controller struct {
	pb.UnimplementedMetricsServer
}

func New() *Controller {
	return &Controller{}
}

func (s *Controller) Updates(stream pb.Metrics_UpdatesServer) error {
	return status.Errorf(codes.Unimplemented, "method Updates not implemented")
}
func (s *Controller) Update(ctx context.Context, in *pb.UpdateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (s *Controller) Value(ctx context.Context, in *pb.ValueRequest) (*pb.ValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Value not implemented")
}
func (s *Controller) Values(_ *emptypb.Empty, stream pb.Metrics_ValuesServer) error {
	return status.Errorf(codes.Unimplemented, "method Values not implemented")
}
