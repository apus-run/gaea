package grpc

import (
	"context"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apus-run/gaea/middleware"
)

func EmptyMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			return handler(ctx, req)
		}
	}
}

func TestUnaryClientInterceptor(t *testing.T) {
	o := &Client{}
	f := o.unaryClientInterceptor([]middleware.Middleware{EmptyMiddleware()}, time.Duration(100))
	req := &struct{}{}
	resp := &struct{}{}

	err := f(context.TODO(), "hello", req, resp, &grpc.ClientConn{},
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWithUnaryInterceptor(t *testing.T) {
	o := &Client{}
	v := []grpc.UnaryClientInterceptor{
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		},
		func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			return nil
		},
	}
	WithUnaryInterceptor(v...)(o)
	if !reflect.DeepEqual(v, o.ints) {
		t.Errorf("expect %v but got %v", v, o.ints)
	}
}

func TestDial(t *testing.T) {
	o := &Client{}
	v := []grpc.DialOption{
		grpc.EmptyDialOption{},
	}
	WithDialOptions(v...)(o)
	if !reflect.DeepEqual(v, o.dialOpts) {
		t.Errorf("expect %v but got %v", v, o.dialOpts)
	}
}

func TestDialConn(t *testing.T) {
	_, err := dial(
		context.Background(),
		true,
		WithDiscovery(&mockRegistry{}),
		WithTimeout(10*time.Second),
		WithEndpoint("abc"),
		WithMiddleware(EmptyMiddleware()),
	)
	if err != nil {
		t.Error(err)
	}
}
