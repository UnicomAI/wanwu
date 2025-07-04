// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.4
// source: proto/rag-service/rag-service.proto

package rag_service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	RagService_ChatRag_FullMethodName         = "/rag_service.RagService/ChatRag"
	RagService_CreateRag_FullMethodName       = "/rag_service.RagService/CreateRag"
	RagService_UpdateRag_FullMethodName       = "/rag_service.RagService/UpdateRag"
	RagService_UpdateRagConfig_FullMethodName = "/rag_service.RagService/UpdateRagConfig"
	RagService_DeleteRag_FullMethodName       = "/rag_service.RagService/DeleteRag"
	RagService_GetRagDetail_FullMethodName    = "/rag_service.RagService/GetRagDetail"
	RagService_ListRag_FullMethodName         = "/rag_service.RagService/ListRag"
	RagService_GetRagByIds_FullMethodName     = "/rag_service.RagService/GetRagByIds"
)

// RagServiceClient is the client API for RagService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RagServiceClient interface {
	// 流式对话
	ChatRag(ctx context.Context, in *ChatRagReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ChatRagResp], error)
	// 创建 rag
	CreateRag(ctx context.Context, in *CreateRagReq, opts ...grpc.CallOption) (*CreateRagResp, error)
	// 更新 rag 基本信息
	UpdateRag(ctx context.Context, in *UpdateRagReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 更新 rag 配置信息
	UpdateRagConfig(ctx context.Context, in *UpdateRagConfigReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 删除 rag
	DeleteRag(ctx context.Context, in *RagDeleteReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// 获取 rag
	GetRagDetail(ctx context.Context, in *RagDetailReq, opts ...grpc.CallOption) (*RagInfo, error)
	// 获取 rag 列表
	ListRag(ctx context.Context, in *RagListReq, opts ...grpc.CallOption) (*RagListResp, error)
	// 根据 ragIds 获取 rag 列表
	GetRagByIds(ctx context.Context, in *GetRagByIdsReq, opts ...grpc.CallOption) (*AppBriefList, error)
}

type ragServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRagServiceClient(cc grpc.ClientConnInterface) RagServiceClient {
	return &ragServiceClient{cc}
}

func (c *ragServiceClient) ChatRag(ctx context.Context, in *ChatRagReq, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ChatRagResp], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &RagService_ServiceDesc.Streams[0], RagService_ChatRag_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[ChatRagReq, ChatRagResp]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RagService_ChatRagClient = grpc.ServerStreamingClient[ChatRagResp]

func (c *ragServiceClient) CreateRag(ctx context.Context, in *CreateRagReq, opts ...grpc.CallOption) (*CreateRagResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateRagResp)
	err := c.cc.Invoke(ctx, RagService_CreateRag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) UpdateRag(ctx context.Context, in *UpdateRagReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, RagService_UpdateRag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) UpdateRagConfig(ctx context.Context, in *UpdateRagConfigReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, RagService_UpdateRagConfig_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) DeleteRag(ctx context.Context, in *RagDeleteReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, RagService_DeleteRag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) GetRagDetail(ctx context.Context, in *RagDetailReq, opts ...grpc.CallOption) (*RagInfo, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RagInfo)
	err := c.cc.Invoke(ctx, RagService_GetRagDetail_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) ListRag(ctx context.Context, in *RagListReq, opts ...grpc.CallOption) (*RagListResp, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RagListResp)
	err := c.cc.Invoke(ctx, RagService_ListRag_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ragServiceClient) GetRagByIds(ctx context.Context, in *GetRagByIdsReq, opts ...grpc.CallOption) (*AppBriefList, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AppBriefList)
	err := c.cc.Invoke(ctx, RagService_GetRagByIds_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RagServiceServer is the server API for RagService service.
// All implementations must embed UnimplementedRagServiceServer
// for forward compatibility.
type RagServiceServer interface {
	// 流式对话
	ChatRag(*ChatRagReq, grpc.ServerStreamingServer[ChatRagResp]) error
	// 创建 rag
	CreateRag(context.Context, *CreateRagReq) (*CreateRagResp, error)
	// 更新 rag 基本信息
	UpdateRag(context.Context, *UpdateRagReq) (*emptypb.Empty, error)
	// 更新 rag 配置信息
	UpdateRagConfig(context.Context, *UpdateRagConfigReq) (*emptypb.Empty, error)
	// 删除 rag
	DeleteRag(context.Context, *RagDeleteReq) (*emptypb.Empty, error)
	// 获取 rag
	GetRagDetail(context.Context, *RagDetailReq) (*RagInfo, error)
	// 获取 rag 列表
	ListRag(context.Context, *RagListReq) (*RagListResp, error)
	// 根据 ragIds 获取 rag 列表
	GetRagByIds(context.Context, *GetRagByIdsReq) (*AppBriefList, error)
	mustEmbedUnimplementedRagServiceServer()
}

// UnimplementedRagServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRagServiceServer struct{}

func (UnimplementedRagServiceServer) ChatRag(*ChatRagReq, grpc.ServerStreamingServer[ChatRagResp]) error {
	return status.Errorf(codes.Unimplemented, "method ChatRag not implemented")
}
func (UnimplementedRagServiceServer) CreateRag(context.Context, *CreateRagReq) (*CreateRagResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRag not implemented")
}
func (UnimplementedRagServiceServer) UpdateRag(context.Context, *UpdateRagReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRag not implemented")
}
func (UnimplementedRagServiceServer) UpdateRagConfig(context.Context, *UpdateRagConfigReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRagConfig not implemented")
}
func (UnimplementedRagServiceServer) DeleteRag(context.Context, *RagDeleteReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteRag not implemented")
}
func (UnimplementedRagServiceServer) GetRagDetail(context.Context, *RagDetailReq) (*RagInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRagDetail not implemented")
}
func (UnimplementedRagServiceServer) ListRag(context.Context, *RagListReq) (*RagListResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRag not implemented")
}
func (UnimplementedRagServiceServer) GetRagByIds(context.Context, *GetRagByIdsReq) (*AppBriefList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRagByIds not implemented")
}
func (UnimplementedRagServiceServer) mustEmbedUnimplementedRagServiceServer() {}
func (UnimplementedRagServiceServer) testEmbeddedByValue()                    {}

// UnsafeRagServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RagServiceServer will
// result in compilation errors.
type UnsafeRagServiceServer interface {
	mustEmbedUnimplementedRagServiceServer()
}

func RegisterRagServiceServer(s grpc.ServiceRegistrar, srv RagServiceServer) {
	// If the following call pancis, it indicates UnimplementedRagServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RagService_ServiceDesc, srv)
}

func _RagService_ChatRag_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ChatRagReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RagServiceServer).ChatRag(m, &grpc.GenericServerStream[ChatRagReq, ChatRagResp]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type RagService_ChatRagServer = grpc.ServerStreamingServer[ChatRagResp]

func _RagService_CreateRag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRagReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).CreateRag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_CreateRag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).CreateRag(ctx, req.(*CreateRagReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_UpdateRag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRagReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).UpdateRag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_UpdateRag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).UpdateRag(ctx, req.(*UpdateRagReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_UpdateRagConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRagConfigReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).UpdateRagConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_UpdateRagConfig_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).UpdateRagConfig(ctx, req.(*UpdateRagConfigReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_DeleteRag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RagDeleteReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).DeleteRag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_DeleteRag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).DeleteRag(ctx, req.(*RagDeleteReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_GetRagDetail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RagDetailReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).GetRagDetail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_GetRagDetail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).GetRagDetail(ctx, req.(*RagDetailReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_ListRag_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RagListReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).ListRag(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_ListRag_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).ListRag(ctx, req.(*RagListReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _RagService_GetRagByIds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRagByIdsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RagServiceServer).GetRagByIds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RagService_GetRagByIds_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RagServiceServer).GetRagByIds(ctx, req.(*GetRagByIdsReq))
	}
	return interceptor(ctx, in, info, handler)
}

// RagService_ServiceDesc is the grpc.ServiceDesc for RagService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RagService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rag_service.RagService",
	HandlerType: (*RagServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateRag",
			Handler:    _RagService_CreateRag_Handler,
		},
		{
			MethodName: "UpdateRag",
			Handler:    _RagService_UpdateRag_Handler,
		},
		{
			MethodName: "UpdateRagConfig",
			Handler:    _RagService_UpdateRagConfig_Handler,
		},
		{
			MethodName: "DeleteRag",
			Handler:    _RagService_DeleteRag_Handler,
		},
		{
			MethodName: "GetRagDetail",
			Handler:    _RagService_GetRagDetail_Handler,
		},
		{
			MethodName: "ListRag",
			Handler:    _RagService_ListRag_Handler,
		},
		{
			MethodName: "GetRagByIds",
			Handler:    _RagService_GetRagByIds_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ChatRag",
			Handler:       _RagService_ChatRag_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/rag-service/rag-service.proto",
}
