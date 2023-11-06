package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"

	"github.com/apus-run/gaea/middleware"
	"github.com/apus-run/gaea/registry"
	"github.com/apus-run/gaea/server/grpc/resolver/discovery"
)

// Client is gRPC Client
type Client struct {
	endpoint               string
	timeout                time.Duration
	tlsConf                *tls.Config
	discovery              registry.Discovery
	ms                     []middleware.Middleware
	ints                   []grpc.UnaryClientInterceptor
	streamInts             []grpc.StreamClientInterceptor
	grpcOpts               []grpc.DialOption
	balancerName           string
	printDiscoveryDebugLog bool
}

// defaultClient return a default config server
func defaultClient() *Client {
	return &Client{
		timeout:                2000 * time.Millisecond,
		balancerName:           roundrobin.Name,
		printDiscoveryDebugLog: true,
	}
}

func ApplyClient(opts ...ClientOption) *Client {
	c := defaultClient()
	for _, o := range opts {
		o(c)
	}
	return c
}

// ClientOption is gRPC client option.
type ClientOption func(o *Client)

// WithEndpoint ...
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithGrpcOptions with gRPC options.
func WithGrpcOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *Client) {
		c.grpcOpts = opts
	}
}

// WithMiddleware with client middleware.
func WithMiddleware(ms ...middleware.Middleware) ClientOption {
	return func(c *Client) {
		c.ms = ms
	}
}

// WithTLSConfig with TLS config.
func WithTLSConfig(conf *tls.Config) ClientOption {
	return func(c *Client) {
		c.tlsConf = conf
	}
}

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for unary RPCs.
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *Client) {
		c.ints = in
	}
}

// WithStreamInterceptor returns a DialOption that specifies the interceptor for streaming RPCs.
func WithStreamInterceptor(in ...grpc.StreamClientInterceptor) ClientOption {
	return func(c *Client) {
		c.streamInts = in
	}
}

// WithBalancerName with balancer name
func WithBalancerName(name string) ClientOption {
	return func(c *Client) {
		c.balancerName = name
	}
}

// WithTimeout with client timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = timeout
	}
}

// WithDiscovery with client discovery.
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(c *Client) {
		c.discovery = d
	}
}

func WithPrintDiscoveryDebugLog(p bool) ClientOption {
	return func(c *Client) {
		c.printDiscoveryDebugLog = p
	}
}

// Dial returns a GRPC connection.
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, false, opts...)
}

// DialInsecure returns an insecure GRPC connection.
func DialInsecure(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	return dial(ctx, true, opts...)
}

func dial(ctx context.Context, insecure bool, opts ...ClientOption) (*grpc.ClientConn, error) {
	client := ApplyClient(opts...)
	ints := []grpc.UnaryClientInterceptor{
		client.unaryClientInterceptor(client.ms, client.timeout),
	}
	sints := []grpc.StreamClientInterceptor{
		client.streamClientInterceptor(),
	}
	if len(client.ints) > 0 {
		ints = append(ints, client.ints...)
	}
	if len(client.streamInts) > 0 {
		sints = append(sints, client.streamInts...)
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingConfig": [{"%s":{}}],"healthCheckConfig":{"serviceName":""}}`, client.balancerName)),
		grpc.WithChainUnaryInterceptor(ints...),
		grpc.WithChainStreamInterceptor(sints...),
	}
	if client.discovery != nil {
		grpcOpts = append(grpcOpts,
			grpc.WithResolvers(
				discovery.NewBuilder(
					client.discovery,
					discovery.WithInsecure(insecure),
					discovery.PrintDebugLog(client.printDiscoveryDebugLog),
				)))
	}
	if insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}
	if client.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(client.tlsConf)))
	}
	if len(client.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, client.grpcOpts...)
	}

	return grpc.DialContext(ctx, client.endpoint, grpcOpts...)
}
