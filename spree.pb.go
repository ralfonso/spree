// Code generated by protoc-gen-go.
// source: spree.proto
// DO NOT EDIT!

/*
Package spree is a generated protocol buffer package.

It is generated from these files:
	spree.proto

It has these top-level messages:
	CreateRequest
	CreateResponse
	Shot
	BackendDetails
	ListRequest
	ListResponse
*/
package spree

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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

type CreateRequest struct {
	Filename string `protobuf:"bytes,1,opt,name=filename" json:"filename,omitempty"`
	Offset   int64  `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	Length   int64  `protobuf:"varint,3,opt,name=length" json:"length,omitempty"`
	Data     []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *CreateRequest) Reset()                    { *m = CreateRequest{} }
func (m *CreateRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest) ProtoMessage()               {}
func (*CreateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CreateResponse struct {
	Shot         *Shot `protobuf:"bytes,1,opt,name=shot" json:"shot,omitempty"`
	Offset       int64 `protobuf:"varint,2,opt,name=offset" json:"offset,omitempty"`
	BytesWritten int64 `protobuf:"varint,3,opt,name=bytes_written,json=bytesWritten" json:"bytes_written,omitempty"`
}

func (m *CreateResponse) Reset()                    { *m = CreateResponse{} }
func (m *CreateResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateResponse) ProtoMessage()               {}
func (*CreateResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CreateResponse) GetShot() *Shot {
	if m != nil {
		return m.Shot
	}
	return nil
}

type Shot struct {
	Id        string          `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	CreatedAt string          `protobuf:"bytes,2,opt,name=created_at,json=createdAt" json:"created_at,omitempty"`
	Filename  string          `protobuf:"bytes,3,opt,name=filename" json:"filename,omitempty"`
	Views     uint64          `protobuf:"varint,4,opt,name=views" json:"views,omitempty"`
	Path      string          `protobuf:"bytes,5,opt,name=path" json:"path,omitempty"`
	Backend   *BackendDetails `protobuf:"bytes,6,opt,name=backend" json:"backend,omitempty"`
}

func (m *Shot) Reset()                    { *m = Shot{} }
func (m *Shot) String() string            { return proto.CompactTextString(m) }
func (*Shot) ProtoMessage()               {}
func (*Shot) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Shot) GetBackend() *BackendDetails {
	if m != nil {
		return m.Backend
	}
	return nil
}

type BackendDetails struct {
	Type string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
}

func (m *BackendDetails) Reset()                    { *m = BackendDetails{} }
func (m *BackendDetails) String() string            { return proto.CompactTextString(m) }
func (*BackendDetails) ProtoMessage()               {}
func (*BackendDetails) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type ListRequest struct {
}

