// Copyright (C) 2024  mieru authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.22.3
// source: history.proto

package updaterpb

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

type UpdateHistory struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Records []*UpdateRecord `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
}

func (x *UpdateHistory) Reset() {
	*x = UpdateHistory{}
	if protoimpl.UnsafeEnabled {
		mi := &file_history_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateHistory) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateHistory) ProtoMessage() {}

func (x *UpdateHistory) ProtoReflect() protoreflect.Message {
	mi := &file_history_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateHistory.ProtoReflect.Descriptor instead.
func (*UpdateHistory) Descriptor() ([]byte, []int) {
	return file_history_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateHistory) GetRecords() []*UpdateRecord {
	if x != nil {
		return x.Records
	}
	return nil
}

type UpdateRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Time in UNIX second when this record is created.
	TimeUnix *int64 `protobuf:"varint,1,opt,name=timeUnix,proto3,oneof" json:"timeUnix,omitempty"`
	// The software version that runs the check update.
	Version *string `protobuf:"bytes,2,opt,name=version,proto3,oneof" json:"version,omitempty"`
	// The latest software version, if found.
	LatestVersion *string `protobuf:"bytes,3,opt,name=latestVersion,proto3,oneof" json:"latestVersion,omitempty"`
	// If a new release is found.
	NewReleaseFound *bool `protobuf:"varint,4,opt,name=newReleaseFound,proto3,oneof" json:"newReleaseFound,omitempty"`
	// An error message when check update fail.
	Error *string `protobuf:"bytes,5,opt,name=error,proto3,oneof" json:"error,omitempty"`
}

func (x *UpdateRecord) Reset() {
	*x = UpdateRecord{}
	if protoimpl.UnsafeEnabled {
		mi := &file_history_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateRecord) ProtoMessage() {}

func (x *UpdateRecord) ProtoReflect() protoreflect.Message {
	mi := &file_history_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateRecord.ProtoReflect.Descriptor instead.
func (*UpdateRecord) Descriptor() ([]byte, []int) {
	return file_history_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateRecord) GetTimeUnix() int64 {
	if x != nil && x.TimeUnix != nil {
		return *x.TimeUnix
	}
	return 0
}

func (x *UpdateRecord) GetVersion() string {
	if x != nil && x.Version != nil {
		return *x.Version
	}
	return ""
}

func (x *UpdateRecord) GetLatestVersion() string {
	if x != nil && x.LatestVersion != nil {
		return *x.LatestVersion
	}
	return ""
}

func (x *UpdateRecord) GetNewReleaseFound() bool {
	if x != nil && x.NewReleaseFound != nil {
		return *x.NewReleaseFound
	}
	return false
}

func (x *UpdateRecord) GetError() string {
	if x != nil && x.Error != nil {
		return *x.Error
	}
	return ""
}

var File_history_proto protoreflect.FileDescriptor

var file_history_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x68, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x72, 0x22, 0x40, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x2f, 0x0a, 0x07, 0x72, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x72, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x52, 0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x22, 0x8c, 0x02, 0x0a, 0x0c, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x1f, 0x0a, 0x08, 0x74,
	0x69, 0x6d, 0x65, 0x55, 0x6e, 0x69, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x48, 0x00, 0x52,
	0x08, 0x74, 0x69, 0x6d, 0x65, 0x55, 0x6e, 0x69, 0x78, 0x88, 0x01, 0x01, 0x12, 0x1d, 0x0a, 0x07,
	0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x01, 0x52,
	0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x29, 0x0a, 0x0d, 0x6c,
	0x61, 0x74, 0x65, 0x73, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x02, 0x52, 0x0d, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x56, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x88, 0x01, 0x01, 0x12, 0x2d, 0x0a, 0x0f, 0x6e, 0x65, 0x77, 0x52, 0x65, 0x6c,
	0x65, 0x61, 0x73, 0x65, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x48,
	0x03, 0x52, 0x0f, 0x6e, 0x65, 0x77, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x46, 0x6f, 0x75,
	0x6e, 0x64, 0x88, 0x01, 0x01, 0x12, 0x19, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x48, 0x04, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x88, 0x01, 0x01,
	0x42, 0x0b, 0x0a, 0x09, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x55, 0x6e, 0x69, 0x78, 0x42, 0x0a, 0x0a,
	0x08, 0x5f, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x10, 0x0a, 0x0e, 0x5f, 0x6c, 0x61,
	0x74, 0x65, 0x73, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x42, 0x12, 0x0a, 0x10, 0x5f,
	0x6e, 0x65, 0x77, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x46, 0x6f, 0x75, 0x6e, 0x64, 0x42,
	0x08, 0x0a, 0x06, 0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x42, 0x3a, 0x5a, 0x38, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x66, 0x65, 0x69, 0x6e, 0x2f, 0x6d,
	0x69, 0x65, 0x72, 0x75, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x76, 0x65, 0x72, 0x73,
	0x69, 0x6f, 0x6e, 0x2f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x72, 0x2f, 0x75, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_history_proto_rawDescOnce sync.Once
	file_history_proto_rawDescData = file_history_proto_rawDesc
)

func file_history_proto_rawDescGZIP() []byte {
	file_history_proto_rawDescOnce.Do(func() {
		file_history_proto_rawDescData = protoimpl.X.CompressGZIP(file_history_proto_rawDescData)
	})
	return file_history_proto_rawDescData
}

var file_history_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_history_proto_goTypes = []interface{}{
	(*UpdateHistory)(nil), // 0: updater.UpdateHistory
	(*UpdateRecord)(nil),  // 1: updater.UpdateRecord
}
var file_history_proto_depIdxs = []int32{
	1, // 0: updater.UpdateHistory.records:type_name -> updater.UpdateRecord
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_history_proto_init() }
func file_history_proto_init() {
	if File_history_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_history_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateHistory); i {
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
		file_history_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateRecord); i {
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
	file_history_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_history_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_history_proto_goTypes,
		DependencyIndexes: file_history_proto_depIdxs,
		MessageInfos:      file_history_proto_msgTypes,
	}.Build()
	File_history_proto = out.File
	file_history_proto_rawDesc = nil
	file_history_proto_goTypes = nil
	file_history_proto_depIdxs = nil
}