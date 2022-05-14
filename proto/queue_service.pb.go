// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: queue_service.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// PushTaskRequest is a request to add a task to the queue.
type PushTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The unique id of this request
	RequestId string `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	// The namespace where the task should be added.
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
	// Whether to asynchronously push the task.
	//
	// Asynchronous task creation does not wait for the task to
	// be persisted on all peers. It is faster but may result in
	// less durable operations.
	Async bool `protobuf:"varint,3,opt,name=async,proto3" json:"async,omitempty"`
	// Arbitrary key/value metadata for the task.
	Metadata map[string]string `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// An arbitrary task payload.
	Payload []byte `protobuf:"bytes,5,opt,name=payload,proto3" json:"payload,omitempty"`
	// Optional schedule for the task if it should be
	// delivered at a specific time.
	//
	// Types that are assignable to Schedule:
	//	*PushTaskRequest_ScheduleTime
	//	*PushTaskRequest_TtsSeconds
	Schedule isPushTaskRequest_Schedule `protobuf_oneof:"schedule"`
}

func (x *PushTaskRequest) Reset() {
	*x = PushTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PushTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PushTaskRequest) ProtoMessage() {}

func (x *PushTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PushTaskRequest.ProtoReflect.Descriptor instead.
func (*PushTaskRequest) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{0}
}

func (x *PushTaskRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *PushTaskRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *PushTaskRequest) GetAsync() bool {
	if x != nil {
		return x.Async
	}
	return false
}

func (x *PushTaskRequest) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *PushTaskRequest) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (m *PushTaskRequest) GetSchedule() isPushTaskRequest_Schedule {
	if m != nil {
		return m.Schedule
	}
	return nil
}

func (x *PushTaskRequest) GetScheduleTime() *timestamppb.Timestamp {
	if x, ok := x.GetSchedule().(*PushTaskRequest_ScheduleTime); ok {
		return x.ScheduleTime
	}
	return nil
}

func (x *PushTaskRequest) GetTtsSeconds() int64 {
	if x, ok := x.GetSchedule().(*PushTaskRequest_TtsSeconds); ok {
		return x.TtsSeconds
	}
	return 0
}

type isPushTaskRequest_Schedule interface {
	isPushTaskRequest_Schedule()
}

type PushTaskRequest_ScheduleTime struct {
	// The time when the task should be delivered to consumers.
	ScheduleTime *timestamppb.Timestamp `protobuf:"bytes,6,opt,name=schedule_time,json=scheduleTime,proto3,oneof"`
}

type PushTaskRequest_TtsSeconds struct {
	// The amount of time to wait before delivering the task to consumers.
	TtsSeconds int64 `protobuf:"varint,7,opt,name=tts_seconds,json=ttsSeconds,proto3,oneof"`
}

func (*PushTaskRequest_ScheduleTime) isPushTaskRequest_Schedule() {}

func (*PushTaskRequest_TtsSeconds) isPushTaskRequest_Schedule() {}

// PushTaskResponse is the result of adding a task to the queue.
type PushTaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The task that was added
	Task *Task `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
	// The peer that handled the task creation
	//
	// Under normal circumstances, this field should be an empty string.
	// If a request to create a task is sent to a follower instead of
	// a leader, the follower will forward the request to the leader
	// and set this field to the leader's address. This is a hint to the
	// client that it should reconnect to the leader.
	ForwardedTo string `protobuf:"bytes,2,opt,name=forwarded_to,json=forwardedTo,proto3" json:"forwarded_to,omitempty"`
}

func (x *PushTaskResponse) Reset() {
	*x = PushTaskResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PushTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PushTaskResponse) ProtoMessage() {}

func (x *PushTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PushTaskResponse.ProtoReflect.Descriptor instead.
func (*PushTaskResponse) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{1}
}

func (x *PushTaskResponse) GetTask() *Task {
	if x != nil {
		return x.Task
	}
	return nil
}

func (x *PushTaskResponse) GetForwardedTo() string {
	if x != nil {
		return x.ForwardedTo
	}
	return ""
}

// PushStreamResponse is a stream response from the server when adding a task.
type PushStreamResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Task        *Task  `protobuf:"bytes,1,opt,name=task,proto3" json:"task,omitempty"`
	Status      uint32 `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
	Message     string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	ForwardedTo string `protobuf:"bytes,4,opt,name=forwarded_to,json=forwardedTo,proto3" json:"forwarded_to,omitempty"`
}

func (x *PushStreamResponse) Reset() {
	*x = PushStreamResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PushStreamResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PushStreamResponse) ProtoMessage() {}

func (x *PushStreamResponse) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PushStreamResponse.ProtoReflect.Descriptor instead.
func (*PushStreamResponse) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{2}
}

func (x *PushStreamResponse) GetTask() *Task {
	if x != nil {
		return x.Task
	}
	return nil
}

func (x *PushStreamResponse) GetStatus() uint32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *PushStreamResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *PushStreamResponse) GetForwardedTo() string {
	if x != nil {
		return x.ForwardedTo
	}
	return ""
}

type GetTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string         `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Reference *TaskReference `protobuf:"bytes,2,opt,name=reference,proto3" json:"reference,omitempty"`
}

func (x *GetTaskRequest) Reset() {
	*x = GetTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTaskRequest) ProtoMessage() {}

func (x *GetTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTaskRequest.ProtoReflect.Descriptor instead.
func (*GetTaskRequest) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetTaskRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *GetTaskRequest) GetReference() *TaskReference {
	if x != nil {
		return x.Reference
	}
	return nil
}

type DeleteTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId string         `protobuf:"bytes,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Async     bool           `protobuf:"varint,2,opt,name=async,proto3" json:"async,omitempty"`
	Reference *TaskReference `protobuf:"bytes,3,opt,name=reference,proto3" json:"reference,omitempty"`
}

func (x *DeleteTaskRequest) Reset() {
	*x = DeleteTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTaskRequest) ProtoMessage() {}

func (x *DeleteTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTaskRequest.ProtoReflect.Descriptor instead.
func (*DeleteTaskRequest) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{4}
}

func (x *DeleteTaskRequest) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *DeleteTaskRequest) GetAsync() bool {
	if x != nil {
		return x.Async
	}
	return false
}

func (x *DeleteTaskRequest) GetReference() *TaskReference {
	if x != nil {
		return x.Reference
	}
	return nil
}

type DeleteTaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ForwardedTo string `protobuf:"bytes,1,opt,name=forwarded_to,json=forwardedTo,proto3" json:"forwarded_to,omitempty"`
}

func (x *DeleteTaskResponse) Reset() {
	*x = DeleteTaskResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_queue_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteTaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteTaskResponse) ProtoMessage() {}

func (x *DeleteTaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_queue_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteTaskResponse.ProtoReflect.Descriptor instead.
func (*DeleteTaskResponse) Descriptor() ([]byte, []int) {
	return file_queue_service_proto_rawDescGZIP(), []int{5}
}

func (x *DeleteTaskResponse) GetForwardedTo() string {
	if x != nil {
		return x.ForwardedTo
	}
	return ""
}

var File_queue_service_proto protoreflect.FileDescriptor

var file_queue_service_proto_rawDesc = []byte{
	0x0a, 0x13, 0x71, 0x75, 0x65, 0x75, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x7a, 0x34, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0b, 0x6d, 0x6f, 0x64, 0x65,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xec, 0x02, 0x0a, 0x0f, 0x50, 0x75, 0x73, 0x68,
	0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6e,
	0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x61, 0x73, 0x79, 0x6e,
	0x63, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61, 0x73, 0x79, 0x6e, 0x63, 0x12, 0x3d,
	0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x21, 0x2e, 0x7a, 0x34, 0x2e, 0x50, 0x75, 0x73, 0x68, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07,
	0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x41, 0x0a, 0x0d, 0x73, 0x63, 0x68, 0x65, 0x64,
	0x75, 0x6c, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x48, 0x00, 0x52, 0x0c, 0x73, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x21, 0x0a, 0x0b, 0x74, 0x74,
	0x73, 0x5f, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x48,
	0x00, 0x52, 0x0a, 0x74, 0x74, 0x73, 0x53, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x73, 0x1a, 0x3b, 0x0a,
	0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x0a, 0x0a, 0x08, 0x73, 0x63,
	0x68, 0x65, 0x64, 0x75, 0x6c, 0x65, 0x22, 0x53, 0x0a, 0x10, 0x50, 0x75, 0x73, 0x68, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x04, 0x74, 0x61,
	0x73, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x7a, 0x34, 0x2e, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x65, 0x64, 0x5f, 0x74, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x65, 0x64, 0x54, 0x6f, 0x22, 0x87, 0x01, 0x0a, 0x12,
	0x50, 0x75, 0x73, 0x68, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x1c, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x08, 0x2e, 0x7a, 0x34, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x04, 0x74, 0x61, 0x73, 0x6b,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x65, 0x64, 0x5f,
	0x74, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x66, 0x6f, 0x72, 0x77, 0x61, 0x72,
	0x64, 0x65, 0x64, 0x54, 0x6f, 0x22, 0x60, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x2f, 0x0a, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65,
	0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x7a, 0x34, 0x2e, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x09, 0x72, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x79, 0x0a, 0x11, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x61,
	0x73, 0x79, 0x6e, 0x63, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x61, 0x73, 0x79, 0x6e,
	0x63, 0x12, 0x2f, 0x0a, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x7a, 0x34, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65,
	0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x09, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e,
	0x63, 0x65, 0x22, 0x37, 0x0a, 0x12, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x21, 0x0a, 0x0c, 0x66, 0x6f, 0x72, 0x77,
	0x61, 0x72, 0x64, 0x65, 0x64, 0x5f, 0x74, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x66, 0x6f, 0x72, 0x77, 0x61, 0x72, 0x64, 0x65, 0x64, 0x54, 0x6f, 0x32, 0x80, 0x02, 0x0a, 0x05,
	0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x50, 0x75, 0x73, 0x68, 0x12, 0x13, 0x2e,
	0x7a, 0x34, 0x2e, 0x50, 0x75, 0x73, 0x68, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x14, 0x2e, 0x7a, 0x34, 0x2e, 0x50, 0x75, 0x73, 0x68, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3f, 0x0a, 0x0a, 0x50, 0x75,
	0x73, 0x68, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x13, 0x2e, 0x7a, 0x34, 0x2e, 0x50, 0x75,
	0x73, 0x68, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x7a, 0x34, 0x2e, 0x50, 0x75, 0x73, 0x68, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x1f, 0x0a, 0x04, 0x50,
	0x75, 0x6c, 0x6c, 0x12, 0x07, 0x2e, 0x7a, 0x34, 0x2e, 0x41, 0x63, 0x6b, 0x1a, 0x08, 0x2e, 0x7a,
	0x34, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x12, 0x25, 0x0a, 0x03,
	0x47, 0x65, 0x74, 0x12, 0x12, 0x2e, 0x7a, 0x34, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x61, 0x73, 0x6b,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x08, 0x2e, 0x7a, 0x34, 0x2e, 0x54, 0x61, 0x73,
	0x6b, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x06, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x12, 0x15, 0x2e,
	0x7a, 0x34, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x7a, 0x34, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x23,
	0x5a, 0x21, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x70,
	0x6f, 0x73, 0x65, 0x79, 0x2f, 0x7a, 0x34, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_queue_service_proto_rawDescOnce sync.Once
	file_queue_service_proto_rawDescData = file_queue_service_proto_rawDesc
)

func file_queue_service_proto_rawDescGZIP() []byte {
	file_queue_service_proto_rawDescOnce.Do(func() {
		file_queue_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_queue_service_proto_rawDescData)
	})
	return file_queue_service_proto_rawDescData
}

