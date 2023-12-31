// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.1
// source: metrics.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MetricsClient is the client API for Metrics service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MetricsClient interface {
	Updates(ctx context.Context, opts ...grpc.CallOption) (Metrics_UpdatesClient, error)
	Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Value(ctx context.Context, in *ValueRequest, opts ...grpc.CallOption) (*ValueResponse, error)
	Values(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (Metrics_ValuesClient, error)
	CheckStorage(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*CheckStorageResponse, error)
}

type metricsClient struct {
	cc grpc.ClientConnInterface
}

func NewMetricsClient(cc grpc.ClientConnInterface) MetricsClient {
	return &metricsClient{cc}
}

func (c *metricsClient) Updates(ctx context.Context, opts ...grpc.CallOption) (Metrics_UpdatesClient, error) {
	stream, err := c.cc.NewStream(ctx, &Metrics_ServiceDesc.Streams[0], "/proto_metrics.Metrics/Updates", opts...)
	if err != nil {
		return nil, err
	}
	x := &metricsUpdatesClient{stream}
	return x, nil
}

type Metrics_UpdatesClient interface {
	Send(*UpdatesRequest) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type metricsUpdatesClient struct {
	grpc.ClientStream
}

func (x *metricsUpdatesClient) Send(m *UpdatesRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *metricsUpdatesClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *metricsClient) Update(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto_metrics.Metrics/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) Value(ctx context.Context, in *ValueRequest, opts ...grpc.CallOption) (*ValueResponse, error) {
	out := new(ValueResponse)
	err := c.cc.Invoke(ctx, "/proto_metrics.Metrics/Value", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *metricsClient) Values(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (Metrics_ValuesClient, error) {
	stream, err := c.cc.NewStream(ctx, &Metrics_ServiceDesc.Streams[1], "/proto_metrics.Metrics/Values", opts...)
	if err != nil {
		return nil, err
	}
	x := &metricsValuesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Metrics_ValuesClient interface {
	Recv() (*ValuesResponse, error)
	grpc.ClientStream
}

type metricsValuesClient struct {
	grpc.ClientStream
}

func (x *metricsValuesClient) Recv() (*ValuesResponse, error) {
	m := new(ValuesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *metricsClient) CheckStorage(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*CheckStorageResponse, error) {
	out := new(CheckStorageResponse)
	err := c.cc.Invoke(ctx, "/proto_metrics.Metrics/CheckStorage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MetricsServer is the server API for Metrics service.
// All implementations must embed UnimplementedMetricsServer
// for forward compatibility
type MetricsServer interface {
	Updates(Metrics_UpdatesServer) error
	Update(context.Context, *UpdateRequest) (*emptypb.Empty, error)
	Value(context.Context, *ValueRequest) (*ValueResponse, error)
	Values(*emptypb.Empty, Metrics_ValuesServer) error
	CheckStorage(context.Context, *emptypb.Empty) (*CheckStorageResponse, error)
	mustEmbedUnimplementedMetricsServer()
}

// UnimplementedMetricsServer must be embedded to have forward compatible implementations.
type UnimplementedMetricsServer struct {
}

func (UnimplementedMetricsServer) Updates(Metrics_UpdatesServer) error {
	return status.Errorf(codes.Unimplemented, "method Updates not implemented")
}
func (UnimplementedMetricsServer) Update(context.Context, *UpdateRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedMetricsServer) Value(context.Context, *ValueRequest) (*ValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Value not implemented")
}
func (UnimplementedMetricsServer) Values(*emptypb.Empty, Metrics_ValuesServer) error {
	return status.Errorf(codes.Unimplemented, "method Values not implemented")
}
func (UnimplementedMetricsServer) CheckStorage(context.Context, *emptypb.Empty) (*CheckStorageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckStorage not implemented")
}
func (UnimplementedMetricsServer) mustEmbedUnimplementedMetricsServer() {}

// UnsafeMetricsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MetricsServer will
// result in compilation errors.
type UnsafeMetricsServer interface {
	mustEmbedUnimplementedMetricsServer()
}

func RegisterMetricsServer(s grpc.ServiceRegistrar, srv MetricsServer) {
	s.RegisterService(&Metrics_ServiceDesc, srv)
}

func _Metrics_Updates_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MetricsServer).Updates(&metricsUpdatesServer{stream})
}

type Metrics_UpdatesServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*UpdatesRequest, error)
	grpc.ServerStream
}

type metricsUpdatesServer struct {
	grpc.ServerStream
}

func (x *metricsUpdatesServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *metricsUpdatesServer) Recv() (*UpdatesRequest, error) {
	m := new(UpdatesRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Metrics_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto_metrics.Metrics/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).Update(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_Value_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).Value(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto_metrics.Metrics/Value",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).Value(ctx, req.(*ValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Metrics_Values_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(MetricsServer).Values(m, &metricsValuesServer{stream})
}

type Metrics_ValuesServer interface {
	Send(*ValuesResponse) error
	grpc.ServerStream
}

type metricsValuesServer struct {
	grpc.ServerStream
}

func (x *metricsValuesServer) Send(m *ValuesResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Metrics_CheckStorage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MetricsServer).CheckStorage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto_metrics.Metrics/CheckStorage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MetricsServer).CheckStorage(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Metrics_ServiceDesc is the grpc.ServiceDesc for Metrics service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Metrics_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto_metrics.Metrics",
	HandlerType: (*MetricsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Update",
			Handler:    _Metrics_Update_Handler,
		},
		{
			MethodName: "Value",
			Handler:    _Metrics_Value_Handler,
		},
		{
			MethodName: "CheckStorage",
			Handler:    _Metrics_CheckStorage_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Updates",
			Handler:       _Metrics_Updates_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "Values",
			Handler:       _Metrics_Values_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "metrics.proto",
}
