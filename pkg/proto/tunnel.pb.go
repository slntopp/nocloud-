// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.18.1
// source: pkg/proto/tunnel.proto

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

type InitTunnelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
}

func (x *InitTunnelRequest) Reset() {
	*x = InitTunnelRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tunnel_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InitTunnelRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitTunnelRequest) ProtoMessage() {}

func (x *InitTunnelRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tunnel_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitTunnelRequest.ProtoReflect.Descriptor instead.
func (*InitTunnelRequest) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tunnel_proto_rawDescGZIP(), []int{0}
}

func (x *InitTunnelRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

type StreamData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *StreamData) Reset() {
	*x = StreamData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tunnel_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamData) ProtoMessage() {}

func (x *StreamData) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tunnel_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamData.ProtoReflect.Descriptor instead.
func (*StreamData) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tunnel_proto_rawDescGZIP(), []int{1}
}

func (x *StreamData) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SendDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Host    string `protobuf:"bytes,1,opt,name=host,proto3" json:"host,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *SendDataRequest) Reset() {
	*x = SendDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tunnel_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendDataRequest) ProtoMessage() {}

func (x *SendDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tunnel_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendDataRequest.ProtoReflect.Descriptor instead.
func (*SendDataRequest) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tunnel_proto_rawDescGZIP(), []int{2}
}

func (x *SendDataRequest) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *SendDataRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SendDataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Result bool `protobuf:"varint,1,opt,name=result,proto3" json:"result,omitempty"`
}

func (x *SendDataResponse) Reset() {
	*x = SendDataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tunnel_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendDataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendDataResponse) ProtoMessage() {}

func (x *SendDataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tunnel_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendDataResponse.ProtoReflect.Descriptor instead.
func (*SendDataResponse) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tunnel_proto_rawDescGZIP(), []int{3}
}

func (x *SendDataResponse) GetResult() bool {
	if x != nil {
		return x.Result
	}
	return false
}

var File_pkg_proto_tunnel_proto protoreflect.FileDescriptor

var file_pkg_proto_tunnel_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x75, 0x6e, 0x6e,
	0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c,
	0x22, 0x27, 0x0a, 0x11, 0x49, 0x6e, 0x69, 0x74, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x22, 0x26, 0x0a, 0x0a, 0x53, 0x74, 0x72,
	0x65, 0x61, 0x6d, 0x44, 0x61, 0x74, 0x61, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x22, 0x3f, 0x0a, 0x0f, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x22, 0x2a, 0x0a, 0x10, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x32, 0x86,
	0x01, 0x0a, 0x06, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x3d, 0x0a, 0x0a, 0x49, 0x6e, 0x69,
	0x74, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x19, 0x2e, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c,
	0x2e, 0x49, 0x6e, 0x69, 0x74, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x12, 0x2e, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x44, 0x61, 0x74, 0x61, 0x30, 0x01, 0x12, 0x3d, 0x0a, 0x08, 0x53, 0x65, 0x6e, 0x64,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x17, 0x2e, 0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x65,
	0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e,
	0x74, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x2e, 0x53, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x77, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x2e, 0x74,
	0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x42, 0x0b, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x24, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x73, 0x6c, 0x6e, 0x74, 0x6f, 0x70, 0x70, 0x2f, 0x6e, 0x6f, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58,
	0xaa, 0x02, 0x06, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0xca, 0x02, 0x06, 0x54, 0x75, 0x6e, 0x6e,
	0x65, 0x6c, 0xe2, 0x02, 0x12, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c, 0x5c, 0x47, 0x50, 0x42, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x06, 0x54, 0x75, 0x6e, 0x6e, 0x65, 0x6c,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_proto_tunnel_proto_rawDescOnce sync.Once
	file_pkg_proto_tunnel_proto_rawDescData = file_pkg_proto_tunnel_proto_rawDesc
)

func file_pkg_proto_tunnel_proto_rawDescGZIP() []byte {
	file_pkg_proto_tunnel_proto_rawDescOnce.Do(func() {
		file_pkg_proto_tunnel_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_proto_tunnel_proto_rawDescData)
	})
	return file_pkg_proto_tunnel_proto_rawDescData
}

var file_pkg_proto_tunnel_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pkg_proto_tunnel_proto_goTypes = []interface{}{
	(*InitTunnelRequest)(nil), // 0: tunnel.InitTunnelRequest
	(*StreamData)(nil),        // 1: tunnel.StreamData
	(*SendDataRequest)(nil),   // 2: tunnel.SendDataRequest
	(*SendDataResponse)(nil),  // 3: tunnel.SendDataResponse
}
var file_pkg_proto_tunnel_proto_depIdxs = []int32{
	0, // 0: tunnel.Tunnel.InitTunnel:input_type -> tunnel.InitTunnelRequest
	2, // 1: tunnel.Tunnel.SendData:input_type -> tunnel.SendDataRequest
	1, // 2: tunnel.Tunnel.InitTunnel:output_type -> tunnel.StreamData
	3, // 3: tunnel.Tunnel.SendData:output_type -> tunnel.SendDataResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_proto_tunnel_proto_init() }
func file_pkg_proto_tunnel_proto_init() {
	if File_pkg_proto_tunnel_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_proto_tunnel_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InitTunnelRequest); i {
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
		file_pkg_proto_tunnel_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamData); i {
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
		file_pkg_proto_tunnel_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendDataRequest); i {
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
		file_pkg_proto_tunnel_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendDataResponse); i {
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
			RawDescriptor: file_pkg_proto_tunnel_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_proto_tunnel_proto_goTypes,
		DependencyIndexes: file_pkg_proto_tunnel_proto_depIdxs,
		MessageInfos:      file_pkg_proto_tunnel_proto_msgTypes,
	}.Build()
	File_pkg_proto_tunnel_proto = out.File
	file_pkg_proto_tunnel_proto_rawDesc = nil
	file_pkg_proto_tunnel_proto_goTypes = nil
	file_pkg_proto_tunnel_proto_depIdxs = nil
}
