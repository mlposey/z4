// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.19.4
// source: misc.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Command is any command replicated by Raft peers and applied to their FSM.
type Command struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Cmd:
	//	*Command_Task
	//	*Command_Ack
	Cmd isCommand_Cmd `protobuf_oneof:"cmd"`
}

func (x *Command) Reset() {
	*x = Command{}
	if protoimpl.UnsafeEnabled {
		mi := &file_misc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Command) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Command) ProtoMessage() {}

func (x *Command) ProtoReflect() protoreflect.Message {
	mi := &file_misc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Command.ProtoReflect.Descriptor instead.
func (*Command) Descriptor() ([]byte, []int) {
	return file_misc_proto_rawDescGZIP(), []int{0}
}

func (m *Command) GetCmd() isCommand_Cmd {
	if m != nil {
		return m.Cmd
	}
	return nil
}

func (x *Command) GetTask() *Task {
	if x, ok := x.GetCmd().(*Command_Task); ok {
		return x.Task
	}
	return nil
}

func (x *Command) GetAck() *Ack {
	if x, ok := x.GetCmd().(*Command_Ack); ok {
		return x.Ack
	}
	return nil
}

type isCommand_Cmd interface {
	isCommand_Cmd()
}

type Command_Task struct {
	// A request to save a task.
	Task *Task `protobuf:"bytes,1,opt,name=task,proto3,oneof"`
}

type Command_Ack struct {
	// A request to acknowledge a task so that it is not redelivered.
	Ack *Ack `protobuf:"bytes,2,opt,name=ack,proto3,oneof"`
}

func (*Command_Task) isCommand_Cmd() {}

func (*Command_Ack) isCommand_Cmd() {}

var File_misc_proto protoreflect.FileDescriptor

var file_misc_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x6d, 0x69, 0x73, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x7a, 0x34,
	0x1a, 0x18, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4d, 0x0a, 0x07, 0x43, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x1e, 0x0a, 0x04, 0x74, 0x61, 0x73, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x7a, 0x34, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x48, 0x00, 0x52,
	0x04, 0x74, 0x61, 0x73, 0x6b, 0x12, 0x1b, 0x0a, 0x03, 0x61, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x07, 0x2e, 0x7a, 0x34, 0x2e, 0x41, 0x63, 0x6b, 0x48, 0x00, 0x52, 0x03, 0x61,
	0x63, 0x6b, 0x42, 0x05, 0x0a, 0x03, 0x63, 0x6d, 0x64, 0x42, 0x23, 0x5a, 0x21, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x70, 0x6f, 0x73, 0x65, 0x79, 0x2f,
	0x7a, 0x34, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_misc_proto_rawDescOnce sync.Once
	file_misc_proto_rawDescData = file_misc_proto_rawDesc
)

func file_misc_proto_rawDescGZIP() []byte {
	file_misc_proto_rawDescOnce.Do(func() {
		file_misc_proto_rawDescData = protoimpl.X.CompressGZIP(file_misc_proto_rawDescData)
	})
	return file_misc_proto_rawDescData
}

var file_misc_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_misc_proto_goTypes = []interface{}{
	(*Command)(nil), // 0: z4.Command
	(*Task)(nil),    // 1: z4.Task
	(*Ack)(nil),     // 2: z4.Ack
}
var file_misc_proto_depIdxs = []int32{
	1, // 0: z4.Command.task:type_name -> z4.Task
	2, // 1: z4.Command.ack:type_name -> z4.Ack
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_misc_proto_init() }
func file_misc_proto_init() {
	if File_misc_proto != nil {
		return
	}
	file_collection_service_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_misc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Command); i {
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
	file_misc_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Command_Task)(nil),
		(*Command_Ack)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_misc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_misc_proto_goTypes,
		DependencyIndexes: file_misc_proto_depIdxs,
		MessageInfos:      file_misc_proto_msgTypes,
	}.Build()
	File_misc_proto = out.File
	file_misc_proto_rawDesc = nil
	file_misc_proto_goTypes = nil
	file_misc_proto_depIdxs = nil
}
