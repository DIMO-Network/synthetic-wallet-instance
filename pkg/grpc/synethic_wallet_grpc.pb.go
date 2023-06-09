// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: pkg/grpc/synethic_wallet.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SyntheticWallet_GetAddress_FullMethodName = "/grpc.SyntheticWallet/GetAddress"
	SyntheticWallet_SignHash_FullMethodName   = "/grpc.SyntheticWallet/SignHash"
)

// SyntheticWalletClient is the client API for SyntheticWallet service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SyntheticWalletClient interface {
	GetAddress(ctx context.Context, in *GetAddressRequest, opts ...grpc.CallOption) (*GetAddressResponse, error)
	SignHash(ctx context.Context, in *SignHashRequest, opts ...grpc.CallOption) (*SignHashResponse, error)
}

type syntheticWalletClient struct {
	cc grpc.ClientConnInterface
}

func NewSyntheticWalletClient(cc grpc.ClientConnInterface) SyntheticWalletClient {
	return &syntheticWalletClient{cc}
}

func (c *syntheticWalletClient) GetAddress(ctx context.Context, in *GetAddressRequest, opts ...grpc.CallOption) (*GetAddressResponse, error) {
	out := new(GetAddressResponse)
	err := c.cc.Invoke(ctx, SyntheticWallet_GetAddress_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *syntheticWalletClient) SignHash(ctx context.Context, in *SignHashRequest, opts ...grpc.CallOption) (*SignHashResponse, error) {
	out := new(SignHashResponse)
	err := c.cc.Invoke(ctx, SyntheticWallet_SignHash_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// SyntheticWalletServer is the server API for SyntheticWallet service.
// All implementations must embed UnimplementedSyntheticWalletServer
// for forward compatibility
type SyntheticWalletServer interface {
	GetAddress(context.Context, *GetAddressRequest) (*GetAddressResponse, error)
	SignHash(context.Context, *SignHashRequest) (*SignHashResponse, error)
	mustEmbedUnimplementedSyntheticWalletServer()
}

// UnimplementedSyntheticWalletServer must be embedded to have forward compatible implementations.
type UnimplementedSyntheticWalletServer struct {
}

func (UnimplementedSyntheticWalletServer) GetAddress(context.Context, *GetAddressRequest) (*GetAddressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAddress not implemented")
}
func (UnimplementedSyntheticWalletServer) SignHash(context.Context, *SignHashRequest) (*SignHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignHash not implemented")
}
func (UnimplementedSyntheticWalletServer) mustEmbedUnimplementedSyntheticWalletServer() {}

// UnsafeSyntheticWalletServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SyntheticWalletServer will
// result in compilation errors.
type UnsafeSyntheticWalletServer interface {
	mustEmbedUnimplementedSyntheticWalletServer()
}

func RegisterSyntheticWalletServer(s grpc.ServiceRegistrar, srv SyntheticWalletServer) {
	s.RegisterService(&SyntheticWallet_ServiceDesc, srv)
}

func _SyntheticWallet_GetAddress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAddressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyntheticWalletServer).GetAddress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SyntheticWallet_GetAddress_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyntheticWalletServer).GetAddress(ctx, req.(*GetAddressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SyntheticWallet_SignHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SyntheticWalletServer).SignHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SyntheticWallet_SignHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SyntheticWalletServer).SignHash(ctx, req.(*SignHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// SyntheticWallet_ServiceDesc is the grpc.ServiceDesc for SyntheticWallet service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SyntheticWallet_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.SyntheticWallet",
	HandlerType: (*SyntheticWalletServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAddress",
			Handler:    _SyntheticWallet_GetAddress_Handler,
		},
		{
			MethodName: "SignHash",
			Handler:    _SyntheticWallet_SignHash_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/synethic_wallet.proto",
}
