// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: admin_service.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CheckHealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A unique id for the request.
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *CheckHealthRequest) Reset() {
	*x = CheckHealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckHealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckHealthRequest) ProtoMessage() {}

func (x *CheckHealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckHealthRequest.ProtoReflect.Descriptor instead.
func (*CheckHealthRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{0}
}

func (x *CheckHealthRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{1}
}

type GetNamespaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId   string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	NamespaceId string `protobuf:"bytes,2,opt,name=namespace_id,json=namespaceId,proto3" json:"namespace_id,omitempty"`
}

func (x *GetNamespaceRequest) Reset() {
	*x = GetNamespaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetNamespaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetNamespaceRequest) ProtoMessage() {}

func (x *GetNamespaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetNamespaceRequest.ProtoReflect.Descriptor instead.
func (*GetNamespaceRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetNamespaceRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *GetNamespaceRequest) GetNamespaceId() string {
	if x != nil {
		return x.NamespaceId
	}
	return ""
}

type UpdateNamespaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// A new version of the namespace
	//
	// The following fields cannot be updated
	// - id
	// - last_delivered_task
	Namespace *Namespace `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
}

func (x *UpdateNamespaceRequest) Reset() {
	*x = UpdateNamespaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateNamespaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateNamespaceRequest) ProtoMessage() {}

func (x *UpdateNamespaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateNamespaceRequest.ProtoReflect.Descriptor instead.
func (*UpdateNamespaceRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{3}
}

func (x *UpdateNamespaceRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *UpdateNamespaceRequest) GetNamespace() *Namespace {
	if x != nil {
		return x.Namespace
	}
	return nil
}

type GetClusterInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
}

func (x *GetClusterInfoRequest) Reset() {
	*x = GetClusterInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetClusterInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetClusterInfoRequest) ProtoMessage() {}

func (x *GetClusterInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetClusterInfoRequest.ProtoReflect.Descriptor instead.
func (*GetClusterInfoRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{4}
}

func (x *GetClusterInfoRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

// ClusterInfo contains information about the cluster.
type ClusterInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique id of the peer that the request was sent to.
	ServerId string `protobuf:"bytes,1,opt,name=server_id,json=serverId,proto3" json:"server_id,omitempty"`
	// The host:port of the cluster leader's Raft server.
	LeaderAddress string `protobuf:"bytes,2,opt,name=leader_address,json=leaderAddress,proto3" json:"leader_address,omitempty"`
	// All peers of the cluster.
	Members []*Server `protobuf:"bytes,3,rep,name=members,proto3" json:"members,omitempty"`
}

func (x *ClusterInfo) Reset() {
	*x = ClusterInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterInfo) ProtoMessage() {}

func (x *ClusterInfo) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterInfo.ProtoReflect.Descriptor instead.
func (*ClusterInfo) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{5}
}

func (x *ClusterInfo) GetServerId() string {
	if x != nil {
		return x.ServerId
	}
	return ""
}

func (x *ClusterInfo) GetLeaderAddress() string {
	if x != nil {
		return x.LeaderAddress
	}
	return ""
}

func (x *ClusterInfo) GetMembers() []*Server {
	if x != nil {
		return x.Members
	}
	return nil
}

// TODO: Rename this to Peer.
// Server is a member of the cluster.
type Server struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique id of the peer.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The host:port of the peer's Raft server.
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *Server) Reset() {
	*x = Server{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Server) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Server) ProtoMessage() {}

func (x *Server) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Server.ProtoReflect.Descriptor instead.
func (*Server) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{6}
}

func (x *Server) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Server) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type AddClusterMemberRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The host:port of the peer's Raft server.
	MemberAddress string `protobuf:"bytes,1,opt,name=member_address,json=memberAddress,proto3" json:"member_address,omitempty"`
	// The unique id of the peer.
	MemberId string `protobuf:"bytes,2,opt,name=member_id,json=memberId,proto3" json:"member_id,omitempty"`
}

func (x *AddClusterMemberRequest) Reset() {
	*x = AddClusterMemberRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddClusterMemberRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddClusterMemberRequest) ProtoMessage() {}

func (x *AddClusterMemberRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddClusterMemberRequest.ProtoReflect.Descriptor instead.
func (*AddClusterMemberRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{7}
}

func (x *AddClusterMemberRequest) GetMemberAddress() string {
	if x != nil {
		return x.MemberAddress
	}
	return ""
}

func (x *AddClusterMemberRequest) GetMemberId() string {
	if x != nil {
		return x.MemberId
	}
	return ""
}

type RemoveClusterMemberRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique id of the peer.
	MemberId string `protobuf:"bytes,1,opt,name=member_id,json=memberId,proto3" json:"member_id,omitempty"`
}

func (x *RemoveClusterMemberRequest) Reset() {
	*x = RemoveClusterMemberRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_admin_service_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveClusterMemberRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveClusterMemberRequest) ProtoMessage() {}

func (x *RemoveClusterMemberRequest) ProtoReflect() protoreflect.Message {
	mi := &file_admin_service_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveClusterMemberRequest.ProtoReflect.Descriptor instead.
func (*RemoveClusterMemberRequest) Descriptor() ([]byte, []int) {
	return file_admin_service_proto_rawDescGZIP(), []int{8}
}

func (x *RemoveClusterMemberRequest) GetMemberId() string {
	if x != nil {
		return x.MemberId
	}
	return ""
}

var File_admin_service_proto protoreflect.FileDescriptor

var file_admin_service_proto_rawDesc = []byte{
	0x0a, 0x13, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x7a, 0x34, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x33, 0x0a, 0x12, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x22, 0x08, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x22, 0x57, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61,
	0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x6e, 0x61, 0x6d, 0x65,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x22, 0x64, 0x0a, 0x16, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x49, 0x64, 0x12, 0x2b, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x7a, 0x34, 0x2e, 0x4e, 0x61, 0x6d,
	0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x22, 0x36, 0x0a, 0x15, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x22, 0x77, 0x0a, 0x0b, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x6c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x5f,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x24, 0x0a, 0x07,
	0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0a, 0x2e,
	0x7a, 0x34, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x52, 0x07, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x22, 0x32, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07,
	0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x5d, 0x0a, 0x17, 0x41, 0x64, 0x64, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x25, 0x0a, 0x0e, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x65, 0x6d,
	0x62, 0x65, 0x72, 0x49, 0x64, 0x22, 0x39, 0x0a, 0x1a, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x49, 0x64,
	0x32, 0x97, 0x04, 0x0a, 0x05, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x33, 0x0a, 0x0b, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x12, 0x16, 0x2e, 0x7a, 0x34, 0x2e, 0x43,
	0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x0a, 0x2e, 0x7a, 0x34, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12,
	0x38, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12,
	0x17, 0x2e, 0x7a, 0x34, 0x2e, 0x47, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x7a, 0x34, 0x2e, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0f, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x1a, 0x2e, 0x7a,
	0x34, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x7a, 0x34, 0x2e, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x0a, 0x50, 0x75, 0x72,
	0x67, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x12, 0x15, 0x2e, 0x7a, 0x34, 0x2e, 0x50, 0x75, 0x72,
	0x67, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x19, 0x2e, 0x7a, 0x34, 0x2e,
	0x47, 0x65, 0x74, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x7a, 0x34, 0x2e, 0x43, 0x6c, 0x75, 0x73, 0x74,
	0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x00, 0x12, 0x49, 0x0a, 0x10, 0x41, 0x64, 0x64, 0x43,
	0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1b, 0x2e, 0x7a,
	0x34, 0x2e, 0x41, 0x64, 0x64, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74,
	0x79, 0x22, 0x00, 0x12, 0x4f, 0x0a, 0x13, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x75,
	0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1e, 0x2e, 0x7a, 0x34, 0x2e,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4d, 0x65, 0x6d,
	0x62, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x00, 0x12, 0x44, 0x0a, 0x10, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61,
	0x70, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x70, 0x6f, 0x73, 0x65, 0x79,
	0x2f, 0x7a, 0x34, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_admin_service_proto_rawDescOnce sync.Once
	file_admin_service_proto_rawDescData = file_admin_service_proto_rawDesc
)

func file_admin_service_proto_rawDescGZIP() []byte {
	file_admin_service_proto_rawDescOnce.Do(func() {
		file_admin_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_admin_service_proto_rawDescData)
	})
	return file_admin_service_proto_rawDescData
}

var file_admin_service_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_admin_service_proto_goTypes = []interface{}{
	(*CheckHealthRequest)(nil),         // 0: z4.CheckHealthRequest
	(*Status)(nil),                     // 1: z4.Status
	(*GetNamespaceRequest)(nil),        // 2: z4.GetNamespaceRequest
	(*UpdateNamespaceRequest)(nil),     // 3: z4.UpdateNamespaceRequest
	(*GetClusterInfoRequest)(nil),      // 4: z4.GetClusterInfoRequest
	(*ClusterInfo)(nil),                // 5: z4.ClusterInfo
	(*Server)(nil),                     // 6: z4.Server
	(*AddClusterMemberRequest)(nil),    // 7: z4.AddClusterMemberRequest
	(*RemoveClusterMemberRequest)(nil), // 8: z4.RemoveClusterMemberRequest
	(*Namespace)(nil),                  // 9: z4.Namespace
	(*PurgeTasksRequest)(nil),          // 10: z4.PurgeTasksRequest
	(*emptypb.Empty)(nil),              // 11: google.protobuf.Empty
}
var file_admin_service_proto_depIdxs = []int32{
	9,  // 0: z4.UpdateNamespaceRequest.namespace:type_name -> z4.Namespace
	6,  // 1: z4.ClusterInfo.members:type_name -> z4.Server
	0,  // 2: z4.Admin.CheckHealth:input_type -> z4.CheckHealthRequest
	2,  // 3: z4.Admin.GetNamespace:input_type -> z4.GetNamespaceRequest
	3,  // 4: z4.Admin.UpdateNamespace:input_type -> z4.UpdateNamespaceRequest
	10, // 5: z4.Admin.PurgeTasks:input_type -> z4.PurgeTasksRequest
	4,  // 6: z4.Admin.GetClusterInfo:input_type -> z4.GetClusterInfoRequest
	7,  // 7: z4.Admin.AddClusterMember:input_type -> z4.AddClusterMemberRequest
	8,  // 8: z4.Admin.RemoveClusterMember:input_type -> z4.RemoveClusterMemberRequest
	11, // 9: z4.Admin.BootstrapCluster:input_type -> google.protobuf.Empty
	1,  // 10: z4.Admin.CheckHealth:output_type -> z4.Status
	9,  // 11: z4.Admin.GetNamespace:output_type -> z4.Namespace
	9,  // 12: z4.Admin.UpdateNamespace:output_type -> z4.Namespace
	11, // 13: z4.Admin.PurgeTasks:output_type -> google.protobuf.Empty
	5,  // 14: z4.Admin.GetClusterInfo:output_type -> z4.ClusterInfo
	11, // 15: z4.Admin.AddClusterMember:output_type -> google.protobuf.Empty
	11, // 16: z4.Admin.RemoveClusterMember:output_type -> google.protobuf.Empty
	11, // 17: z4.Admin.BootstrapCluster:output_type -> google.protobuf.Empty
	10, // [10:18] is the sub-list for method output_type
	2,  // [2:10] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_admin_service_proto_init() }
func file_admin_service_proto_init() {
	if File_admin_service_proto != nil {
		return
	}
	file_model_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_admin_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckHealthRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetNamespaceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateNamespaceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetClusterInfoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Server); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddClusterMemberRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_admin_service_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveClusterMemberRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_admin_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_admin_service_proto_goTypes,
		DependencyIndexes: file_admin_service_proto_depIdxs,
		MessageInfos:      file_admin_service_proto_msgTypes,
	}.Build()
	File_admin_service_proto = out.File
	file_admin_service_proto_rawDesc = nil
	file_admin_service_proto_goTypes = nil
	file_admin_service_proto_depIdxs = nil
}
