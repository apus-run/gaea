package server

import (
	"context"
	"net/url"
)

// Server ...
type Server interface {
	Start(context.Context) error
	Stop(context.Context) error

	// Endpoint return server or client endpoint
	// Server Transport: grpc://127.0.0.1:9000
	Endpoint() (*url.URL, error)
}

type (
	grpcServerKey struct{}
	grpcClientKey struct{}
)

func NewServerContext(ctx context.Context, srv Server) context.Context {
	return context.WithValue(ctx, grpcServerKey{}, srv)
}

func FromServerContext(ctx context.Context) (srv Server, ok bool) {
	srv, ok = ctx.Value(grpcServerKey{}).(Server)
	return
}

func NewClientContext(ctx context.Context, srv Server) context.Context {
	return context.WithValue(ctx, grpcClientKey{}, srv)
}

func FromClientContext(ctx context.Context) (srv Server, ok bool) {
	srv, ok = ctx.Value(grpcClientKey{}).(Server)
	return
}
