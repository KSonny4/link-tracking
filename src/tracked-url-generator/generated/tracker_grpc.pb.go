// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package generated

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

// TrackerClient is the client API for Tracker service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TrackerClient interface {
	GetUrl(ctx context.Context, in *URLGenerateRequest, opts ...grpc.CallOption) (*Url, error)
}

type trackerClient struct {
	cc grpc.ClientConnInterface
}

func NewTrackerClient(cc grpc.ClientConnInterface) TrackerClient {
	return &trackerClient{cc}
}

func (c *trackerClient) GetUrl(ctx context.Context, in *URLGenerateRequest, opts ...grpc.CallOption) (*Url, error) {
	out := new(Url)
	err := c.cc.Invoke(ctx, "/protos.Tracker/GetUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrackerServer is the server API for Tracker service.
// All implementations must embed UnimplementedTrackerServer
// for forward compatibility
type TrackerServer interface {
	GetUrl(context.Context, *URLGenerateRequest) (*Url, error)
	mustEmbedUnimplementedTrackerServer()
}

// UnimplementedTrackerServer must be embedded to have forward compatible implementations.
type UnimplementedTrackerServer struct {
}

func (UnimplementedTrackerServer) GetUrl(context.Context, *URLGenerateRequest) (*Url, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUrl not implemented")
}
func (UnimplementedTrackerServer) mustEmbedUnimplementedTrackerServer() {}

// UnsafeTrackerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TrackerServer will
// result in compilation errors.
type UnsafeTrackerServer interface {
	mustEmbedUnimplementedTrackerServer()
}

func RegisterTrackerServer(s grpc.ServiceRegistrar, srv TrackerServer) {
	s.RegisterService(&Tracker_ServiceDesc, srv)
}

func _Tracker_GetUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(URLGenerateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrackerServer).GetUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.Tracker/GetUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrackerServer).GetUrl(ctx, req.(*URLGenerateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Tracker_ServiceDesc is the grpc.ServiceDesc for Tracker service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tracker_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protos.Tracker",
	HandlerType: (*TrackerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetUrl",
			Handler:    _Tracker_GetUrl_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "protos/tracker.proto",
}
