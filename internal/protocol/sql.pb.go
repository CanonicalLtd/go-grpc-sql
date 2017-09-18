// Code generated by protoc-gen-gogo.
// source: internal/protocol/sql.proto
// DO NOT EDIT!

/*
	Package protocol is a generated protocol buffer package.

	It is generated from these files:
		internal/protocol/sql.proto

	It has these top-level messages:
		Request
		Response
		Statement
		Value
		ValueInt64
		ValueFloat64
		ValueBool
		ValueBytes
		ValueString
		ValueTime
		ValueNull
*/
package protocol

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

import io "io"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// ValueCode is a numberic code identifying the Go type of a value.
//
// It supports all types that should be handle by driver.Value in database/sql/driver.
type ValueCode int32

const (
	ValueCode_INT64   ValueCode = 0
	ValueCode_FLOAT64 ValueCode = 1
	ValueCode_BOOL    ValueCode = 2
	ValueCode_BYTES   ValueCode = 3
	ValueCode_STRING  ValueCode = 4
	ValueCode_TIME    ValueCode = 5
	ValueCode_NULL    ValueCode = 6
)

var ValueCode_name = map[int32]string{
	0: "INT64",
	1: "FLOAT64",
	2: "BOOL",
	3: "BYTES",
	4: "STRING",
	5: "TIME",
	6: "NULL",
}
var ValueCode_value = map[string]int32{
	"INT64":   0,
	"FLOAT64": 1,
	"BOOL":    2,
	"BYTES":   3,
	"STRING":  4,
	"TIME":    5,
	"NULL":    6,
}

func (x ValueCode) String() string {
	return proto.EnumName(ValueCode_name, int32(x))
}
func (ValueCode) EnumDescriptor() ([]byte, []int) { return fileDescriptorSql, []int{0} }

type Request struct {
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{0} }

type Response struct {
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{1} }

// Statement received via HTTP POST and meant to be executed by a SQL
// database.
type Statement struct {
	Text string   `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	Args []*Value `protobuf:"bytes,2,rep,name=args" json:"args,omitempty"`
}

func (m *Statement) Reset()                    { *m = Statement{} }
func (m *Statement) String() string            { return proto.CompactTextString(m) }
func (*Statement) ProtoMessage()               {}
func (*Statement) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{2} }

func (m *Statement) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func (m *Statement) GetArgs() []*Value {
	if m != nil {
		return m.Args
	}
	return nil
}

// Value of a single statement argument or row column.
type Value struct {
	Code ValueCode `protobuf:"varint,1,opt,name=code,proto3,enum=protocol.ValueCode" json:"code,omitempty"`
	Data []byte    `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *Value) Reset()                    { *m = Value{} }
func (m *Value) String() string            { return proto.CompactTextString(m) }
func (*Value) ProtoMessage()               {}
func (*Value) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{3} }

func (m *Value) GetCode() ValueCode {
	if m != nil {
		return m.Code
	}
	return ValueCode_INT64
}

func (m *Value) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type ValueInt64 struct {
	Value int64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueInt64) Reset()                    { *m = ValueInt64{} }
func (m *ValueInt64) String() string            { return proto.CompactTextString(m) }
func (*ValueInt64) ProtoMessage()               {}
func (*ValueInt64) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{4} }

func (m *ValueInt64) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

type ValueFloat64 struct {
	Value float64 `protobuf:"fixed64,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueFloat64) Reset()                    { *m = ValueFloat64{} }
func (m *ValueFloat64) String() string            { return proto.CompactTextString(m) }
func (*ValueFloat64) ProtoMessage()               {}
func (*ValueFloat64) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{5} }

func (m *ValueFloat64) GetValue() float64 {
	if m != nil {
		return m.Value
	}
	return 0
}

type ValueBool struct {
	Value bool `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueBool) Reset()                    { *m = ValueBool{} }
func (m *ValueBool) String() string            { return proto.CompactTextString(m) }
func (*ValueBool) ProtoMessage()               {}
func (*ValueBool) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{6} }

func (m *ValueBool) GetValue() bool {
	if m != nil {
		return m.Value
	}
	return false
}

type ValueBytes struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueBytes) Reset()                    { *m = ValueBytes{} }
func (m *ValueBytes) String() string            { return proto.CompactTextString(m) }
func (*ValueBytes) ProtoMessage()               {}
func (*ValueBytes) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{7} }

func (m *ValueBytes) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

type ValueString struct {
	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueString) Reset()                    { *m = ValueString{} }
func (m *ValueString) String() string            { return proto.CompactTextString(m) }
func (*ValueString) ProtoMessage()               {}
func (*ValueString) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{8} }

