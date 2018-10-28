// Code generated by protoc-gen-box. DO NOT EDIT.
// source: xsearch.proto

package xsearchpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Token struct {
	Term                 string   `protobuf:"bytes,1,opt,name=term,proto3" json:"term,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Token) Reset()         { *m = Token{} }
func (m *Token) String() string { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()    {}
func (*Token) Descriptor() ([]byte, []int) {
	return fileDescriptor_xsearch_08e3ca2bec97f626, []int{0}
}
func (m *Token) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Token.Unmarshal(m, b)
}
func (m *Token) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Token.Marshal(b, m, deterministic)
}
func (dst *Token) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Token.Merge(dst, src)
}
func (m *Token) XXX_Size() int {
	return xxx_messageInfo_Token.Size(m)
}
func (m *Token) XXX_DiscardUnknown() {
	xxx_messageInfo_Token.DiscardUnknown(m)
}

var xxx_messageInfo_Token proto.InternalMessageInfo

func (m *Token) GetTerm() string {
	if m != nil {
		return m.Term
	}
	return ""
}

type Query struct {
	Raw                  string   `protobuf:"bytes,1,opt,name=raw,proto3" json:"raw,omitempty"`
	Tokens               []*Token `protobuf:"bytes,2,rep,name=tokens,proto3" json:"tokens,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}
func (*Query) Descriptor() ([]byte, []int) {
	return fileDescriptor_xsearch_08e3ca2bec97f626, []int{1}
}
func (m *Query) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Query.Unmarshal(m, b)
}
func (m *Query) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Query.Marshal(b, m, deterministic)
}
func (dst *Query) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Query.Merge(dst, src)
}
func (m *Query) XXX_Size() int {
	return xxx_messageInfo_Query.Size(m)
}
func (m *Query) XXX_DiscardUnknown() {
	xxx_messageInfo_Query.DiscardUnknown(m)
}

var xxx_messageInfo_Query proto.InternalMessageInfo

func (m *Query) GetRaw() string {
	if m != nil {
		return m.Raw
	}
	return ""
}

func (m *Query) GetTokens() []*Token {
	if m != nil {
		return m.Tokens
	}
	return nil
}

type Document struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Content              string   `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	Tokens               []*Token `protobuf:"bytes,3,rep,name=tokens,proto3" json:"tokens,omitempty"`
	BM25Score            float32  `protobuf:"fixed32,4,opt,name=BM25Score,proto3" json:"BM25Score,omitempty"`
	SimNetScore          float32  `protobuf:"fixed32,5,opt,name=SimNetScore,proto3" json:"SimNetScore,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Document) Reset()         { *m = Document{} }
func (m *Document) String() string { return proto.CompactTextString(m) }
func (*Document) ProtoMessage()    {}
func (*Document) Descriptor() ([]byte, []int) {
	return fileDescriptor_xsearch_08e3ca2bec97f626, []int{2}
}
func (m *Document) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Document.Unmarshal(m, b)
}
func (m *Document) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Document.Marshal(b, m, deterministic)
}
func (dst *Document) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Document.Merge(dst, src)
}
func (m *Document) XXX_Size() int {
	return xxx_messageInfo_Document.Size(m)
}
func (m *Document) XXX_DiscardUnknown() {
	xxx_messageInfo_Document.DiscardUnknown(m)
}

var xxx_messageInfo_Document proto.InternalMessageInfo

func (m *Document) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Document) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func (m *Document) GetTokens() []*Token {
	if m != nil {
		return m.Tokens
	}
	return nil
}

func (m *Document) GetBM25Score() float32 {
	if m != nil {
		return m.BM25Score
	}
	return 0
}

func (m *Document) GetSimNetScore() float32 {
	if m != nil {
		return m.SimNetScore
	}
	return 0
}

func init() {
	proto.RegisterType((*Token)(nil), "xsearchpb.Token")
	proto.RegisterType((*Query)(nil), "xsearchpb.Query")
	proto.RegisterType((*Document)(nil), "xsearchpb.Document")
}

func init() { proto.RegisterFile("xsearch.proto", fileDescriptor_xsearch_08e3ca2bec97f626) }

var fileDescriptor_xsearch_08e3ca2bec97f626 = []byte{
	// 206 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xad, 0x28, 0x4e, 0x4d,
	0x2c, 0x4a, 0xce, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x72, 0x0b, 0x92, 0x94,
	0xa4, 0xb9, 0x58, 0x43, 0xf2, 0xb3, 0x53, 0xf3, 0x84, 0x84, 0xb8, 0x58, 0x4a, 0x52, 0x8b, 0x72,
	0x25, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0xc0, 0x6c, 0x25, 0x67, 0x2e, 0xd6, 0xc0, 0xd2, 0xd4,
	0xa2, 0x4a, 0x21, 0x01, 0x2e, 0xe6, 0xa2, 0xc4, 0x72, 0xa8, 0x1c, 0x88, 0x29, 0xa4, 0xc1, 0xc5,
	0x56, 0x02, 0xd2, 0x57, 0x2c, 0xc1, 0xa4, 0xc0, 0xac, 0xc1, 0x6d, 0x24, 0xa0, 0x07, 0x37, 0x53,
	0x0f, 0x6c, 0x60, 0x10, 0x54, 0x5e, 0x69, 0x1e, 0x23, 0x17, 0x87, 0x4b, 0x7e, 0x72, 0x69, 0x6e,
	0x6a, 0x5e, 0x89, 0x10, 0x1f, 0x17, 0x93, 0xa7, 0x0b, 0xd4, 0x1c, 0x26, 0x4f, 0x17, 0x21, 0x09,
	0x2e, 0xf6, 0xe4, 0xfc, 0xbc, 0x92, 0xd4, 0xbc, 0x12, 0x09, 0x26, 0xb0, 0x20, 0x8c, 0x8b, 0x64,
	0x01, 0x33, 0x7e, 0x0b, 0x84, 0x64, 0xb8, 0x38, 0x9d, 0x7c, 0x8d, 0x4c, 0x83, 0x93, 0xf3, 0x8b,
	0x52, 0x25, 0x58, 0x14, 0x18, 0x35, 0x98, 0x82, 0x10, 0x02, 0x42, 0x0a, 0x5c, 0xdc, 0xc1, 0x99,
	0xb9, 0x7e, 0xa9, 0x25, 0x10, 0x79, 0x56, 0xb0, 0x3c, 0xb2, 0x50, 0x12, 0x1b, 0x38, 0x50, 0x8c,
	0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xff, 0x23, 0x49, 0x41, 0x25, 0x01, 0x00, 0x00,
}
