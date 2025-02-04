// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.4.0
// - protoc             (unknown)
// source: extention.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.62.0 or later.
const _ = grpc.SupportPackageIsVersion8

const (
	AuthzExtentionService_BatchCheck_FullMethodName = "/authz.extention.v1.AuthzExtentionService/BatchCheck"
	AuthzExtentionService_Read_FullMethodName       = "/authz.extention.v1.AuthzExtentionService/Read"
	AuthzExtentionService_Write_FullMethodName      = "/authz.extention.v1.AuthzExtentionService/Write"
)

// AuthzExtentionServiceClient is the client API for AuthzExtentionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthzExtentionServiceClient interface {
	BatchCheck(ctx context.Context, in *BatchCheckRequest, opts ...grpc.CallOption) (*BatchCheckResponse, error)
	Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error)
	Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error)
}

type authzExtentionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthzExtentionServiceClient(cc grpc.ClientConnInterface) AuthzExtentionServiceClient {
	return &authzExtentionServiceClient{cc}
}

func (c *authzExtentionServiceClient) BatchCheck(ctx context.Context, in *BatchCheckRequest, opts ...grpc.CallOption) (*BatchCheckResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BatchCheckResponse)
	err := c.cc.Invoke(ctx, AuthzExtentionService_BatchCheck_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authzExtentionServiceClient) Read(ctx context.Context, in *ReadRequest, opts ...grpc.CallOption) (*ReadResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReadResponse)
	err := c.cc.Invoke(ctx, AuthzExtentionService_Read_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authzExtentionServiceClient) Write(ctx context.Context, in *WriteRequest, opts ...grpc.CallOption) (*WriteResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(WriteResponse)
	err := c.cc.Invoke(ctx, AuthzExtentionService_Write_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthzExtentionServiceServer is the server API for AuthzExtentionService service.
// All implementations should embed UnimplementedAuthzExtentionServiceServer
// for forward compatibility
type AuthzExtentionServiceServer interface {
	BatchCheck(context.Context, *BatchCheckRequest) (*BatchCheckResponse, error)
	Read(context.Context, *ReadRequest) (*ReadResponse, error)
	Write(context.Context, *WriteRequest) (*WriteResponse, error)
}

// UnimplementedAuthzExtentionServiceServer should be embedded to have forward compatible implementations.
type UnimplementedAuthzExtentionServiceServer struct {
}

func (UnimplementedAuthzExtentionServiceServer) BatchCheck(context.Context, *BatchCheckRequest) (*BatchCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BatchCheck not implemented")
}
func (UnimplementedAuthzExtentionServiceServer) Read(context.Context, *ReadRequest) (*ReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedAuthzExtentionServiceServer) Write(context.Context, *WriteRequest) (*WriteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Write not implemented")
}

// UnsafeAuthzExtentionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthzExtentionServiceServer will
// result in compilation errors.
type UnsafeAuthzExtentionServiceServer interface {
	mustEmbedUnimplementedAuthzExtentionServiceServer()
}

func RegisterAuthzExtentionServiceServer(s grpc.ServiceRegistrar, srv AuthzExtentionServiceServer) {
	s.RegisterService(&AuthzExtentionService_ServiceDesc, srv)
}

func _AuthzExtentionService_BatchCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzExtentionServiceServer).BatchCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthzExtentionService_BatchCheck_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzExtentionServiceServer).BatchCheck(ctx, req.(*BatchCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthzExtentionService_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzExtentionServiceServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthzExtentionService_Read_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzExtentionServiceServer).Read(ctx, req.(*ReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthzExtentionService_Write_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthzExtentionServiceServer).Write(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthzExtentionService_Write_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthzExtentionServiceServer).Write(ctx, req.(*WriteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthzExtentionService_ServiceDesc is the grpc.ServiceDesc for AuthzExtentionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthzExtentionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "authz.extention.v1.AuthzExtentionService",
	HandlerType: (*AuthzExtentionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BatchCheck",
			Handler:    _AuthzExtentionService_BatchCheck_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _AuthzExtentionService_Read_Handler,
		},
		{
			MethodName: "Write",
			Handler:    _AuthzExtentionService_Write_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "extention.proto",
}
