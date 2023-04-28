// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1-devel
// 	protoc        (unknown)
// source: v1/account/service.proto

package account

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type RegisterAccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pubkey string `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
}

func (x *RegisterAccountRequest) Reset() {
	*x = RegisterAccountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_account_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterAccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterAccountRequest) ProtoMessage() {}

func (x *RegisterAccountRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_account_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterAccountRequest.ProtoReflect.Descriptor instead.
func (*RegisterAccountRequest) Descriptor() ([]byte, []int) {
	return file_v1_account_service_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterAccountRequest) GetPubkey() string {
	if x != nil {
		return x.Pubkey
	}
	return ""
}

type RegisterAccountResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RegisterAccountResponse) Reset() {
	*x = RegisterAccountResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_account_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegisterAccountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterAccountResponse) ProtoMessage() {}

func (x *RegisterAccountResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_account_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterAccountResponse.ProtoReflect.Descriptor instead.
func (*RegisterAccountResponse) Descriptor() ([]byte, []int) {
	return file_v1_account_service_proto_rawDescGZIP(), []int{1}
}

type GetPublicKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *GetPublicKeyRequest) Reset() {
	*x = GetPublicKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_account_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPublicKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPublicKeyRequest) ProtoMessage() {}

func (x *GetPublicKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_account_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPublicKeyRequest.ProtoReflect.Descriptor instead.
func (*GetPublicKeyRequest) Descriptor() ([]byte, []int) {
	return file_v1_account_service_proto_rawDescGZIP(), []int{2}
}

func (x *GetPublicKeyRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

type GetPublicKeyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pubkey string `protobuf:"bytes,1,opt,name=pubkey,proto3" json:"pubkey,omitempty"`
}

func (x *GetPublicKeyResponse) Reset() {
	*x = GetPublicKeyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_account_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPublicKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPublicKeyResponse) ProtoMessage() {}

func (x *GetPublicKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_account_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPublicKeyResponse.ProtoReflect.Descriptor instead.
func (*GetPublicKeyResponse) Descriptor() ([]byte, []int) {
	return file_v1_account_service_proto_rawDescGZIP(), []int{3}
}

func (x *GetPublicKeyResponse) GetPubkey() string {
	if x != nil {
		return x.Pubkey
	}
	return ""
}

var File_v1_account_service_proto protoreflect.FileDescriptor

var file_v1_account_service_proto_rawDesc = []byte{
	0x0a, 0x18, 0x76, 0x31, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x76, 0x31, 0x2e, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x30, 0x0a, 0x16, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x70, 0x75, 0x62, 0x6b, 0x65, 0x79, 0x22, 0x19, 0x0a, 0x17, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74,
	0x65, 0x72, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x2f, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x22, 0x2e, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x75,
	0x62, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x75, 0x62, 0x6b,
	0x65, 0x79, 0x32, 0x87, 0x02, 0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x82,
	0x01, 0x0a, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x22, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x2e,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x41, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x26, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x20, 0x3a, 0x01, 0x2a, 0x22, 0x1b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2e,
	0x30, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x12, 0x77, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63,
	0x4b, 0x65, 0x79, 0x12, 0x1f, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x2e, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x76, 0x31, 0x2e, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x2e, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x24, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1e, 0x12, 0x1c,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2e, 0x30, 0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x73, 0x2f, 0x7b, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x7d, 0x42, 0x3a, 0x5a, 0x38,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x62, 0x61, 0x64, 0x61,
	0x2d, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x72, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x79, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x76, 0x31,
	0x2f, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_account_service_proto_rawDescOnce sync.Once
	file_v1_account_service_proto_rawDescData = file_v1_account_service_proto_rawDesc
)

func file_v1_account_service_proto_rawDescGZIP() []byte {
	file_v1_account_service_proto_rawDescOnce.Do(func() {
		file_v1_account_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_account_service_proto_rawDescData)
	})
	return file_v1_account_service_proto_rawDescData
}

var file_v1_account_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_v1_account_service_proto_goTypes = []interface{}{
	(*RegisterAccountRequest)(nil),  // 0: v1.account.RegisterAccountRequest
	(*RegisterAccountResponse)(nil), // 1: v1.account.RegisterAccountResponse
	(*GetPublicKeyRequest)(nil),     // 2: v1.account.GetPublicKeyRequest
	(*GetPublicKeyResponse)(nil),    // 3: v1.account.GetPublicKeyResponse
}
var file_v1_account_service_proto_depIdxs = []int32{
	0, // 0: v1.account.Account.RegisterAccount:input_type -> v1.account.RegisterAccountRequest
	2, // 1: v1.account.Account.GetPublicKey:input_type -> v1.account.GetPublicKeyRequest
	1, // 2: v1.account.Account.RegisterAccount:output_type -> v1.account.RegisterAccountResponse
	3, // 3: v1.account.Account.GetPublicKey:output_type -> v1.account.GetPublicKeyResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_v1_account_service_proto_init() }
func file_v1_account_service_proto_init() {
	if File_v1_account_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_account_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterAccountRequest); i {
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
		file_v1_account_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RegisterAccountResponse); i {
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
		file_v1_account_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPublicKeyRequest); i {
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
		file_v1_account_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPublicKeyResponse); i {
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
			RawDescriptor: file_v1_account_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_account_service_proto_goTypes,
		DependencyIndexes: file_v1_account_service_proto_depIdxs,
		MessageInfos:      file_v1_account_service_proto_msgTypes,
	}.Build()
	File_v1_account_service_proto = out.File
	file_v1_account_service_proto_rawDesc = nil
	file_v1_account_service_proto_goTypes = nil
	file_v1_account_service_proto_depIdxs = nil
}
