// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: admin_service.proto

package proto

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

// AdminClient is the client API for Admin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AdminClient interface {
	// CheckHealth determines whether the service is in a healthy state.
	CheckHealth(ctx context.Context, in *CheckHealthRequest, opts ...grpc.CallOption) (*Status, error)
	// GetQueue gets settings for a task queue.
	GetQueue(ctx context.Context, in *GetQueueRequest, opts ...grpc.CallOption) (*QueueConfig, error)
	// UpdateQueue updates the settings of a task queue.
	UpdateQueue(ctx context.Context, in *UpdateQueueRequest, opts ...grpc.CallOption) (*QueueConfig, error)
	// GetClusterInfo returns information about the structure of the Raft cluster.
	GetClusterInfo(ctx context.Context, in *GetClusterInfoRequest, opts ...grpc.CallOption) (*ClusterInfo, error)
	// AddClusterMember adds a peer to the Raft cluster.
	//
	// This rpc should be called on the leader but can be invoked on any peer that
	// is connected to the leader.
	AddClusterMember(ctx context.Context, in *AddClusterMemberRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// RemoveClusterMember removes a peer from the Raft cluster.
	//
	// This rpc should be called on the leader but can be invoked on any peer that
	// is connected to the leader.
	RemoveClusterMember(ctx context.Context, in *RemoveClusterMemberRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	BootstrapCluster(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type adminClient struct {
	cc grpc.ClientConnInterface
}

func NewAdminClient(cc grpc.ClientConnInterface) AdminClient {
	return &adminClient{cc}
}

func (c *adminClient) CheckHealth(ctx context.Context, in *CheckHealthRequest, opts ...grpc.CallOption) (*Status, error) {
	out := new(Status)
	err := c.cc.Invoke(ctx, "/z4.Admin/CheckHealth", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) GetQueue(ctx context.Context, in *GetQueueRequest, opts ...grpc.CallOption) (*QueueConfig, error) {
	out := new(QueueConfig)
	err := c.cc.Invoke(ctx, "/z4.Admin/GetQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) UpdateQueue(ctx context.Context, in *UpdateQueueRequest, opts ...grpc.CallOption) (*QueueConfig, error) {
	out := new(QueueConfig)
	err := c.cc.Invoke(ctx, "/z4.Admin/UpdateQueue", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) GetClusterInfo(ctx context.Context, in *GetClusterInfoRequest, opts ...grpc.CallOption) (*ClusterInfo, error) {
	out := new(ClusterInfo)
	err := c.cc.Invoke(ctx, "/z4.Admin/GetClusterInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) AddClusterMember(ctx context.Context, in *AddClusterMemberRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/z4.Admin/AddClusterMember", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) RemoveClusterMember(ctx context.Context, in *RemoveClusterMemberRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/z4.Admin/RemoveClusterMember", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *adminClient) BootstrapCluster(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/z4.Admin/BootstrapCluster", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AdminServer is the server API for Admin service.
// All implementations must embed UnimplementedAdminServer
// for forward compatibility
type AdminServer interface {
	// CheckHealth determines whether the service is in a healthy state.
	CheckHealth(context.Context, *CheckHealthRequest) (*Status, error)
	// GetQueue gets settings for a task queue.
	GetQueue(context.Context, *GetQueueRequest) (*QueueConfig, error)
	// UpdateQueue updates the settings of a task queue.
	UpdateQueue(context.Context, *UpdateQueueRequest) (*QueueConfig, error)
	// GetClusterInfo returns information about the structure of the Raft cluster.
	GetClusterInfo(context.Context, *GetClusterInfoRequest) (*ClusterInfo, error)
	// AddClusterMember adds a peer to the Raft cluster.
	//
	// This rpc should be called on the leader but can be invoked on any peer that
	// is connected to the leader.
	AddClusterMember(context.Context, *AddClusterMemberRequest) (*emptypb.Empty, error)
	// RemoveClusterMember removes a peer from the Raft cluster.
	//
	// This rpc should be called on the leader but can be invoked on any peer that
	// is connected to the leader.
	RemoveClusterMember(context.Context, *RemoveClusterMemberRequest) (*emptypb.Empty, error)
	BootstrapCluster(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedAdminServer()
}

// UnimplementedAdminServer must be embedded to have forward compatible implementations.
type UnimplementedAdminServer struct {
}

func (UnimplementedAdminServer) CheckHealth(context.Context, *CheckHealthRequest) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckHealth not implemented")
}
func (UnimplementedAdminServer) GetQueue(context.Context, *GetQueueRequest) (*QueueConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetQueue not implemented")
}
func (UnimplementedAdminServer) UpdateQueue(context.Context, *UpdateQueueRequest) (*QueueConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateQueue not implemented")
}
func (UnimplementedAdminServer) GetClusterInfo(context.Context, *GetClusterInfoRequest) (*ClusterInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetClusterInfo not implemented")
}
func (UnimplementedAdminServer) AddClusterMember(context.Context, *AddClusterMemberRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddClusterMember not implemented")
}
func (UnimplementedAdminServer) RemoveClusterMember(context.Context, *RemoveClusterMemberRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveClusterMember not implemented")
}
func (UnimplementedAdminServer) BootstrapCluster(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BootstrapCluster not implemented")
}
func (UnimplementedAdminServer) mustEmbedUnimplementedAdminServer() {}

// UnsafeAdminServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AdminServer will
// result in compilation errors.
type UnsafeAdminServer interface {
	mustEmbedUnimplementedAdminServer()
}

func RegisterAdminServer(s grpc.ServiceRegistrar, srv AdminServer) {
	s.RegisterService(&Admin_ServiceDesc, srv)
}

func _Admin_CheckHealth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckHealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).CheckHealth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/CheckHealth",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).CheckHealth(ctx, req.(*CheckHealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_GetQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetQueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).GetQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/GetQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).GetQueue(ctx, req.(*GetQueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_UpdateQueue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateQueueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).UpdateQueue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/UpdateQueue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).UpdateQueue(ctx, req.(*UpdateQueueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_GetClusterInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetClusterInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).GetClusterInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/GetClusterInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).GetClusterInfo(ctx, req.(*GetClusterInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_AddClusterMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddClusterMemberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).AddClusterMember(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/AddClusterMember",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).AddClusterMember(ctx, req.(*AddClusterMemberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_RemoveClusterMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveClusterMemberRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).RemoveClusterMember(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/RemoveClusterMember",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).RemoveClusterMember(ctx, req.(*RemoveClusterMemberRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Admin_BootstrapCluster_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AdminServer).BootstrapCluster(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/z4.Admin/BootstrapCluster",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AdminServer).BootstrapCluster(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// Admin_ServiceDesc is the grpc.ServiceDesc for Admin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Admin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "z4.Admin",
	HandlerType: (*AdminServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckHealth",
			Handler:    _Admin_CheckHealth_Handler,
		},
		{
			MethodName: "GetQueue",
			Handler:    _Admin_GetQueue_Handler,
		},
		{
			MethodName: "UpdateQueue",
			Handler:    _Admin_UpdateQueue_Handler,
		},
		{
			MethodName: "GetClusterInfo",
			Handler:    _Admin_GetClusterInfo_Handler,
		},
		{
			MethodName: "AddClusterMember",
			Handler:    _Admin_AddClusterMember_Handler,
		},
		{
			MethodName: "RemoveClusterMember",
			Handler:    _Admin_RemoveClusterMember_Handler,
		},
		{
			MethodName: "BootstrapCluster",
			Handler:    _Admin_BootstrapCluster_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "admin_service.proto",
}
