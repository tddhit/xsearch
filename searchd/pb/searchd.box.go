// Code generated by protoc-gen-box. DO NOT EDIT.
// source: searchd.proto

package searchdpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import pb "github.com/tddhit/xsearch/pb"

import (
	tr "github.com/tddhit/box/transport"
	tropt "github.com/tddhit/box/transport/option"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import (
	context1 "golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Command_Type int32

const (
	Command_INDEX  Command_Type = 0
	Command_REMOVE Command_Type = 1
)

var Command_Type_name = map[int32]string{
	0: "INDEX",
	1: "REMOVE",
}
var Command_Type_value = map[string]int32{
	"INDEX":  0,
	"REMOVE": 1,
}

func (x Command_Type) String() string {
	return proto.EnumName(Command_Type_name, int32(x))
}
func (Command_Type) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{10, 0}
}

type CreateShardReq struct {
	ShardID              string   `protobuf:"bytes,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateShardReq) Reset()         { *m = CreateShardReq{} }
func (m *CreateShardReq) String() string { return proto.CompactTextString(m) }
func (*CreateShardReq) ProtoMessage()    {}
func (*CreateShardReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{0}
}
func (m *CreateShardReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateShardReq.Unmarshal(m, b)
}
func (m *CreateShardReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateShardReq.Marshal(b, m, deterministic)
}
func (dst *CreateShardReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateShardReq.Merge(dst, src)
}
func (m *CreateShardReq) XXX_Size() int {
	return xxx_messageInfo_CreateShardReq.Size(m)
}
func (m *CreateShardReq) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateShardReq.DiscardUnknown(m)
}

var xxx_messageInfo_CreateShardReq proto.InternalMessageInfo

func (m *CreateShardReq) GetShardID() string {
	if m != nil {
		return m.ShardID
	}
	return ""
}

type CreateShardRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateShardRsp) Reset()         { *m = CreateShardRsp{} }
func (m *CreateShardRsp) String() string { return proto.CompactTextString(m) }
func (*CreateShardRsp) ProtoMessage()    {}
func (*CreateShardRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{1}
}
func (m *CreateShardRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateShardRsp.Unmarshal(m, b)
}
func (m *CreateShardRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateShardRsp.Marshal(b, m, deterministic)
}
func (dst *CreateShardRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateShardRsp.Merge(dst, src)
}
func (m *CreateShardRsp) XXX_Size() int {
	return xxx_messageInfo_CreateShardRsp.Size(m)
}
func (m *CreateShardRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateShardRsp.DiscardUnknown(m)
}

var xxx_messageInfo_CreateShardRsp proto.InternalMessageInfo

type RemoveShardReq struct {
	ShardID              string   `protobuf:"bytes,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveShardReq) Reset()         { *m = RemoveShardReq{} }
func (m *RemoveShardReq) String() string { return proto.CompactTextString(m) }
func (*RemoveShardReq) ProtoMessage()    {}
func (*RemoveShardReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{2}
}
func (m *RemoveShardReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveShardReq.Unmarshal(m, b)
}
func (m *RemoveShardReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveShardReq.Marshal(b, m, deterministic)
}
func (dst *RemoveShardReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveShardReq.Merge(dst, src)
}
func (m *RemoveShardReq) XXX_Size() int {
	return xxx_messageInfo_RemoveShardReq.Size(m)
}
func (m *RemoveShardReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveShardReq.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveShardReq proto.InternalMessageInfo

func (m *RemoveShardReq) GetShardID() string {
	if m != nil {
		return m.ShardID
	}
	return ""
}

type RemoveShardRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveShardRsp) Reset()         { *m = RemoveShardRsp{} }
func (m *RemoveShardRsp) String() string { return proto.CompactTextString(m) }
func (*RemoveShardRsp) ProtoMessage()    {}
func (*RemoveShardRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{3}
}
func (m *RemoveShardRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveShardRsp.Unmarshal(m, b)
}
func (m *RemoveShardRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveShardRsp.Marshal(b, m, deterministic)
}
func (dst *RemoveShardRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveShardRsp.Merge(dst, src)
}
func (m *RemoveShardRsp) XXX_Size() int {
	return xxx_messageInfo_RemoveShardRsp.Size(m)
}
func (m *RemoveShardRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveShardRsp.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveShardRsp proto.InternalMessageInfo

type IndexDocReq struct {
	ShardID              string       `protobuf:"bytes,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	Doc                  *pb.Document `protobuf:"bytes,2,opt,name=doc,proto3" json:"doc,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-"`
	XXX_unrecognized     []byte       `json:"-"`
	XXX_sizecache        int32        `json:"-"`
}

func (m *IndexDocReq) Reset()         { *m = IndexDocReq{} }
func (m *IndexDocReq) String() string { return proto.CompactTextString(m) }
func (*IndexDocReq) ProtoMessage()    {}
func (*IndexDocReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{4}
}
func (m *IndexDocReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IndexDocReq.Unmarshal(m, b)
}
func (m *IndexDocReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IndexDocReq.Marshal(b, m, deterministic)
}
func (dst *IndexDocReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexDocReq.Merge(dst, src)
}
func (m *IndexDocReq) XXX_Size() int {
	return xxx_messageInfo_IndexDocReq.Size(m)
}
func (m *IndexDocReq) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexDocReq.DiscardUnknown(m)
}

var xxx_messageInfo_IndexDocReq proto.InternalMessageInfo

func (m *IndexDocReq) GetShardID() string {
	if m != nil {
		return m.ShardID
	}
	return ""
}

func (m *IndexDocReq) GetDoc() *pb.Document {
	if m != nil {
		return m.Doc
	}
	return nil
}

type IndexDocRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IndexDocRsp) Reset()         { *m = IndexDocRsp{} }
func (m *IndexDocRsp) String() string { return proto.CompactTextString(m) }
func (*IndexDocRsp) ProtoMessage()    {}
func (*IndexDocRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{5}
}
func (m *IndexDocRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IndexDocRsp.Unmarshal(m, b)
}
func (m *IndexDocRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IndexDocRsp.Marshal(b, m, deterministic)
}
func (dst *IndexDocRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IndexDocRsp.Merge(dst, src)
}
func (m *IndexDocRsp) XXX_Size() int {
	return xxx_messageInfo_IndexDocRsp.Size(m)
}
func (m *IndexDocRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_IndexDocRsp.DiscardUnknown(m)
}

var xxx_messageInfo_IndexDocRsp proto.InternalMessageInfo

type RemoveDocReq struct {
	ShardID              string   `protobuf:"bytes,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	DocID                string   `protobuf:"bytes,2,opt,name=docID,proto3" json:"docID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveDocReq) Reset()         { *m = RemoveDocReq{} }
func (m *RemoveDocReq) String() string { return proto.CompactTextString(m) }
func (*RemoveDocReq) ProtoMessage()    {}
func (*RemoveDocReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{6}
}
func (m *RemoveDocReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveDocReq.Unmarshal(m, b)
}
func (m *RemoveDocReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveDocReq.Marshal(b, m, deterministic)
}
func (dst *RemoveDocReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveDocReq.Merge(dst, src)
}
func (m *RemoveDocReq) XXX_Size() int {
	return xxx_messageInfo_RemoveDocReq.Size(m)
}
func (m *RemoveDocReq) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveDocReq.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveDocReq proto.InternalMessageInfo

func (m *RemoveDocReq) GetShardID() string {
	if m != nil {
		return m.ShardID
	}
	return ""
}

func (m *RemoveDocReq) GetDocID() string {
	if m != nil {
		return m.DocID
	}
	return ""
}

type RemoveDocRsp struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoveDocRsp) Reset()         { *m = RemoveDocRsp{} }
func (m *RemoveDocRsp) String() string { return proto.CompactTextString(m) }
func (*RemoveDocRsp) ProtoMessage()    {}
func (*RemoveDocRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{7}
}
func (m *RemoveDocRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoveDocRsp.Unmarshal(m, b)
}
func (m *RemoveDocRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoveDocRsp.Marshal(b, m, deterministic)
}
func (dst *RemoveDocRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoveDocRsp.Merge(dst, src)
}
func (m *RemoveDocRsp) XXX_Size() int {
	return xxx_messageInfo_RemoveDocRsp.Size(m)
}
func (m *RemoveDocRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoveDocRsp.DiscardUnknown(m)
}

var xxx_messageInfo_RemoveDocRsp proto.InternalMessageInfo

type SearchReq struct {
	ShardID              string    `protobuf:"bytes,1,opt,name=shardID,proto3" json:"shardID,omitempty"`
	Query                *pb.Query `protobuf:"bytes,2,opt,name=query,proto3" json:"query,omitempty"`
	Start                uint64    `protobuf:"varint,3,opt,name=start,proto3" json:"start,omitempty"`
	Count                uint32    `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *SearchReq) Reset()         { *m = SearchReq{} }
func (m *SearchReq) String() string { return proto.CompactTextString(m) }
func (*SearchReq) ProtoMessage()    {}
func (*SearchReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{8}
}
func (m *SearchReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchReq.Unmarshal(m, b)
}
func (m *SearchReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchReq.Marshal(b, m, deterministic)
}
func (dst *SearchReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchReq.Merge(dst, src)
}
func (m *SearchReq) XXX_Size() int {
	return xxx_messageInfo_SearchReq.Size(m)
}
func (m *SearchReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchReq.DiscardUnknown(m)
}

var xxx_messageInfo_SearchReq proto.InternalMessageInfo

func (m *SearchReq) GetShardID() string {
	if m != nil {
		return m.ShardID
	}
	return ""
}

func (m *SearchReq) GetQuery() *pb.Query {
	if m != nil {
		return m.Query
	}
	return nil
}

func (m *SearchReq) GetStart() uint64 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *SearchReq) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

type SearchRsp struct {
	Docs                 []*pb.Document `protobuf:"bytes,1,rep,name=docs,proto3" json:"docs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *SearchRsp) Reset()         { *m = SearchRsp{} }
func (m *SearchRsp) String() string { return proto.CompactTextString(m) }
func (*SearchRsp) ProtoMessage()    {}
func (*SearchRsp) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{9}
}
func (m *SearchRsp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SearchRsp.Unmarshal(m, b)
}
func (m *SearchRsp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SearchRsp.Marshal(b, m, deterministic)
}
func (dst *SearchRsp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SearchRsp.Merge(dst, src)
}
func (m *SearchRsp) XXX_Size() int {
	return xxx_messageInfo_SearchRsp.Size(m)
}
func (m *SearchRsp) XXX_DiscardUnknown() {
	xxx_messageInfo_SearchRsp.DiscardUnknown(m)
}

var xxx_messageInfo_SearchRsp proto.InternalMessageInfo

func (m *SearchRsp) GetDocs() []*pb.Document {
	if m != nil {
		return m.Docs
	}
	return nil
}

type Command struct {
	Type Command_Type `protobuf:"varint,1,opt,name=type,proto3,enum=searchdpb.Command_Type" json:"type,omitempty"`
	// Types that are valid to be assigned to DocOneof:
	//	*Command_Doc
	//	*Command_DocID
	DocOneof             isCommand_DocOneof `protobuf_oneof:"doc_oneof"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *Command) Reset()         { *m = Command{} }
func (m *Command) String() string { return proto.CompactTextString(m) }
func (*Command) ProtoMessage()    {}
func (*Command) Descriptor() ([]byte, []int) {
	return fileDescriptor_searchd_7e33600c41f609fc, []int{10}
}
func (m *Command) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Command.Unmarshal(m, b)
}
func (m *Command) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Command.Marshal(b, m, deterministic)
}
func (dst *Command) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Command.Merge(dst, src)
}
func (m *Command) XXX_Size() int {
	return xxx_messageInfo_Command.Size(m)
}
func (m *Command) XXX_DiscardUnknown() {
	xxx_messageInfo_Command.DiscardUnknown(m)
}

var xxx_messageInfo_Command proto.InternalMessageInfo

type isCommand_DocOneof interface {
	isCommand_DocOneof()
}

type Command_Doc struct {
	Doc *pb.Document `protobuf:"bytes,3,opt,name=doc,proto3,oneof"`
}
type Command_DocID struct {
	DocID string `protobuf:"bytes,4,opt,name=docID,proto3,oneof"`
}

func (*Command_Doc) isCommand_DocOneof()   {}
func (*Command_DocID) isCommand_DocOneof() {}

func (m *Command) GetDocOneof() isCommand_DocOneof {
	if m != nil {
		return m.DocOneof
	}
	return nil
}

func (m *Command) GetType() Command_Type {
	if m != nil {
		return m.Type
	}
	return Command_INDEX
}

func (m *Command) GetDoc() *pb.Document {
	if x, ok := m.GetDocOneof().(*Command_Doc); ok {
		return x.Doc
	}
	return nil
}

func (m *Command) GetDocID() string {
	if x, ok := m.GetDocOneof().(*Command_DocID); ok {
		return x.DocID
	}
	return ""
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Command) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Command_OneofMarshaler, _Command_OneofUnmarshaler, _Command_OneofSizer, []interface{}{
		(*Command_Doc)(nil),
		(*Command_DocID)(nil),
	}
}

func _Command_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Command)
	// doc_oneof
	switch x := m.DocOneof.(type) {
	case *Command_Doc:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Doc); err != nil {
			return err
		}
	case *Command_DocID:
		b.EncodeVarint(4<<3 | proto.WireBytes)
		b.EncodeStringBytes(x.DocID)
	case nil:
	default:
		return fmt.Errorf("Command.DocOneof has unexpected type %T", x)
	}
	return nil
}

func _Command_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Command)
	switch tag {
	case 3: // doc_oneof.doc
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(pb.Document)
		err := b.DecodeMessage(msg)
		m.DocOneof = &Command_Doc{msg}
		return true, err
	case 4: // doc_oneof.docID
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeStringBytes()
		m.DocOneof = &Command_DocID{x}
		return true, err
	default:
		return false, nil
	}
}

func _Command_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Command)
	// doc_oneof
	switch x := m.DocOneof.(type) {
	case *Command_Doc:
		s := proto.Size(x.Doc)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Command_DocID:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(len(x.DocID)))
		n += len(x.DocID)
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

func init() {
	proto.RegisterType((*CreateShardReq)(nil), "searchdpb.CreateShardReq")
	proto.RegisterType((*CreateShardRsp)(nil), "searchdpb.CreateShardRsp")
	proto.RegisterType((*RemoveShardReq)(nil), "searchdpb.RemoveShardReq")
	proto.RegisterType((*RemoveShardRsp)(nil), "searchdpb.RemoveShardRsp")
	proto.RegisterType((*IndexDocReq)(nil), "searchdpb.IndexDocReq")
	proto.RegisterType((*IndexDocRsp)(nil), "searchdpb.IndexDocRsp")
	proto.RegisterType((*RemoveDocReq)(nil), "searchdpb.RemoveDocReq")
	proto.RegisterType((*RemoveDocRsp)(nil), "searchdpb.RemoveDocRsp")
	proto.RegisterType((*SearchReq)(nil), "searchdpb.SearchReq")
	proto.RegisterType((*SearchRsp)(nil), "searchdpb.SearchRsp")
	proto.RegisterType((*Command)(nil), "searchdpb.Command")
	proto.RegisterEnum("searchdpb.Command_Type", Command_Type_name, Command_Type_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ tr.Server
var _ tr.ClientConn
var _ tropt.CallOption

type SearchdAdminGrpcClient interface {
	CreateShard(ctx context.Context, in *CreateShardReq, opts ...tropt.CallOption) (*CreateShardRsp, error)
	RemoveShard(ctx context.Context, in *RemoveShardReq, opts ...tropt.CallOption) (*RemoveShardRsp, error)
}

type searchdAdminGrpcClient struct {
	cc tr.ClientConn
}

func NewSearchdAdminGrpcClient(cc tr.ClientConn) SearchdAdminGrpcClient {
	return &searchdAdminGrpcClient{cc}
}

func (c *searchdAdminGrpcClient) CreateShard(ctx context.Context, in *CreateShardReq, opts ...tropt.CallOption) (*CreateShardRsp, error) {
	out := new(CreateShardRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.SearchdAdmin/CreateShard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdAdminGrpcClient) RemoveShard(ctx context.Context, in *RemoveShardReq, opts ...tropt.CallOption) (*RemoveShardRsp, error) {
	out := new(RemoveShardRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.SearchdAdmin/RemoveShard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type searchdAdminGrpcServiceDesc struct {
	desc *grpc.ServiceDesc
}

func (d *searchdAdminGrpcServiceDesc) Desc() interface{} {
	return d.desc
}

var SearchdAdminGrpcServiceDesc = &searchdAdminGrpcServiceDesc{&_SearchdAdmin_serviceDesc}

type SearchdGrpcClient interface {
	IndexDoc(ctx context.Context, in *IndexDocReq, opts ...tropt.CallOption) (*IndexDocRsp, error)
	RemoveDoc(ctx context.Context, in *RemoveDocReq, opts ...tropt.CallOption) (*RemoveDocRsp, error)
	Search(ctx context.Context, in *SearchReq, opts ...tropt.CallOption) (*SearchRsp, error)
}

type searchdGrpcClient struct {
	cc tr.ClientConn
}

func NewSearchdGrpcClient(cc tr.ClientConn) SearchdGrpcClient {
	return &searchdGrpcClient{cc}
}

func (c *searchdGrpcClient) IndexDoc(ctx context.Context, in *IndexDocReq, opts ...tropt.CallOption) (*IndexDocRsp, error) {
	out := new(IndexDocRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/IndexDoc", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdGrpcClient) RemoveDoc(ctx context.Context, in *RemoveDocReq, opts ...tropt.CallOption) (*RemoveDocRsp, error) {
	out := new(RemoveDocRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/RemoveDoc", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdGrpcClient) Search(ctx context.Context, in *SearchReq, opts ...tropt.CallOption) (*SearchRsp, error) {
	out := new(SearchRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type searchdGrpcServiceDesc struct {
	desc *grpc.ServiceDesc
}

func (d *searchdGrpcServiceDesc) Desc() interface{} {
	return d.desc
}

var SearchdGrpcServiceDesc = &searchdGrpcServiceDesc{&_Searchd_serviceDesc}

// Reference imports to suppress errors if they are not otherwise used.
var _ context1.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// SearchdAdminClient is the client API for SearchdAdmin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SearchdAdminClient interface {
	CreateShard(ctx context1.Context, in *CreateShardReq, opts ...grpc1.CallOption) (*CreateShardRsp, error)
	RemoveShard(ctx context1.Context, in *RemoveShardReq, opts ...grpc1.CallOption) (*RemoveShardRsp, error)
}

type searchdAdminClient struct {
	cc *grpc1.ClientConn
}

func NewSearchdAdminClient(cc *grpc1.ClientConn) SearchdAdminClient {
	return &searchdAdminClient{cc}
}

func (c *searchdAdminClient) CreateShard(ctx context1.Context, in *CreateShardReq, opts ...grpc1.CallOption) (*CreateShardRsp, error) {
	out := new(CreateShardRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.SearchdAdmin/CreateShard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdAdminClient) RemoveShard(ctx context1.Context, in *RemoveShardReq, opts ...grpc1.CallOption) (*RemoveShardRsp, error) {
	out := new(RemoveShardRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.SearchdAdmin/RemoveShard", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SearchdAdminServer is the server API for SearchdAdmin service.
type SearchdAdminServer interface {
	CreateShard(context1.Context, *CreateShardReq) (*CreateShardRsp, error)
	RemoveShard(context1.Context, *RemoveShardReq) (*RemoveShardRsp, error)
}

func RegisterSearchdAdminServer(s *grpc1.Server, srv SearchdAdminServer) {
	s.RegisterService(&_SearchdAdmin_serviceDesc, srv)
}

func _SearchdAdmin_CreateShard_Handler(srv interface{}, ctx context1.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateShardReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchdAdminServer).CreateShard(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/searchdpb.SearchdAdmin/CreateShard",
	}
	handler := func(ctx context1.Context, req interface{}) (interface{}, error) {
		return srv.(SearchdAdminServer).CreateShard(ctx, req.(*CreateShardReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _SearchdAdmin_RemoveShard_Handler(srv interface{}, ctx context1.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveShardReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchdAdminServer).RemoveShard(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/searchdpb.SearchdAdmin/RemoveShard",
	}
	handler := func(ctx context1.Context, req interface{}) (interface{}, error) {
		return srv.(SearchdAdminServer).RemoveShard(ctx, req.(*RemoveShardReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _SearchdAdmin_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "searchdpb.SearchdAdmin",
	HandlerType: (*SearchdAdminServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "CreateShard",
			Handler:    _SearchdAdmin_CreateShard_Handler,
		},
		{
			MethodName: "RemoveShard",
			Handler:    _SearchdAdmin_RemoveShard_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "searchd.proto",
}

// SearchdClient is the client API for Searchd service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SearchdClient interface {
	IndexDoc(ctx context1.Context, in *IndexDocReq, opts ...grpc1.CallOption) (*IndexDocRsp, error)
	RemoveDoc(ctx context1.Context, in *RemoveDocReq, opts ...grpc1.CallOption) (*RemoveDocRsp, error)
	Search(ctx context1.Context, in *SearchReq, opts ...grpc1.CallOption) (*SearchRsp, error)
}

type searchdClient struct {
	cc *grpc1.ClientConn
}

func NewSearchdClient(cc *grpc1.ClientConn) SearchdClient {
	return &searchdClient{cc}
}

func (c *searchdClient) IndexDoc(ctx context1.Context, in *IndexDocReq, opts ...grpc1.CallOption) (*IndexDocRsp, error) {
	out := new(IndexDocRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/IndexDoc", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdClient) RemoveDoc(ctx context1.Context, in *RemoveDocReq, opts ...grpc1.CallOption) (*RemoveDocRsp, error) {
	out := new(RemoveDocRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/RemoveDoc", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *searchdClient) Search(ctx context1.Context, in *SearchReq, opts ...grpc1.CallOption) (*SearchRsp, error) {
	out := new(SearchRsp)
	err := c.cc.Invoke(ctx, "/searchdpb.Searchd/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SearchdServer is the server API for Searchd service.
type SearchdServer interface {
	IndexDoc(context1.Context, *IndexDocReq) (*IndexDocRsp, error)
	RemoveDoc(context1.Context, *RemoveDocReq) (*RemoveDocRsp, error)
	Search(context1.Context, *SearchReq) (*SearchRsp, error)
}

func RegisterSearchdServer(s *grpc1.Server, srv SearchdServer) {
	s.RegisterService(&_Searchd_serviceDesc, srv)
}

func _Searchd_IndexDoc_Handler(srv interface{}, ctx context1.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(IndexDocReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchdServer).IndexDoc(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/searchdpb.Searchd/IndexDoc",
	}
	handler := func(ctx context1.Context, req interface{}) (interface{}, error) {
		return srv.(SearchdServer).IndexDoc(ctx, req.(*IndexDocReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Searchd_RemoveDoc_Handler(srv interface{}, ctx context1.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveDocReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchdServer).RemoveDoc(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/searchdpb.Searchd/RemoveDoc",
	}
	handler := func(ctx context1.Context, req interface{}) (interface{}, error) {
		return srv.(SearchdServer).RemoveDoc(ctx, req.(*RemoveDocReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Searchd_Search_Handler(srv interface{}, ctx context1.Context, dec func(interface{}) error, interceptor grpc1.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SearchdServer).Search(ctx, in)
	}
	info := &grpc1.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/searchdpb.Searchd/Search",
	}
	handler := func(ctx context1.Context, req interface{}) (interface{}, error) {
		return srv.(SearchdServer).Search(ctx, req.(*SearchReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Searchd_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "searchdpb.Searchd",
	HandlerType: (*SearchdServer)(nil),
	Methods: []grpc1.MethodDesc{
		{
			MethodName: "IndexDoc",
			Handler:    _Searchd_IndexDoc_Handler,
		},
		{
			MethodName: "RemoveDoc",
			Handler:    _Searchd_RemoveDoc_Handler,
		},
		{
			MethodName: "Search",
			Handler:    _Searchd_Search_Handler,
		},
	},
	Streams:  []grpc1.StreamDesc{},
	Metadata: "searchd.proto",
}

func init() { proto.RegisterFile("searchd.proto", fileDescriptor_searchd_7e33600c41f609fc) }

var fileDescriptor_searchd_7e33600c41f609fc = []byte{
	// 467 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xd1, 0x8a, 0xd3, 0x40,
	0x14, 0xed, 0xd8, 0xb4, 0x35, 0xb7, 0xdb, 0x52, 0xc6, 0x65, 0x37, 0x06, 0x84, 0x30, 0xa0, 0x5b,
	0x14, 0xb2, 0x50, 0xc5, 0x27, 0x51, 0x74, 0x53, 0xd8, 0x3e, 0xb8, 0xe2, 0xac, 0x88, 0x6f, 0x92,
	0xce, 0x8c, 0xac, 0x0f, 0xc9, 0xcc, 0x66, 0x52, 0xd9, 0xfe, 0x8b, 0x6f, 0xfe, 0x88, 0x9f, 0x26,
	0x33, 0x93, 0x96, 0x29, 0x36, 0xd5, 0xb7, 0xb9, 0xf7, 0x9e, 0x73, 0xee, 0xc9, 0xe5, 0x04, 0x46,
	0x5a, 0xe4, 0x15, 0xbb, 0xe1, 0xa9, 0xaa, 0x64, 0x2d, 0x71, 0xd8, 0x94, 0x6a, 0x19, 0x47, 0x77,
	0xee, 0x7d, 0xae, 0x96, 0xe7, 0xcd, 0xd3, 0x81, 0xc8, 0x53, 0x18, 0x5f, 0x54, 0x22, 0xaf, 0xc5,
	0xf5, 0x4d, 0x5e, 0x71, 0x2a, 0x6e, 0x71, 0x04, 0x03, 0x6d, 0xde, 0x8b, 0x2c, 0x42, 0x09, 0x9a,
	0x86, 0x74, 0x53, 0x92, 0xc9, 0x2e, 0x56, 0x2b, 0xc3, 0xa6, 0xa2, 0x90, 0x3f, 0xfe, 0x93, 0xed,
	0x63, 0xb5, 0x22, 0x57, 0x30, 0x5c, 0x94, 0x5c, 0xdc, 0x65, 0x92, 0x1d, 0xa4, 0xe2, 0xc7, 0xd0,
	0xe5, 0x92, 0x45, 0xf7, 0x12, 0x34, 0x1d, 0xce, 0x1e, 0xa4, 0xcd, 0x17, 0xa8, 0x65, 0x9a, 0x49,
	0xb6, 0x2a, 0x44, 0x59, 0x53, 0x33, 0x27, 0x23, 0x4f, 0x4f, 0x2b, 0xf2, 0x1a, 0x8e, 0xdc, 0xc2,
	0x7f, 0xea, 0x1f, 0x43, 0x8f, 0x4b, 0xb6, 0xc8, 0xec, 0x86, 0x90, 0xba, 0x82, 0x8c, 0x7d, 0xbe,
	0x56, 0x64, 0x0d, 0xe1, 0xb5, 0x5d, 0x7c, 0x58, 0xec, 0x09, 0xf4, 0x6e, 0x57, 0xa2, 0x5a, 0x37,
	0x76, 0x27, 0x9e, 0xdd, 0x8f, 0xa6, 0x4f, 0xdd, 0xd8, 0x2c, 0xd5, 0x75, 0x5e, 0xd5, 0x51, 0x37,
	0x41, 0xd3, 0x80, 0xba, 0xc2, 0x74, 0x99, 0x5c, 0x95, 0x75, 0x14, 0x24, 0x68, 0x3a, 0xa2, 0xae,
	0x20, 0x2f, 0xb6, 0xab, 0xb5, 0xc2, 0x67, 0x10, 0x70, 0xc9, 0x74, 0x84, 0x92, 0x6e, 0xdb, 0x39,
	0x2c, 0x80, 0xfc, 0x42, 0x30, 0xb8, 0x90, 0x45, 0x91, 0x97, 0x1c, 0x3f, 0x83, 0xa0, 0x5e, 0x2b,
	0x61, 0xcd, 0x8e, 0x67, 0xa7, 0xe9, 0x36, 0x1b, 0x69, 0x83, 0x48, 0x3f, 0xad, 0x95, 0xa0, 0x16,
	0x84, 0xcf, 0xdc, 0xbd, 0xbb, 0xad, 0xf7, 0xbe, 0xec, 0xd8, 0x8b, 0xe3, 0x93, 0xcd, 0xe1, 0x8c,
	0xdb, 0xf0, 0xb2, 0xb3, 0x39, 0xdd, 0x23, 0x08, 0x8c, 0x1c, 0x0e, 0xa1, 0xb7, 0xb8, 0xca, 0xe6,
	0x5f, 0x26, 0x1d, 0x0c, 0xd0, 0xa7, 0xf3, 0xf7, 0x1f, 0x3e, 0xcf, 0x27, 0xe8, 0xdd, 0x10, 0x42,
	0x2e, 0xd9, 0x57, 0x59, 0x0a, 0xf9, 0x6d, 0xf6, 0x13, 0xc1, 0x91, 0xfb, 0x38, 0xfe, 0x96, 0x17,
	0xdf, 0x4b, 0x3c, 0x87, 0xa1, 0x17, 0x33, 0xfc, 0xd0, 0xf7, 0xba, 0x13, 0xd5, 0xb8, 0x6d, 0xa4,
	0x15, 0xe9, 0x18, 0x19, 0x2f, 0x6f, 0x3b, 0x32, 0xbb, 0x99, 0x8d, 0xdb, 0x46, 0x46, 0x66, 0xf6,
	0x1b, 0xc1, 0xa0, 0xb1, 0x87, 0x5f, 0xc1, 0xfd, 0x4d, 0xc0, 0xf0, 0x89, 0x47, 0xf2, 0x52, 0x1c,
	0xef, 0xed, 0x5b, 0x43, 0x6f, 0x20, 0xdc, 0xe6, 0x09, 0x9f, 0xfe, 0xb5, 0xb3, 0xe1, 0xef, 0x1f,
	0x58, 0x81, 0x97, 0xd0, 0x77, 0x4e, 0xf0, 0xb1, 0x07, 0xda, 0x66, 0x32, 0xde, 0xd3, 0x35, 0xbc,
	0x65, 0xdf, 0xfe, 0xea, 0xcf, 0xff, 0x04, 0x00, 0x00, 0xff, 0xff, 0xd5, 0xbd, 0xff, 0xd2, 0x20,
	0x04, 0x00, 0x00,
}
