package gateway

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// refer: https://github.com/golang/protobuf/blob/v1.4.3/jsonpb/encode.go#L30
var defaultServerMuxOption = gwRuntime.WithMarshalerOption(gwRuntime.MIMEWildcard, &gwRuntime.JSONPb{})

// AnnotatorFunc is the annotator function is for injecting metadata from http request into gRPC context
type AnnotatorFunc func(context.Context, *http.Request) metadata.MD

// HandlerFromEndpoint is the callback that the caller should implement
// to steps to reverse-proxy the HTTP/1 requests to gRPC
// handlerFromEndpoint http gw endPoint
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
type HandlerFromEndpoint func(ctx context.Context, mux *gwRuntime.ServeMux,
	endpoint string, opts []grpc.DialOption) error

type HandlerFunc func(ctx context.Context, mux *gwRuntime.ServeMux, conn *grpc.ClientConn) error

type Gateway struct {
	*http.Server // if you need gRPC gw,please use it

	lis          net.Listener
	tlsConf      *tls.Config
	endpoint     *url.URL
	err          error
	network      string
	address      string
	shutdownFunc func() // shutdown func
	timeout      time.Duration

	conn                    *grpc.ClientConn
	mux                     *gwRuntime.ServeMux
	serveMuxOptions         []gwRuntime.ServeMuxOption
	registerServiceHandlers []HandlerFunc
	annotators              []AnnotatorFunc
}

// GatewayOption is gRPC Gateway option.
type GatewayOption func(o *Gateway)

func WithNetwork(network string) GatewayOption {
	return func(s *Gateway) {
		s.network = network
	}
}

func WithAddress(addr string) GatewayOption {
	return func(s *Gateway) {
		s.address = addr
	}
}

func WithEndpoint(endpoint *url.URL) GatewayOption {
	return func(s *Gateway) {
		s.endpoint = endpoint
	}
}

func WithListener(lis net.Listener) GatewayOption {
	return func(s *Gateway) {
		s.lis = lis
	}
}

func WithTLSConfig(c *tls.Config) GatewayOption {
	return func(o *Gateway) {
		o.tlsConf = c
	}
}

func WithTimeout(timeout time.Duration) GatewayOption {
	return func(s *Gateway) {
		s.timeout = timeout
	}
}

// WithShutdownFunc returns an Option to register a function which will be called when server shutdown
func WithShutdownFunc(f func()) GatewayOption {
	return func(s *Gateway) {
		s.shutdownFunc = f
	}
}

func WithConn(conn *grpc.ClientConn) GatewayOption {
	return func(g *Gateway) {
		g.conn = conn
	}
}

func WithHandlers(handlers ...HandlerFunc) GatewayOption {
	return func(g *Gateway) {
		g.registerServiceHandlers = handlers
	}
}

func WithServeMuxOptions(serveMuxOptions ...gwRuntime.ServeMuxOption) GatewayOption {
	return func(g *Gateway) {
		g.serveMuxOptions = serveMuxOptions
	}
}

// WithAnnotator returns an Option to append some annotator
func WithAnnotator(annotator ...AnnotatorFunc) GatewayOption {
	return func(s *Gateway) {
		s.annotators = append(s.annotators, annotator...)
	}
}

func defaultGateway() *Gateway {
	g := &Gateway{
		network:      "tcp",
		address:      ":0",
		shutdownFunc: func() {},
		timeout:      1 * time.Second,
	}

	g.serveMuxOptions = append(g.serveMuxOptions, defaultServerMuxOption)

	return g
}

func ApplyGateway(opts ...GatewayOption) *Gateway {
	g := defaultGateway()
	for _, o := range opts {
		o(g)
	}
	return g
}
