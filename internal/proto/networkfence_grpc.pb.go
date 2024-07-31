// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.0
// - protoc             v3.20.2
// source: networkfence.proto

package proto

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
	NetworkFence_FenceClusterNetwork_FullMethodName   = "/proto.NetworkFence/FenceClusterNetwork"
	NetworkFence_UnFenceClusterNetwork_FullMethodName = "/proto.NetworkFence/UnFenceClusterNetwork"
)

// NetworkFenceClient is the client API for NetworkFence service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// NetworkFence holds the RPC method for allowing the communication between
// the CSIAddons controller and the sidecar for fencing operations.
type NetworkFenceClient interface {
	// FenceClusterNetwork RPC call to fence the cluster network.
	FenceClusterNetwork(ctx context.Context, in *NetworkFenceRequest, opts ...grpc.CallOption) (*NetworkFenceResponse, error)
	// UnFenceClusterNetwork RPC call to un-fence the cluster network.
	UnFenceClusterNetwork(ctx context.Context, in *NetworkFenceRequest, opts ...grpc.CallOption) (*NetworkFenceResponse, error)
}

type networkFenceClient struct {
	cc grpc.ClientConnInterface
}

func NewNetworkFenceClient(cc grpc.ClientConnInterface) NetworkFenceClient {
	return &networkFenceClient{cc}
}

func (c *networkFenceClient) FenceClusterNetwork(ctx context.Context, in *NetworkFenceRequest, opts ...grpc.CallOption) (*NetworkFenceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NetworkFenceResponse)
	err := c.cc.Invoke(ctx, NetworkFence_FenceClusterNetwork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *networkFenceClient) UnFenceClusterNetwork(ctx context.Context, in *NetworkFenceRequest, opts ...grpc.CallOption) (*NetworkFenceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(NetworkFenceResponse)
	err := c.cc.Invoke(ctx, NetworkFence_UnFenceClusterNetwork_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NetworkFenceServer is the server API for NetworkFence service.
// All implementations must embed UnimplementedNetworkFenceServer
// for forward compatibility.
//
// NetworkFence holds the RPC method for allowing the communication between
// the CSIAddons controller and the sidecar for fencing operations.
type NetworkFenceServer interface {
	// FenceClusterNetwork RPC call to fence the cluster network.
	FenceClusterNetwork(context.Context, *NetworkFenceRequest) (*NetworkFenceResponse, error)
	// UnFenceClusterNetwork RPC call to un-fence the cluster network.
	UnFenceClusterNetwork(context.Context, *NetworkFenceRequest) (*NetworkFenceResponse, error)
	mustEmbedUnimplementedNetworkFenceServer()
}

// UnimplementedNetworkFenceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedNetworkFenceServer struct{}

func (UnimplementedNetworkFenceServer) FenceClusterNetwork(context.Context, *NetworkFenceRequest) (*NetworkFenceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FenceClusterNetwork not implemented")
}
func (UnimplementedNetworkFenceServer) UnFenceClusterNetwork(context.Context, *NetworkFenceRequest) (*NetworkFenceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnFenceClusterNetwork not implemented")
}
func (UnimplementedNetworkFenceServer) mustEmbedUnimplementedNetworkFenceServer() {}
func (UnimplementedNetworkFenceServer) testEmbeddedByValue()                      {}

// UnsafeNetworkFenceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NetworkFenceServer will
// result in compilation errors.
type UnsafeNetworkFenceServer interface {
	mustEmbedUnimplementedNetworkFenceServer()
}

func RegisterNetworkFenceServer(s grpc.ServiceRegistrar, srv NetworkFenceServer) {
	// If the following call pancis, it indicates UnimplementedNetworkFenceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&NetworkFence_ServiceDesc, srv)
}

func _NetworkFence_FenceClusterNetwork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkFenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkFenceServer).FenceClusterNetwork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NetworkFence_FenceClusterNetwork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkFenceServer).FenceClusterNetwork(ctx, req.(*NetworkFenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NetworkFence_UnFenceClusterNetwork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetworkFenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NetworkFenceServer).UnFenceClusterNetwork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NetworkFence_UnFenceClusterNetwork_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NetworkFenceServer).UnFenceClusterNetwork(ctx, req.(*NetworkFenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NetworkFence_ServiceDesc is the grpc.ServiceDesc for NetworkFence service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NetworkFence_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.NetworkFence",
	HandlerType: (*NetworkFenceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FenceClusterNetwork",
			Handler:    _NetworkFence_FenceClusterNetwork_Handler,
		},
		{
			MethodName: "UnFenceClusterNetwork",
			Handler:    _NetworkFence_UnFenceClusterNetwork_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "networkfence.proto",
}
