// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.19.6
// source: networkfence.proto

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

// NetworkFenceRequest holds the required information to fence/unfence
// the cluster network.
type NetworkFenceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Plugin specific parameters passed in as opaque key-value pairs.
	Parameters map[string]string `protobuf:"bytes,1,rep,name=parameters,proto3" json:"parameters,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Secrets required by the driver to complete the request.
	SecretName      string `protobuf:"bytes,2,opt,name=secret_name,json=secretName,proto3" json:"secret_name,omitempty"`
	SecretNamespace string `protobuf:"bytes,3,opt,name=secret_namespace,json=secretNamespace,proto3" json:"secret_namespace,omitempty"`
	// list of CIDR blocks on which the fencing/unfencing operation is expected
	// to be performed.
	Cidrs []string `protobuf:"bytes,4,rep,name=cidrs,proto3" json:"cidrs,omitempty"`
}

func (x *NetworkFenceRequest) Reset() {
	*x = NetworkFenceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_networkfence_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkFenceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkFenceRequest) ProtoMessage() {}

func (x *NetworkFenceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_networkfence_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkFenceRequest.ProtoReflect.Descriptor instead.
func (*NetworkFenceRequest) Descriptor() ([]byte, []int) {
	return file_networkfence_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkFenceRequest) GetParameters() map[string]string {
	if x != nil {
		return x.Parameters
	}
	return nil
}

func (x *NetworkFenceRequest) GetSecretName() string {
	if x != nil {
		return x.SecretName
	}
	return ""
}

func (x *NetworkFenceRequest) GetSecretNamespace() string {
	if x != nil {
		return x.SecretNamespace
	}
	return ""
}

func (x *NetworkFenceRequest) GetCidrs() []string {
	if x != nil {
		return x.Cidrs
	}
	return nil
}

// NetworkFenceResponse is returned by the CSI-driver as a result of
// the FenceRequest call.
type NetworkFenceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NetworkFenceResponse) Reset() {
	*x = NetworkFenceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_networkfence_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkFenceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkFenceResponse) ProtoMessage() {}

func (x *NetworkFenceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_networkfence_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkFenceResponse.ProtoReflect.Descriptor instead.
func (*NetworkFenceResponse) Descriptor() ([]byte, []int) {
	return file_networkfence_proto_rawDescGZIP(), []int{1}
}

var File_networkfence_proto protoreflect.FileDescriptor

var file_networkfence_proto_rawDesc = []byte{
	0x0a, 0x12, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x65, 0x6e, 0x63, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x82, 0x02, 0x0a, 0x13,
	0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x4a, 0x0a, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x2e, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x52, 0x0a, 0x70, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x12,
	0x1f, 0x0a, 0x0b, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x29, 0x0a, 0x10, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x73, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x63,
	0x69, 0x64, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x05, 0x63, 0x69, 0x64, 0x72,
	0x73, 0x1a, 0x3d, 0x0a, 0x0f, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x22, 0x16, 0x0a, 0x14, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xb4, 0x01, 0x0a, 0x0c, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x50, 0x0a, 0x13, 0x46, 0x65, 0x6e,
	0x63, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x12, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x46, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x15, 0x55,
	0x6e, 0x46, 0x65, 0x6e, 0x63, 0x65, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x12, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x46, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x1b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x46, 0x65, 0x6e, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x3c, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x73,
	0x69, 0x2d, 0x61, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x2f, 0x6b, 0x75, 0x62, 0x65, 0x72, 0x6e, 0x65,
	0x74, 0x65, 0x73, 0x2d, 0x63, 0x73, 0x69, 0x2d, 0x61, 0x64, 0x64, 0x6f, 0x6e, 0x73, 0x2f, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_networkfence_proto_rawDescOnce sync.Once
	file_networkfence_proto_rawDescData = file_networkfence_proto_rawDesc
)

func file_networkfence_proto_rawDescGZIP() []byte {
	file_networkfence_proto_rawDescOnce.Do(func() {
		file_networkfence_proto_rawDescData = protoimpl.X.CompressGZIP(file_networkfence_proto_rawDescData)
	})
	return file_networkfence_proto_rawDescData
}

var file_networkfence_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_networkfence_proto_goTypes = []interface{}{
	(*NetworkFenceRequest)(nil),  // 0: proto.NetworkFenceRequest
	(*NetworkFenceResponse)(nil), // 1: proto.NetworkFenceResponse
	nil,                          // 2: proto.NetworkFenceRequest.ParametersEntry
}
var file_networkfence_proto_depIdxs = []int32{
	2, // 0: proto.NetworkFenceRequest.parameters:type_name -> proto.NetworkFenceRequest.ParametersEntry
	0, // 1: proto.NetworkFence.FenceClusterNetwork:input_type -> proto.NetworkFenceRequest
	0, // 2: proto.NetworkFence.UnFenceClusterNetwork:input_type -> proto.NetworkFenceRequest
	1, // 3: proto.NetworkFence.FenceClusterNetwork:output_type -> proto.NetworkFenceResponse
	1, // 4: proto.NetworkFence.UnFenceClusterNetwork:output_type -> proto.NetworkFenceResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_networkfence_proto_init() }
func file_networkfence_proto_init() {
	if File_networkfence_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_networkfence_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkFenceRequest); i {
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
		file_networkfence_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkFenceResponse); i {
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
			RawDescriptor: file_networkfence_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_networkfence_proto_goTypes,
		DependencyIndexes: file_networkfence_proto_depIdxs,
		MessageInfos:      file_networkfence_proto_msgTypes,
	}.Build()
	File_networkfence_proto = out.File
	file_networkfence_proto_rawDesc = nil
	file_networkfence_proto_goTypes = nil
	file_networkfence_proto_depIdxs = nil
}
