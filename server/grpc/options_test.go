package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"reflect"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/apus-run/gaea/middleware"
	"github.com/apus-run/gaea/registry"
)

func TestNetwork(t *testing.T) {
	o := &Server{}
	v := "abc"
	Network(v)(o)
	if !reflect.DeepEqual(v, o.network) {
		t.Errorf("expect %s, got %s", v, o.network)
	}
}

func TestAddress(t *testing.T) {
	v := "abc"
	o := NewServer(Address(v))
	if !reflect.DeepEqual(v, o.address) {
		t.Errorf("expect %s, got %s", v, o.address)
	}
	u, err := o.Endpoint()
	if err == nil {
		t.Errorf("expect %s, got %s", v, err)
	}
	if u != nil {
		t.Errorf("expect %s, got %s", v, u)
	}
}

func TestMiddleware(t *testing.T) {
	o := &Server{}
	v := []middleware.Middleware{
		func(middleware.Handler) middleware.Handler { return nil },
	}
	Middleware(v...)(o)
	if !reflect.DeepEqual(v, o.middleware) {
		t.Errorf("expect %v, got %v", v, o.middleware)
	}
}

func TestTLSConfig(t *testing.T) {
	o := &Server{}
	v := &tls.Config{}
	TLSConfig(v)(o)
	if !reflect.DeepEqual(v, o.tlsConf) {
		t.Errorf("expect %v, got %v", v, o.tlsConf)
	}
}

func TestUnaryInterceptor(t *testing.T) {
	o := &Server{}
	v := []grpc.UnaryServerInterceptor{
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return nil, nil
		},
		func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			return nil, nil
		},
	}
	UnaryInterceptor(v...)(o)
	if !reflect.DeepEqual(v, o.unaryInterceptor) {
		t.Errorf("expect %v, got %v", v, o.unaryInterceptor)
	}
}

func TestStreamInterceptor(t *testing.T) {
	o := &Server{}
	v := []grpc.StreamServerInterceptor{
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return nil
		},
		func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return nil
		},
	}
	StreamInterceptor(v...)(o)
	if !reflect.DeepEqual(v, o.streamInterceptor) {
		t.Errorf("expect %v, got %v", v, o.streamInterceptor)
	}
}

func TestListener(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{}
	Listener(lis)(s)
	if !reflect.DeepEqual(lis, s.lis) {
		t.Errorf("expect %v, got %v", lis, s.lis)
	}
	if e, err := s.Endpoint(); err != nil || e == nil {
		t.Errorf("expect not empty")
	}
}

func TestOptions(t *testing.T) {
	o := &Server{}
	v := []grpc.ServerOption{
		grpc.EmptyServerOption{},
	}
	GrpcOptions(v...)(o)
	if !reflect.DeepEqual(v, o.grpcOpts) {
		t.Errorf("expect %v, got %v", v, o.grpcOpts)
	}
}

func TestWithEndpoint(t *testing.T) {
	o := &Client{}
	v := "abc"
	WithEndpoint(v)(o)
	if !reflect.DeepEqual(v, o.endpoint) {
		t.Errorf("expect %v but got %v", v, o.endpoint)
	}
}

func TestWithTimeout(t *testing.T) {
	o := &Client{}
	v := time.Duration(123)
	WithTimeout(v)(o)
	if !reflect.DeepEqual(v, o.timeout) {
		t.Errorf("expect %v but got %v", v, o.timeout)
	}
}

func TestWithMiddleware(t *testing.T) {
	o := &Client{}
	v := []middleware.Middleware{
		func(middleware.Handler) middleware.Handler { return nil },
	}
	WithMiddleware(v...)(o)
	if !reflect.DeepEqual(v, o.ms) {
		t.Errorf("expect %v but got %v", v, o.ms)
	}
}

func TestWithDiscovery(t *testing.T) {
	o := &Client{}
	v := &mockRegistry{}
	WithDiscovery(v)(o)
	if !reflect.DeepEqual(v, o.discovery) {
		t.Errorf("expect %v but got %v", v, o.discovery)
	}
}

func TestWithTLSConfig(t *testing.T) {
	o := &Client{}
	v := &tls.Config{}
	WithTLSConfig(v)(o)
	if !reflect.DeepEqual(v, o.tlsConf) {
		t.Errorf("expect %v but got %v", v, o.tlsConf)
	}
}

func TestWithOptions(t *testing.T) {
	o := &Client{}
	v := []grpc.DialOption{
		grpc.EmptyDialOption{},
	}
	WithGrpcOptions(v...)(o)
	if !reflect.DeepEqual(v, o.grpcOpts) {
		t.Errorf("expect %v but got %v", v, o.grpcOpts)
	}
}

type mockRegistry struct{}

func (m *mockRegistry) GetServiceList(ctx context.Context) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

func (m *mockRegistry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	return nil, nil
}

func (m *mockRegistry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return nil, nil
}
