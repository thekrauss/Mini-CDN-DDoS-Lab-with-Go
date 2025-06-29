// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.21.12
// source: node.proto

package workerpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	WorkerService_RestartService_FullMethodName = "/worker.WorkerService/RestartService"
	WorkerService_StopService_FullMethodName    = "/worker.WorkerService/StopService"
	WorkerService_UpdateConfig_FullMethodName   = "/worker.WorkerService/UpdateConfig"
	WorkerService_SendMetrics_FullMethodName    = "/worker.WorkerService/SendMetrics"
	WorkerService_Hello_FullMethodName          = "/worker.WorkerService/Hello"
	WorkerService_StreamCommands_FullMethodName = "/worker.WorkerService/StreamCommands"
)

// WorkerServiceClient is the client API for WorkerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// API exposée par le worker-node
type WorkerServiceClient interface {
	RestartService(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error)
	StopService(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error)
	UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error)
	// Nouveau : le worker envoie ses métriques
	SendMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error)
	// Nouveau : le worker s'annonce au boot
	Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
	// Nouveau : canal bidirectionnel pour recevoir les commandes du control-plane
	StreamCommands(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[CommandMessage, CommandResult], error)
}

type workerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerServiceClient(cc grpc.ClientConnInterface) WorkerServiceClient {
	return &workerServiceClient{cc}
}

func (c *workerServiceClient) RestartService(ctx context.Context, in *RestartRequest, opts ...grpc.CallOption) (*RestartResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RestartResponse)
	err := c.cc.Invoke(ctx, WorkerService_RestartService_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) StopService(ctx context.Context, in *StopRequest, opts ...grpc.CallOption) (*StopResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(StopResponse)
	err := c.cc.Invoke(ctx, WorkerService_StopService_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) UpdateConfig(ctx context.Context, in *UpdateConfigRequest, opts ...grpc.CallOption) (*UpdateConfigResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UpdateConfigResponse)
	err := c.cc.Invoke(ctx, WorkerService_UpdateConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) SendMetrics(ctx context.Context, in *MetricsRequest, opts ...grpc.CallOption) (*MetricsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(MetricsResponse)
	err := c.cc.Invoke(ctx, WorkerService_SendMetrics_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) Hello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(HelloResponse)
	err := c.cc.Invoke(ctx, WorkerService_Hello_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) StreamCommands(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[CommandMessage, CommandResult], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &WorkerService_ServiceDesc.Streams[0], WorkerService_StreamCommands_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[CommandMessage, CommandResult]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type WorkerService_StreamCommandsClient = grpc.BidiStreamingClient[CommandMessage, CommandResult]

// WorkerServiceServer is the server API for WorkerService service.
// All implementations must embed UnimplementedWorkerServiceServer
// for forward compatibility.
//
// API exposée par le worker-node
type WorkerServiceServer interface {
	RestartService(context.Context, *RestartRequest) (*RestartResponse, error)
	StopService(context.Context, *StopRequest) (*StopResponse, error)
	UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error)
	// Nouveau : le worker envoie ses métriques
	SendMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
	// Nouveau : le worker s'annonce au boot
	Hello(context.Context, *HelloRequest) (*HelloResponse, error)
	// Nouveau : canal bidirectionnel pour recevoir les commandes du control-plane
	StreamCommands(grpc.BidiStreamingServer[CommandMessage, CommandResult]) error
	mustEmbedUnimplementedWorkerServiceServer()
}

// UnimplementedWorkerServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedWorkerServiceServer struct{}

func (UnimplementedWorkerServiceServer) RestartService(context.Context, *RestartRequest) (*RestartResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RestartService not implemented")
}
func (UnimplementedWorkerServiceServer) StopService(context.Context, *StopRequest) (*StopResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopService not implemented")
}
func (UnimplementedWorkerServiceServer) UpdateConfig(context.Context, *UpdateConfigRequest) (*UpdateConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConfig not implemented")
}
func (UnimplementedWorkerServiceServer) SendMetrics(context.Context, *MetricsRequest) (*MetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMetrics not implemented")
}
func (UnimplementedWorkerServiceServer) Hello(context.Context, *HelloRequest) (*HelloResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedWorkerServiceServer) StreamCommands(grpc.BidiStreamingServer[CommandMessage, CommandResult]) error {
	return status.Errorf(codes.Unimplemented, "method StreamCommands not implemented")
}
func (UnimplementedWorkerServiceServer) mustEmbedUnimplementedWorkerServiceServer() {}
func (UnimplementedWorkerServiceServer) testEmbeddedByValue()                       {}

// UnsafeWorkerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerServiceServer will
// result in compilation errors.
type UnsafeWorkerServiceServer interface {
	mustEmbedUnimplementedWorkerServiceServer()
}

func RegisterWorkerServiceServer(s grpc.ServiceRegistrar, srv WorkerServiceServer) {
	// If the following call pancis, it indicates UnimplementedWorkerServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&WorkerService_ServiceDesc, srv)
}

func _WorkerService_RestartService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RestartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).RestartService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_RestartService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).RestartService(ctx, req.(*RestartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_StopService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).StopService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_StopService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).StopService(ctx, req.(*StopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_UpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).UpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_UpdateConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).UpdateConfig(ctx, req.(*UpdateConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_SendMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).SendMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_SendMetrics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).SendMetrics(ctx, req.(*MetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkerService_Hello_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).Hello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_StreamCommands_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(WorkerServiceServer).StreamCommands(&grpc.GenericServerStream[CommandMessage, CommandResult]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type WorkerService_StreamCommandsServer = grpc.BidiStreamingServer[CommandMessage, CommandResult]

// WorkerService_ServiceDesc is the grpc.ServiceDesc for WorkerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "worker.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RestartService",
			Handler:    _WorkerService_RestartService_Handler,
		},
		{
			MethodName: "StopService",
			Handler:    _WorkerService_StopService_Handler,
		},
		{
			MethodName: "UpdateConfig",
			Handler:    _WorkerService_UpdateConfig_Handler,
		},
		{
			MethodName: "SendMetrics",
			Handler:    _WorkerService_SendMetrics_Handler,
		},
		{
			MethodName: "Hello",
			Handler:    _WorkerService_Hello_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamCommands",
			Handler:       _WorkerService_StreamCommands_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "node.proto",
}
