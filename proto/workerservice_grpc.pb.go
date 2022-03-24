// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: proto/workerservice.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WorkerServiceClient is the client API for WorkerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkerServiceClient interface {
	StartJob(ctx context.Context, in *StartJobRequest, opts ...grpc.CallOption) (*StartJobResponse, error)
	StopJob(ctx context.Context, in *StopJobRequest, opts ...grpc.CallOption) (*StopJobResponse, error)
	GetJobStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error)
	GetOutputStream(ctx context.Context, in *GetStreamRequest, opts ...grpc.CallOption) (WorkerService_GetOutputStreamClient, error)
}

type workerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkerServiceClient(cc grpc.ClientConnInterface) WorkerServiceClient {
	return &workerServiceClient{cc}
}

func (c *workerServiceClient) StartJob(ctx context.Context, in *StartJobRequest, opts ...grpc.CallOption) (*StartJobResponse, error) {
	out := new(StartJobResponse)
	err := c.cc.Invoke(ctx, "/proto.WorkerService/StartJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) StopJob(ctx context.Context, in *StopJobRequest, opts ...grpc.CallOption) (*StopJobResponse, error) {
	out := new(StopJobResponse)
	err := c.cc.Invoke(ctx, "/proto.WorkerService/StopJob", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) GetJobStatus(ctx context.Context, in *GetStatusRequest, opts ...grpc.CallOption) (*GetStatusResponse, error) {
	out := new(GetStatusResponse)
	err := c.cc.Invoke(ctx, "/proto.WorkerService/GetJobStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workerServiceClient) GetOutputStream(ctx context.Context, in *GetStreamRequest, opts ...grpc.CallOption) (WorkerService_GetOutputStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &WorkerService_ServiceDesc.Streams[0], "/proto.WorkerService/GetOutputStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &workerServiceGetOutputStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WorkerService_GetOutputStreamClient interface {
	Recv() (*GetStreamResponse, error)
	grpc.ClientStream
}

type workerServiceGetOutputStreamClient struct {
	grpc.ClientStream
}

func (x *workerServiceGetOutputStreamClient) Recv() (*GetStreamResponse, error) {
	m := new(GetStreamResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// WorkerServiceServer is the server API for WorkerService service.
// All implementations must embed UnimplementedWorkerServiceServer
// for forward compatibility
type WorkerServiceServer interface {
	StartJob(context.Context, *StartJobRequest) (*StartJobResponse, error)
	StopJob(context.Context, *StopJobRequest) (*StopJobResponse, error)
	GetJobStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error)
	GetOutputStream(*GetStreamRequest, WorkerService_GetOutputStreamServer) error
	mustEmbedUnimplementedWorkerServiceServer()
}

// UnimplementedWorkerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWorkerServiceServer struct {
}

func (UnimplementedWorkerServiceServer) StartJob(context.Context, *StartJobRequest) (*StartJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartJob not implemented")
}
func (UnimplementedWorkerServiceServer) StopJob(context.Context, *StopJobRequest) (*StopJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopJob not implemented")
}
func (UnimplementedWorkerServiceServer) GetJobStatus(context.Context, *GetStatusRequest) (*GetStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetJobStatus not implemented")
}
func (UnimplementedWorkerServiceServer) GetOutputStream(*GetStreamRequest, WorkerService_GetOutputStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method GetOutputStream not implemented")
}
func (UnimplementedWorkerServiceServer) mustEmbedUnimplementedWorkerServiceServer() {}

// UnsafeWorkerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkerServiceServer will
// result in compilation errors.
type UnsafeWorkerServiceServer interface {
	mustEmbedUnimplementedWorkerServiceServer()
}

func RegisterWorkerServiceServer(s grpc.ServiceRegistrar, srv WorkerServiceServer) {
	s.RegisterService(&WorkerService_ServiceDesc, srv)
}

func _WorkerService_StartJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).StartJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WorkerService/StartJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).StartJob(ctx, req.(*StartJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_StopJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).StopJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WorkerService/StopJob",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).StopJob(ctx, req.(*StopJobRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_GetJobStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkerServiceServer).GetJobStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WorkerService/GetJobStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkerServiceServer).GetJobStatus(ctx, req.(*GetStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkerService_GetOutputStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkerServiceServer).GetOutputStream(m, &workerServiceGetOutputStreamServer{stream})
}

type WorkerService_GetOutputStreamServer interface {
	Send(*GetStreamResponse) error
	grpc.ServerStream
}

type workerServiceGetOutputStreamServer struct {
	grpc.ServerStream
}

func (x *workerServiceGetOutputStreamServer) Send(m *GetStreamResponse) error {
	return x.ServerStream.SendMsg(m)
}

// WorkerService_ServiceDesc is the grpc.ServiceDesc for WorkerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.WorkerService",
	HandlerType: (*WorkerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartJob",
			Handler:    _WorkerService_StartJob_Handler,
		},
		{
			MethodName: "StopJob",
			Handler:    _WorkerService_StopJob_Handler,
		},
		{
			MethodName: "GetJobStatus",
			Handler:    _WorkerService_GetJobStatus_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetOutputStream",
			Handler:       _WorkerService_GetOutputStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/workerservice.proto",
}
