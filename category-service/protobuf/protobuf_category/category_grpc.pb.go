// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.0
// source: protobuf/category.proto

package protobuf_category

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	CategoryService_CreateCategory_FullMethodName  = "/protobuf_category.CategoryService/CreateCategory"
	CategoryService_GetCategoryList_FullMethodName = "/protobuf_category.CategoryService/GetCategoryList"
	CategoryService_GetCategoryById_FullMethodName = "/protobuf_category.CategoryService/GetCategoryById"
	CategoryService_UpdateCategory_FullMethodName  = "/protobuf_category.CategoryService/UpdateCategory"
	CategoryService_DeleteCategory_FullMethodName  = "/protobuf_category.CategoryService/DeleteCategory"
)

// CategoryServiceClient is the client API for CategoryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CategoryServiceClient interface {
	CreateCategory(ctx context.Context, in *ProtoCategoryRepo_ProtoCategory, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error)
	GetCategoryList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ProtoCategoryRepo_ProtoCategory], error)
	GetCategoryById(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error)
	UpdateCategory(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error)
	DeleteCategory(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*DeleteCategoryResponse, error)
}

type categoryServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCategoryServiceClient(cc grpc.ClientConnInterface) CategoryServiceClient {
	return &categoryServiceClient{cc}
}

func (c *categoryServiceClient) CreateCategory(ctx context.Context, in *ProtoCategoryRepo_ProtoCategory, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProtoCategoryRepo_ProtoCategory)
	err := c.cc.Invoke(ctx, CategoryService_CreateCategory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *categoryServiceClient) GetCategoryList(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (grpc.ServerStreamingClient[ProtoCategoryRepo_ProtoCategory], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &CategoryService_ServiceDesc.Streams[0], CategoryService_GetCategoryList_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[emptypb.Empty, ProtoCategoryRepo_ProtoCategory]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CategoryService_GetCategoryListClient = grpc.ServerStreamingClient[ProtoCategoryRepo_ProtoCategory]

func (c *categoryServiceClient) GetCategoryById(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProtoCategoryRepo_ProtoCategory)
	err := c.cc.Invoke(ctx, CategoryService_GetCategoryById_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *categoryServiceClient) UpdateCategory(ctx context.Context, in *UpdateRequest, opts ...grpc.CallOption) (*ProtoCategoryRepo_ProtoCategory, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProtoCategoryRepo_ProtoCategory)
	err := c.cc.Invoke(ctx, CategoryService_UpdateCategory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *categoryServiceClient) DeleteCategory(ctx context.Context, in *wrapperspb.Int64Value, opts ...grpc.CallOption) (*DeleteCategoryResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(DeleteCategoryResponse)
	err := c.cc.Invoke(ctx, CategoryService_DeleteCategory_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CategoryServiceServer is the server API for CategoryService service.
// All implementations should embed UnimplementedCategoryServiceServer
// for forward compatibility.
type CategoryServiceServer interface {
	CreateCategory(context.Context, *ProtoCategoryRepo_ProtoCategory) (*ProtoCategoryRepo_ProtoCategory, error)
	GetCategoryList(*emptypb.Empty, grpc.ServerStreamingServer[ProtoCategoryRepo_ProtoCategory]) error
	GetCategoryById(context.Context, *wrapperspb.Int64Value) (*ProtoCategoryRepo_ProtoCategory, error)
	UpdateCategory(context.Context, *UpdateRequest) (*ProtoCategoryRepo_ProtoCategory, error)
	DeleteCategory(context.Context, *wrapperspb.Int64Value) (*DeleteCategoryResponse, error)
}

// UnimplementedCategoryServiceServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCategoryServiceServer struct{}

func (UnimplementedCategoryServiceServer) CreateCategory(context.Context, *ProtoCategoryRepo_ProtoCategory) (*ProtoCategoryRepo_ProtoCategory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCategory not implemented")
}
func (UnimplementedCategoryServiceServer) GetCategoryList(*emptypb.Empty, grpc.ServerStreamingServer[ProtoCategoryRepo_ProtoCategory]) error {
	return status.Errorf(codes.Unimplemented, "method GetCategoryList not implemented")
}
func (UnimplementedCategoryServiceServer) GetCategoryById(context.Context, *wrapperspb.Int64Value) (*ProtoCategoryRepo_ProtoCategory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCategoryById not implemented")
}
func (UnimplementedCategoryServiceServer) UpdateCategory(context.Context, *UpdateRequest) (*ProtoCategoryRepo_ProtoCategory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCategory not implemented")
}
func (UnimplementedCategoryServiceServer) DeleteCategory(context.Context, *wrapperspb.Int64Value) (*DeleteCategoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCategory not implemented")
}
func (UnimplementedCategoryServiceServer) testEmbeddedByValue() {}

// UnsafeCategoryServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CategoryServiceServer will
// result in compilation errors.
type UnsafeCategoryServiceServer interface {
	mustEmbedUnimplementedCategoryServiceServer()
}

func RegisterCategoryServiceServer(s grpc.ServiceRegistrar, srv CategoryServiceServer) {
	// If the following call pancis, it indicates UnimplementedCategoryServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CategoryService_ServiceDesc, srv)
}

func _CategoryService_CreateCategory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProtoCategoryRepo_ProtoCategory)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CategoryServiceServer).CreateCategory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CategoryService_CreateCategory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CategoryServiceServer).CreateCategory(ctx, req.(*ProtoCategoryRepo_ProtoCategory))
	}
	return interceptor(ctx, in, info, handler)
}

func _CategoryService_GetCategoryList_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(emptypb.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CategoryServiceServer).GetCategoryList(m, &grpc.GenericServerStream[emptypb.Empty, ProtoCategoryRepo_ProtoCategory]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type CategoryService_GetCategoryListServer = grpc.ServerStreamingServer[ProtoCategoryRepo_ProtoCategory]

func _CategoryService_GetCategoryById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.Int64Value)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CategoryServiceServer).GetCategoryById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CategoryService_GetCategoryById_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CategoryServiceServer).GetCategoryById(ctx, req.(*wrapperspb.Int64Value))
	}
	return interceptor(ctx, in, info, handler)
}

func _CategoryService_UpdateCategory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CategoryServiceServer).UpdateCategory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CategoryService_UpdateCategory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CategoryServiceServer).UpdateCategory(ctx, req.(*UpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CategoryService_DeleteCategory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(wrapperspb.Int64Value)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CategoryServiceServer).DeleteCategory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CategoryService_DeleteCategory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CategoryServiceServer).DeleteCategory(ctx, req.(*wrapperspb.Int64Value))
	}
	return interceptor(ctx, in, info, handler)
}

// CategoryService_ServiceDesc is the grpc.ServiceDesc for CategoryService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CategoryService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "protobuf_category.CategoryService",
	HandlerType: (*CategoryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateCategory",
			Handler:    _CategoryService_CreateCategory_Handler,
		},
		{
			MethodName: "GetCategoryById",
			Handler:    _CategoryService_GetCategoryById_Handler,
		},
		{
			MethodName: "UpdateCategory",
			Handler:    _CategoryService_UpdateCategory_Handler,
		},
		{
			MethodName: "DeleteCategory",
			Handler:    _CategoryService_DeleteCategory_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetCategoryList",
			Handler:       _CategoryService_GetCategoryList_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "protobuf/category.proto",
}