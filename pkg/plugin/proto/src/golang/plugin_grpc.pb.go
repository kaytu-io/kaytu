// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: pkg/plugin/proto/plugin.proto

package golang

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
	Plugin_Register_FullMethodName = "/kaytu.plugin.v1.Plugin/Register"
)

// PluginClient is the client API for Plugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginClient interface {
	Register(ctx context.Context, opts ...grpc.CallOption) (Plugin_RegisterClient, error)
}

type pluginClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginClient(cc grpc.ClientConnInterface) PluginClient {
	return &pluginClient{cc}
}

func (c *pluginClient) Register(ctx context.Context, opts ...grpc.CallOption) (Plugin_RegisterClient, error) {
	stream, err := c.cc.NewStream(ctx, &Plugin_ServiceDesc.Streams[0], Plugin_Register_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &pluginRegisterClient{stream}
	return x, nil
}

type Plugin_RegisterClient interface {
	Send(*PluginMessage) error
	Recv() (*ServerMessage, error)
	grpc.ClientStream
}

type pluginRegisterClient struct {
	grpc.ClientStream
}

func (x *pluginRegisterClient) Send(m *PluginMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *pluginRegisterClient) Recv() (*ServerMessage, error) {
	m := new(ServerMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PluginServer is the server API for Plugin service.
// All implementations must embed UnimplementedPluginServer
// for forward compatibility
type PluginServer interface {
	Register(Plugin_RegisterServer) error
	mustEmbedUnimplementedPluginServer()
}

// UnimplementedPluginServer must be embedded to have forward compatible implementations.
type UnimplementedPluginServer struct {
}

func (UnimplementedPluginServer) Register(Plugin_RegisterServer) error {
	return status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedPluginServer) mustEmbedUnimplementedPluginServer() {}

// UnsafePluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginServer will
// result in compilation errors.
type UnsafePluginServer interface {
	mustEmbedUnimplementedPluginServer()
}

func RegisterPluginServer(s grpc.ServiceRegistrar, srv PluginServer) {
	s.RegisterService(&Plugin_ServiceDesc, srv)
}

func _Plugin_Register_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PluginServer).Register(&pluginRegisterServer{stream})
}

type Plugin_RegisterServer interface {
	Send(*ServerMessage) error
	Recv() (*PluginMessage, error)
	grpc.ServerStream
}

type pluginRegisterServer struct {
	grpc.ServerStream
}

func (x *pluginRegisterServer) Send(m *ServerMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *pluginRegisterServer) Recv() (*PluginMessage, error) {
	m := new(PluginMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Plugin_ServiceDesc is the grpc.ServiceDesc for Plugin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Plugin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kaytu.plugin.v1.Plugin",
	HandlerType: (*PluginServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Register",
			Handler:       _Plugin_Register_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "pkg/plugin/proto/plugin.proto",
}