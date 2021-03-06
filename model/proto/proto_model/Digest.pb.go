// Code generated by protoc-gen-go. DO NOT EDIT.
// source: Digest.proto

package protobuf

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

//RFID采集记录
type DigestRecord struct {
	//记录标识
	RecordId string `protobuf:"bytes,1,opt,name=recordId,proto3" json:"recordId,omitempty"`
	//数据大类型
	DataCategory string `protobuf:"bytes,2,opt,name=dataCategory,proto3" json:"dataCategory,omitempty"`
	//数据类型
	DataType string `protobuf:"bytes,3,opt,name=dataType,proto3" json:"dataType,omitempty"`
	//资源ID
	ResourceId string `protobuf:"bytes,4,opt,name=resourceId,proto3" json:"resourceId,omitempty"`
	//来源平台ID
	SourcePlatformId string `protobuf:"bytes,5,opt,name=sourcePlatformId,proto3" json:"sourcePlatformId,omitempty"`
	//目标平台ID
	TargetPlatformId string `protobuf:"bytes,6,opt,name=targetPlatformId,proto3" json:"targetPlatformId,omitempty"`
	//数据时间
	EventTime int64 `protobuf:"varint,7,opt,name=eventTime,proto3" json:"eventTime,omitempty"`
	//接入时间
	AccessTime int64 `protobuf:"varint,8,opt,name=accessTime,proto3" json:"accessTime,omitempty"`
	//转出时间
	TransmitTime         int64    `protobuf:"varint,9,opt,name=transmitTime,proto3" json:"transmitTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DigestRecord) Reset()         { *m = DigestRecord{} }
func (m *DigestRecord) String() string { return proto.CompactTextString(m) }
func (*DigestRecord) ProtoMessage()    {}
func (*DigestRecord) Descriptor() ([]byte, []int) {
	return fileDescriptor_7b130f390c0aac61, []int{0}
}

func (m *DigestRecord) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DigestRecord.Unmarshal(m, b)
}
func (m *DigestRecord) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DigestRecord.Marshal(b, m, deterministic)
}
func (m *DigestRecord) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DigestRecord.Merge(m, src)
}
func (m *DigestRecord) XXX_Size() int {
	return xxx_messageInfo_DigestRecord.Size(m)
}
func (m *DigestRecord) XXX_DiscardUnknown() {
	xxx_messageInfo_DigestRecord.DiscardUnknown(m)
}

var xxx_messageInfo_DigestRecord proto.InternalMessageInfo

func (m *DigestRecord) GetRecordId() string {
	if m != nil {
		return m.RecordId
	}
	return ""
}

func (m *DigestRecord) GetDataCategory() string {
	if m != nil {
		return m.DataCategory
	}
	return ""
}

func (m *DigestRecord) GetDataType() string {
	if m != nil {
		return m.DataType
	}
	return ""
}

func (m *DigestRecord) GetResourceId() string {
	if m != nil {
		return m.ResourceId
	}
	return ""
}

func (m *DigestRecord) GetSourcePlatformId() string {
	if m != nil {
		return m.SourcePlatformId
	}
	return ""
}

func (m *DigestRecord) GetTargetPlatformId() string {
	if m != nil {
		return m.TargetPlatformId
	}
	return ""
}

func (m *DigestRecord) GetEventTime() int64 {
	if m != nil {
		return m.EventTime
	}
	return 0
}

func (m *DigestRecord) GetAccessTime() int64 {
	if m != nil {
		return m.AccessTime
	}
	return 0
}

func (m *DigestRecord) GetTransmitTime() int64 {
	if m != nil {
		return m.TransmitTime
	}
	return 0
}

//RFID采集记录列表
type DigestRecordList struct {
	RecordList           []*DigestRecord `protobuf:"bytes,1,rep,name=recordList,proto3" json:"recordList,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *DigestRecordList) Reset()         { *m = DigestRecordList{} }
func (m *DigestRecordList) String() string { return proto.CompactTextString(m) }
func (*DigestRecordList) ProtoMessage()    {}
func (*DigestRecordList) Descriptor() ([]byte, []int) {
	return fileDescriptor_7b130f390c0aac61, []int{1}
}

func (m *DigestRecordList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DigestRecordList.Unmarshal(m, b)
}
func (m *DigestRecordList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DigestRecordList.Marshal(b, m, deterministic)
}
func (m *DigestRecordList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DigestRecordList.Merge(m, src)
}
func (m *DigestRecordList) XXX_Size() int {
	return xxx_messageInfo_DigestRecordList.Size(m)
}
func (m *DigestRecordList) XXX_DiscardUnknown() {
	xxx_messageInfo_DigestRecordList.DiscardUnknown(m)
}

var xxx_messageInfo_DigestRecordList proto.InternalMessageInfo

func (m *DigestRecordList) GetRecordList() []*DigestRecord {
	if m != nil {
		return m.RecordList
	}
	return nil
}

func init() {
	proto.RegisterType((*DigestRecord)(nil), "protobuf.DigestRecord")
	proto.RegisterType((*DigestRecordList)(nil), "protobuf.DigestRecordList")
}

func init() { proto.RegisterFile("Digest.proto", fileDescriptor_7b130f390c0aac61) }

var fileDescriptor_7b130f390c0aac61 = []byte{
	// 251 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xc1, 0x4a, 0xc4, 0x30,
	0x10, 0x86, 0xe9, 0x56, 0xd7, 0x36, 0xee, 0x61, 0xc9, 0x41, 0x82, 0x88, 0x94, 0x9e, 0x8a, 0x87,
	0x1e, 0x14, 0x7c, 0x01, 0xbd, 0x54, 0x3c, 0x48, 0xd9, 0x17, 0xc8, 0x36, 0xb3, 0xa5, 0x60, 0x37,
	0xcb, 0x64, 0x56, 0xd8, 0xf7, 0xf3, 0xc1, 0x24, 0x53, 0xa2, 0xa9, 0x7b, 0x6a, 0xfe, 0xef, 0xff,
	0x52, 0x32, 0x23, 0x56, 0xaf, 0x43, 0x0f, 0x8e, 0xea, 0x03, 0x5a, 0xb2, 0x32, 0xe3, 0xcf, 0xf6,
	0xb8, 0x2b, 0xbf, 0x17, 0xa1, 0x6a, 0xa1, 0xb3, 0x68, 0xe4, 0xad, 0xc8, 0x90, 0x4f, 0x8d, 0x51,
	0x49, 0x91, 0x54, 0x79, 0xfb, 0x9b, 0x65, 0x29, 0x56, 0x46, 0x93, 0x7e, 0xd1, 0x04, 0xbd, 0xc5,
	0x93, 0x5a, 0x70, 0x3f, 0x63, 0xfe, 0xbe, 0xcf, 0x9b, 0xd3, 0x01, 0x54, 0x3a, 0xdd, 0x0f, 0x59,
	0xde, 0x0b, 0x81, 0xe0, 0xec, 0x11, 0x3b, 0x68, 0x8c, 0xba, 0xe0, 0x36, 0x22, 0xf2, 0x41, 0xac,
	0xa7, 0xf3, 0xc7, 0xa7, 0xa6, 0x9d, 0xc5, 0xb1, 0x31, 0xea, 0x92, 0xad, 0x33, 0xee, 0x5d, 0xd2,
	0xd8, 0x03, 0x45, 0xee, 0x72, 0x72, 0xff, 0x73, 0x79, 0x27, 0x72, 0xf8, 0x82, 0x3d, 0x6d, 0x86,
	0x11, 0xd4, 0x55, 0x91, 0x54, 0x69, 0xfb, 0x07, 0xfc, 0xab, 0x74, 0xd7, 0x81, 0x73, 0x5c, 0x67,
	0x5c, 0x47, 0xc4, 0x4f, 0x4d, 0xa8, 0xf7, 0x6e, 0x1c, 0xa6, 0x1f, 0xe4, 0x6c, 0xcc, 0x58, 0xf9,
	0x26, 0xd6, 0xf1, 0x16, 0xdf, 0x07, 0x47, 0xf2, 0xd9, 0x4f, 0x1b, 0x92, 0x4a, 0x8a, 0xb4, 0xba,
	0x7e, 0xbc, 0xa9, 0xc3, 0xe6, 0xeb, 0xd8, 0x6f, 0x23, 0x73, 0xbb, 0x64, 0xe5, 0xe9, 0x27, 0x00,
	0x00, 0xff, 0xff, 0x70, 0x2f, 0x52, 0x7e, 0xb3, 0x01, 0x00, 0x00,
}