func (m *ListRequest) Reset()                    { *m = ListRequest{} }
func (m *ListRequest) String() string            { return proto.CompactTextString(m) }
func (*ListRequest) ProtoMessage()               {}
func (*ListRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type ListResponse struct {
	Shots []*Shot `protobuf:"bytes,1,rep,name=shots" json:"shots,omitempty"`
}

func (m *ListResponse) Reset()                    { *m = ListResponse{} }
func (m *ListResponse) String() string            { return proto.CompactTextString(m) }
func (*ListResponse) ProtoMessage()               {}
func (*ListResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *ListResponse) GetShots() []*Shot {
	if m != nil {
		return m.Shots
	}
	return nil
}

func init() {
	proto.RegisterType((*CreateRequest)(nil), "CreateRequest")
	proto.RegisterType((*CreateResponse)(nil), "CreateResponse")
	proto.RegisterType((*Shot)(nil), "Shot")
	proto.RegisterType((*BackendDetails)(nil), "BackendDetails")
	proto.RegisterType((*ListRequest)(nil), "ListRequest")
	proto.RegisterType((*ListResponse)(nil), "ListResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for Spree service

type SpreeClient interface {
	Create(ctx context.Context, opts ...grpc.CallOption) (Spree_CreateClient, error)
	List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error)
}

type spreeClient struct {
	cc *grpc.ClientConn
}

func NewSpreeClient(cc *grpc.ClientConn) SpreeClient {
	return &spreeClient{cc}
}

func (c *spreeClient) Create(ctx context.Context, opts ...grpc.CallOption) (Spree_CreateClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Spree_serviceDesc.Streams[0], c.cc, "/Spree/Create", opts...)
	if err != nil {
		return nil, err
	}
	x := &spreeCreateClient{stream}
	return x, nil
}

type Spree_CreateClient interface {
	Send(*CreateRequest) error
	Recv() (*CreateResponse, error)
	grpc.ClientStream
}

type spreeCreateClient struct {
	grpc.ClientStream
}

func (x *spreeCreateClient) Send(m *CreateRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *spreeCreateClient) Recv() (*CreateResponse, error) {
	m := new(CreateResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *spreeClient) List(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListResponse, error) {
	out := new(ListResponse)
	err := grpc.Invoke(ctx, "/Spree/List", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Spree service

type SpreeServer interface {
	Create(Spree_CreateServer) error
	List(context.Context, *ListRequest) (*ListResponse, error)
}

func RegisterSpreeServer(s *grpc.Server, srv SpreeServer) {
	s.RegisterService(&_Spree_serviceDesc, srv)
}

func _Spree_Create_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SpreeServer).Create(&spreeCreateServer{stream})
}

type Spree_CreateServer interface {
	Send(*CreateResponse) error
	Recv() (*CreateRequest, error)
	grpc.ServerStream
}

type spreeCreateServer struct {
	grpc.ServerStream
}

func (x *spreeCreateServer) Send(m *CreateResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *spreeCreateServer) Recv() (*CreateRequest, error) {
	m := new(CreateRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Spree_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SpreeServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Spree/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SpreeServer).List(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Spree_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Spree",
	HandlerType: (*SpreeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "List",
			Handler:    _Spree_List_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Create",
			Handler:       _Spree_Create_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("spree.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 357 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0x52, 0x5d, 0x6b, 0xdb, 0x40,
	0x10, 0xf4, 0x59, 0x1f, 0xad, 0xd7, 0x92, 0x0c, 0x47, 0x29, 0xaa, 0x4b, 0x41, 0x5c, 0x5b, 0x50,
	0x29, 0xa8, 0xc5, 0xf9, 0x05, 0xf9, 0x78, 0xcc, 0xd3, 0xf9, 0x21, 0x8f, 0xe6, 0x6c, 0xad, 0x23,
	0x11, 0x47, 0x52, 0x74, 0x9b, 0x18, 0xff, 0xa1, 0xfc, 0xce, 0xa0, 0x3b, 0x39, 0xb1, 0x02, 0x79,
	0xdb, 0x99, 0x15, 0x3b, 0x9a, 0x99, 0x83, 0xa9, 0x6e, 0x5a, 0xc4, 0xac, 0x69, 0x6b, 0xaa, 0x45,
	0x0d, 0xe1, 0x65, 0x8b, 0x8a, 0x50, 0xe2, 0xc3, 0x23, 0x6a, 0xe2, 0x73, 0xf8, 0xbc, 0x2d, 0x77,
	0x58, 0xa9, 0x7b, 0x8c, 0x59, 0xc2, 0xd2, 0x89, 0x7c, 0xc5, 0xfc, 0x2b, 0xf8, 0xf5, 0x76, 0xab,
	0x91, 0xe2, 0x71, 0xc2, 0x52, 0x47, 0xf6, 0xa8, 0xe3, 0x77, 0x58, 0xdd, 0x52, 0x11, 0x3b, 0x96,
	0xb7, 0x88, 0x73, 0x70, 0x73, 0x45, 0x2a, 0x76, 0x13, 0x96, 0x06, 0xd2, 0xcc, 0xa2, 0x80, 0xe8,
	0x28, 0xa8, 0x9b, 0xba, 0xd2, 0xc8, 0xbf, 0x81, 0xab, 0x8b, 0x9a, 0x8c, 0xda, 0x74, 0xe1, 0x65,
	0xcb, 0xa2, 0x26, 0x69, 0xa8, 0x0f, 0x05, 0x7f, 0x42, 0xb8, 0x3e, 0x10, 0xea, 0xd5, 0xbe, 0x2d,
	0x89, 0xb0, 0xea, 0x75, 0x03, 0x43, 0xde, 0x58, 0x4e, 0x3c, 0x33, 0x70, 0xbb, 0x5b, 0x3c, 0x82,
	0x71, 0x99, 0xf7, 0x66, 0xc6, 0x65, 0xce, 0x7f, 0x00, 0x6c, 0xcc, 0x2f, 0xe4, 0x2b, 0x65, 0x2f,
	0x4f, 0xe4, 0xa4, 0x67, 0xce, 0x87, 0x09, 0x38, 0xef, 0x12, 0xf8, 0x02, 0xde, 0x53, 0x89, 0x7b,
	0x6d, 0x2c, 0xb9, 0xd2, 0x82, 0xce, 0x67, 0xa3, 0xa8, 0x88, 0x3d, 0xf3, 0xb5, 0x99, 0xf9, 0x1f,
	0xf8, 0xb4, 0x56, 0x9b, 0x3b, 0xac, 0xf2, 0xd8, 0x37, 0xc6, 0x66, 0xd9, 0x85, 0xc5, 0x57, 0x48,
	0xaa, 0xdc, 0x69, 0x79, 0xdc, 0x8b, 0x5f, 0x10, 0x0d, 0x57, 0xdd, 0x41, 0x3a, 0x34, 0xc7, 0x02,
	0xcc, 0x2c, 0x42, 0x98, 0x5e, 0x97, 0x9a, 0xfa, 0x9e, 0xc4, 0x5f, 0x08, 0x2c, 0xec, 0x53, 0xfc,
	0x0e, 0x5e, 0x17, 0x99, 0x8e, 0x59, 0xe2, 0xbc, 0xc5, 0x68, 0xb9, 0xc5, 0x0a, 0xbc, 0x65, 0x57,
	0x3a, 0xff, 0x07, 0xbe, 0x4d, 0x9f, 0x47, 0xd9, 0xa0, 0xf7, 0xf9, 0x2c, 0x1b, 0xd6, 0x22, 0x46,
	0x29, 0xfb, 0xcf, 0xf8, 0x6f, 0x70, 0x3b, 0x19, 0x1e, 0x64, 0x27, 0xe2, 0xf3, 0x30, 0x3b, 0xd5,
	0x16, 0xa3, 0xb5, 0x6f, 0x5e, 0xd3, 0xd9, 0x4b, 0x00, 0x00, 0x00, 0xff, 0xff, 0x63, 0x28, 0x1a,
	0xa9, 0x5c, 0x02, 0x00, 0x00,
}