func (m *ValueString) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type ValueTime struct {
	Value int64 `protobuf:"varint,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ValueTime) Reset()                    { *m = ValueTime{} }
func (m *ValueTime) String() string            { return proto.CompactTextString(m) }
func (*ValueTime) ProtoMessage()               {}
func (*ValueTime) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{9} }

func (m *ValueTime) GetValue() int64 {
	if m != nil {
		return m.Value
	}
	return 0
}

type ValueNull struct {
}

func (m *ValueNull) Reset()                    { *m = ValueNull{} }
func (m *ValueNull) String() string            { return proto.CompactTextString(m) }
func (*ValueNull) ProtoMessage()               {}
func (*ValueNull) Descriptor() ([]byte, []int) { return fileDescriptorSql, []int{10} }

func init() {
	proto.RegisterType((*Request)(nil), "protocol.Request")
	proto.RegisterType((*Response)(nil), "protocol.Response")
	proto.RegisterType((*Statement)(nil), "protocol.Statement")
	proto.RegisterType((*Value)(nil), "protocol.Value")
	proto.RegisterType((*ValueInt64)(nil), "protocol.ValueInt64")
	proto.RegisterType((*ValueFloat64)(nil), "protocol.ValueFloat64")
	proto.RegisterType((*ValueBool)(nil), "protocol.ValueBool")
	proto.RegisterType((*ValueBytes)(nil), "protocol.ValueBytes")
	proto.RegisterType((*ValueString)(nil), "protocol.ValueString")
	proto.RegisterType((*ValueTime)(nil), "protocol.ValueTime")
	proto.RegisterType((*ValueNull)(nil), "protocol.ValueNull")
	proto.RegisterEnum("protocol.ValueCode", ValueCode_name, ValueCode_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SQL service

type SQLClient interface {
	Conn(ctx context.Context, opts ...grpc.CallOption) (SQL_ConnClient, error)
}

type sQLClient struct {
	cc *grpc.ClientConn
}

func NewSQLClient(cc *grpc.ClientConn) SQLClient {
	return &sQLClient{cc}
}

func (c *sQLClient) Conn(ctx context.Context, opts ...grpc.CallOption) (SQL_ConnClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SQL_serviceDesc.Streams[0], c.cc, "/protocol.SQL/Conn", opts...)
	if err != nil {
		return nil, err
	}
	x := &sQLConnClient{stream}
	return x, nil
}

type SQL_ConnClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type sQLConnClient struct {
	grpc.ClientStream
}

func (x *sQLConnClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sQLConnClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for SQL service

type SQLServer interface {
	Conn(SQL_ConnServer) error
}

func RegisterSQLServer(s *grpc.Server, srv SQLServer) {
	s.RegisterService(&_SQL_serviceDesc, srv)
}

func _SQL_Conn_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SQLServer).Conn(&sQLConnServer{stream})
}

type SQL_ConnServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type sQLConnServer struct {
	grpc.ServerStream
}

func (x *sQLConnServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sQLConnServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _SQL_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protocol.SQL",
	HandlerType: (*SQLServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Conn",
			Handler:       _SQL_Conn_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "internal/protocol/sql.proto",
}

func (m *Request) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Request) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *Response) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Response) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func (m *Statement) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Statement) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Text) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSql(dAtA, i, uint64(len(m.Text)))
		i += copy(dAtA[i:], m.Text)
	}
	if len(m.Args) > 0 {
		for _, msg := range m.Args {
			dAtA[i] = 0x12
			i++
			i = encodeVarintSql(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	return i, nil
}

func (m *Value) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Value) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Code != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintSql(dAtA, i, uint64(m.Code))
	}
	if len(m.Data) > 0 {
		dAtA[i] = 0x12
		i++
		i = encodeVarintSql(dAtA, i, uint64(len(m.Data)))
		i += copy(dAtA[i:], m.Data)
	}
	return i, nil
}

func (m *ValueInt64) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueInt64) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintSql(dAtA, i, uint64(m.Value))
	}
	return i, nil
}

func (m *ValueFloat64) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueFloat64) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value != 0 {
		dAtA[i] = 0x9
		i++
		i = encodeFixed64Sql(dAtA, i, uint64(math.Float64bits(float64(m.Value))))
	}
	return i, nil
}

func (m *ValueBool) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueBool) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value {
		dAtA[i] = 0x8
		i++
		if m.Value {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i++
	}
	return i, nil
}

func (m *ValueBytes) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueBytes) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Value) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSql(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func (m *ValueString) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueString) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.Value) > 0 {
		dAtA[i] = 0xa
		i++
		i = encodeVarintSql(dAtA, i, uint64(len(m.Value)))
		i += copy(dAtA[i:], m.Value)
	}
	return i, nil
}

func (m *ValueTime) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueTime) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Value != 0 {
		dAtA[i] = 0x8
		i++
		i = encodeVarintSql(dAtA, i, uint64(m.Value))
	}
	return i, nil
}

func (m *ValueNull) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ValueNull) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	return i, nil
}

func encodeFixed64Sql(dAtA []byte, offset int, v uint64) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	dAtA[offset+4] = uint8(v >> 32)
	dAtA[offset+5] = uint8(v >> 40)
	dAtA[offset+6] = uint8(v >> 48)
	dAtA[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32Sql(dAtA []byte, offset int, v uint32) int {
	dAtA[offset] = uint8(v)
	dAtA[offset+1] = uint8(v >> 8)
	dAtA[offset+2] = uint8(v >> 16)
	dAtA[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintSql(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Request) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *Response) Size() (n int) {
	var l int
	_ = l
	return n
}

func (m *Statement) Size() (n int) {
	var l int
	_ = l
	l = len(m.Text)
	if l > 0 {
		n += 1 + l + sovSql(uint64(l))
	}
	if len(m.Args) > 0 {
		for _, e := range m.Args {
			l = e.Size()
			n += 1 + l + sovSql(uint64(l))
		}
	}
	return n
}

func (m *Value) Size() (n int) {
	var l int
	_ = l
	if m.Code != 0 {
		n += 1 + sovSql(uint64(m.Code))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovSql(uint64(l))
	}
	return n
}

func (m *ValueInt64) Size() (n int) {
	var l int
	_ = l
	if m.Value != 0 {
		n += 1 + sovSql(uint64(m.Value))
	}
	return n
}

func (m *ValueFloat64) Size() (n int) {
	var l int
	_ = l
	if m.Value != 0 {
		n += 9
	}
	return n
}

func (m *ValueBool) Size() (n int) {
	var l int
	_ = l
	if m.Value {
		n += 2
	}
	return n
}

func (m *ValueBytes) Size() (n int) {
	var l int
	_ = l
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovSql(uint64(l))
	}
	return n
}

func (m *ValueString) Size() (n int) {
	var l int
	_ = l
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovSql(uint64(l))
	}
	return n
}

func (m *ValueTime) Size() (n int) {
	var l int
	_ = l
	if m.Value != 0 {
		n += 1 + sovSql(uint64(m.Value))
	}
	return n
}

func (m *ValueNull) Size() (n int) {
	var l int
	_ = l
	return n
}

func sovSql(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozSql(x uint64) (n int) {
	return sovSql(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Request) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Request: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Request: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Response) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Response: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Response: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Statement) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Statement: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Statement: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Text", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSql
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Text = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Args", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthSql
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Args = append(m.Args, &Value{})
			if err := m.Args[len(m.Args)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Value) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Value: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Value: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Code", wireType)
			}
			m.Code = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Code |= (ValueCode(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSql
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueInt64) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueInt64: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueInt64: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			m.Value = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Value |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueFloat64) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueFloat64: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueFloat64: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 1 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var v uint64
			if (iNdEx + 8) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += 8
			v = uint64(dAtA[iNdEx-8])
			v |= uint64(dAtA[iNdEx-7]) << 8
			v |= uint64(dAtA[iNdEx-6]) << 16
			v |= uint64(dAtA[iNdEx-5]) << 24
			v |= uint64(dAtA[iNdEx-4]) << 32
			v |= uint64(dAtA[iNdEx-3]) << 40
			v |= uint64(dAtA[iNdEx-2]) << 48
			v |= uint64(dAtA[iNdEx-1]) << 56
			m.Value = float64(math.Float64frombits(v))
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueBool) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueBool: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueBool: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Value = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueBytes) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueBytes: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueBytes: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthSql
			}
			postIndex := iNdEx + byteLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueString) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueString: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueString: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthSql
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueTime) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueTime: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueTime: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			m.Value = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowSql
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Value |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ValueNull) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowSql
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ValueNull: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ValueNull: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipSql(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthSql
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipSql(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowSql
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSql
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowSql
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthSql
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowSql
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipSql(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthSql = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowSql   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("internal/protocol/sql.proto", fileDescriptorSql) }

var fileDescriptorSql = []byte{
	// 385 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x74, 0xd2, 0xdd, 0x8e, 0xd2, 0x40,
	0x14, 0x07, 0x70, 0x86, 0xb6, 0xd0, 0x9e, 0x12, 0xad, 0xa3, 0x17, 0x44, 0x93, 0x06, 0x8b, 0x89,
	0x8d, 0x17, 0xa0, 0x48, 0x88, 0xb7, 0x96, 0x0f, 0xd3, 0xa4, 0x96, 0x38, 0xad, 0x46, 0x2f, 0x2b,
	0x4c, 0x08, 0xc9, 0x30, 0x03, 0xed, 0x60, 0xf4, 0x4d, 0x7c, 0x24, 0x2f, 0x7d, 0x84, 0x0d, 0xfb,
	0x22, 0x9b, 0x0e, 0x34, 0xbb, 0xb0, 0xd9, 0xbb, 0xf3, 0x3f, 0xfd, 0xe5, 0x9c, 0xe6, 0xb4, 0xf0,
	0x62, 0xcd, 0x25, 0xcd, 0x79, 0xc6, 0xfa, 0xdb, 0x5c, 0x48, 0xb1, 0x10, 0xac, 0x5f, 0xec, 0x58,
	0x4f, 0x05, 0x6c, 0x56, 0x3d, 0xcf, 0x82, 0x26, 0xa1, 0xbb, 0x3d, 0x2d, 0xa4, 0x07, 0x60, 0x12,
	0x5a, 0x6c, 0x05, 0x2f, 0xa8, 0x37, 0x01, 0x2b, 0x91, 0x99, 0xa4, 0x1b, 0xca, 0x25, 0xc6, 0xa0,
	0x4b, 0xfa, 0x5b, 0xb6, 0x51, 0x07, 0xf9, 0x16, 0x51, 0x35, 0xee, 0x82, 0x9e, 0xe5, 0xab, 0xa2,
	0x5d, 0xef, 0x68, 0xbe, 0x3d, 0x78, 0xdc, 0xab, 0x06, 0xf6, 0xbe, 0x65, 0x6c, 0x4f, 0x89, 0x7a,
	0xe8, 0x4d, 0xc0, 0x50, 0x11, 0xbf, 0x06, 0x7d, 0x21, 0x96, 0x54, 0x4d, 0x78, 0x34, 0x78, 0x7a,
	0xa1, 0xc7, 0x62, 0x49, 0x89, 0x02, 0xe5, 0xaa, 0x65, 0x26, 0xb3, 0x76, 0xbd, 0x83, 0xfc, 0x16,
	0x51, 0xb5, 0xe7, 0x01, 0x28, 0x16, 0x72, 0x39, 0x1a, 0xe2, 0x67, 0x60, 0xfc, 0x2a, 0x93, 0x9a,
	0xa5, 0x91, 0x63, 0xf0, 0x5e, 0x41, 0x4b, 0x99, 0x19, 0x13, 0xd9, 0x3d, 0x85, 0x2a, 0xf5, 0x12,
	0x2c, 0xa5, 0x02, 0x21, 0xd8, 0x39, 0x31, 0x2b, 0x52, 0x2d, 0x0b, 0xfe, 0x48, 0x5a, 0x9c, 0x9b,
	0x56, 0x65, 0xba, 0x60, 0x2b, 0x93, 0xc8, 0x7c, 0xcd, 0x57, 0xe7, 0xc8, 0xba, 0xdc, 0x95, 0xae,
	0x37, 0xf4, 0x81, 0x97, 0xb6, 0x4f, 0x24, 0xde, 0x33, 0xf6, 0xe6, 0xfb, 0x29, 0x94, 0xc7, 0xc0,
	0x16, 0x18, 0x61, 0x9c, 0x8e, 0x86, 0x4e, 0x0d, 0xdb, 0xd0, 0x9c, 0x45, 0xf3, 0x8f, 0x65, 0x40,
	0xd8, 0x04, 0x3d, 0x98, 0xcf, 0x23, 0xa7, 0x5e, 0x8a, 0xe0, 0x47, 0x3a, 0x4d, 0x1c, 0x0d, 0x03,
	0x34, 0x92, 0x94, 0x84, 0xf1, 0x27, 0x47, 0x2f, 0x41, 0x1a, 0x7e, 0x9e, 0x3a, 0x46, 0x59, 0xc5,
	0x5f, 0xa3, 0xc8, 0x69, 0x0c, 0x3e, 0x80, 0x96, 0x7c, 0x89, 0xf0, 0x3b, 0xd0, 0xc7, 0x82, 0x73,
	0xfc, 0xe4, 0xf6, 0xfa, 0xa7, 0x2f, 0xff, 0x1c, 0xdf, 0x6d, 0x1d, 0xff, 0x00, 0x1f, 0xbd, 0x45,
	0x81, 0xf3, 0xef, 0xe0, 0xa2, 0xff, 0x07, 0x17, 0x5d, 0x1d, 0x5c, 0xf4, 0xf7, 0xda, 0xad, 0xfd,
	0x6c, 0x28, 0xf8, 0xfe, 0x26, 0x00, 0x00, 0xff, 0xff, 0xe0, 0xa8, 0x9b, 0x0c, 0x5e, 0x02, 0x00,
	0x00,
}