var file_queue_service_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_queue_service_proto_goTypes = []interface{}{
	(*PushTaskRequest)(nil),       // 0: z4.PushTaskRequest
	(*PushTaskResponse)(nil),      // 1: z4.PushTaskResponse
	(*PushStreamResponse)(nil),    // 2: z4.PushStreamResponse
	(*GetTaskRequest)(nil),        // 3: z4.GetTaskRequest
	(*DeleteTaskRequest)(nil),     // 4: z4.DeleteTaskRequest
	(*DeleteTaskResponse)(nil),    // 5: z4.DeleteTaskResponse
	nil,                           // 6: z4.PushTaskRequest.MetadataEntry
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
	(*Task)(nil),                  // 8: z4.Task
	(*TaskReference)(nil),         // 9: z4.TaskReference
	(*Ack)(nil),                   // 10: z4.Ack
}
var file_queue_service_proto_depIdxs = []int32{
	6,  // 0: z4.PushTaskRequest.metadata:type_name -> z4.PushTaskRequest.MetadataEntry
	7,  // 1: z4.PushTaskRequest.schedule_time:type_name -> google.protobuf.Timestamp
	8,  // 2: z4.PushTaskResponse.task:type_name -> z4.Task
	8,  // 3: z4.PushStreamResponse.task:type_name -> z4.Task
	9,  // 4: z4.GetTaskRequest.reference:type_name -> z4.TaskReference
	9,  // 5: z4.DeleteTaskRequest.reference:type_name -> z4.TaskReference
	0,  // 6: z4.Queue.Push:input_type -> z4.PushTaskRequest
	0,  // 7: z4.Queue.PushStream:input_type -> z4.PushTaskRequest
	10, // 8: z4.Queue.Pull:input_type -> z4.Ack
	3,  // 9: z4.Queue.Get:input_type -> z4.GetTaskRequest
	4,  // 10: z4.Queue.Delete:input_type -> z4.DeleteTaskRequest
	1,  // 11: z4.Queue.Push:output_type -> z4.PushTaskResponse
	2,  // 12: z4.Queue.PushStream:output_type -> z4.PushStreamResponse
	8,  // 13: z4.Queue.Pull:output_type -> z4.Task
	8,  // 14: z4.Queue.Get:output_type -> z4.Task
	5,  // 15: z4.Queue.Delete:output_type -> z4.DeleteTaskResponse
	11, // [11:16] is the sub-list for method output_type
	6,  // [6:11] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_queue_service_proto_init() }
func file_queue_service_proto_init() {
	if File_queue_service_proto != nil {
		return
	}
	file_model_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_queue_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PushTaskRequest); i {
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
		file_queue_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PushTaskResponse); i {
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
		file_queue_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PushStreamResponse); i {
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
		file_queue_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTaskRequest); i {
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
		file_queue_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTaskRequest); i {
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
		file_queue_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteTaskResponse); i {
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
	file_queue_service_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*PushTaskRequest_ScheduleTime)(nil),
		(*PushTaskRequest_TtsSeconds)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_queue_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_queue_service_proto_goTypes,
		DependencyIndexes: file_queue_service_proto_depIdxs,
		MessageInfos:      file_queue_service_proto_msgTypes,
	}.Build()
	File_queue_service_proto = out.File
	file_queue_service_proto_rawDesc = nil
	file_queue_service_proto_goTypes = nil
	file_queue_service_proto_depIdxs = nil
}
