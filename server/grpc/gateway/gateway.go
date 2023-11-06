package gateway

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	log "google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"

	"github.com/apus-run/gaea/internal/endpoint"
	"github.com/apus-run/gaea/internal/host"
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

func NewGateway(ctx context.Context, opts ...GatewayOption) *Gateway {
	g := ApplyGateway(opts...)

	gwmux, err := CreateGateway(
		ctx,
		g.conn,
		g.serveMuxOptions,
		g.annotators,
		g.registerServiceHandlers...,
	)

	if err != nil {
		log.Errorf("new gateway server error: %v", err)
	}

	g.mux = gwmux

	if g.Server == nil {
		g.Server = &http.Server{
			ReadHeaderTimeout: 5 * time.Second,  // read header timeout
			ReadTimeout:       5 * time.Second,  // read request timeout
			WriteTimeout:      10 * time.Second, // write timeout
			IdleTimeout:       20 * time.Second, // tcp idle time
		}
	}

	return g
}

// CreateGateway returns new grpc gateway
func CreateGateway(
	ctx context.Context,
	conn *grpc.ClientConn,
	serveMuxOptions []gwRuntime.ServeMuxOption,
	annotators []AnnotatorFunc,
	handlers ...HandlerFunc,
) (*gwRuntime.ServeMux, error) {
	// init gateway mux
	serveMuxOptions = append(serveMuxOptions, gwRuntime.WithErrorHandler(gwRuntime.DefaultHTTPErrorHandler))

	// init annotators
	for _, annotator := range annotators {
		serveMuxOptions = append(serveMuxOptions, gwRuntime.WithMetadata(annotator))
	}

	mux := gwRuntime.NewServeMux(serveMuxOptions...)

	if len(handlers) < 1 {
		return nil, errors.New("no handlers defined")
	}

	for _, handler := range handlers {
		err := handler(ctx, mux, conn)
		if err != nil {
			return nil, err
		}
	}

	return mux, nil
}

// CreateGatewayEndpoint returns new grpc gateway
func CreateGatewayEndpoint(
	ctx context.Context,
	endpoint string,
	serveMuxOptions []gwRuntime.ServeMuxOption,
	dialOptions []grpc.DialOption,
	annotators []AnnotatorFunc,
	handlers ...HandlerFromEndpoint,
) (*gwRuntime.ServeMux, error) {
	// init gateway mux
	serveMuxOptions = append(serveMuxOptions, gwRuntime.WithErrorHandler(gwRuntime.DefaultHTTPErrorHandler))

	// init annotators
	for _, annotator := range annotators {
		serveMuxOptions = append(serveMuxOptions, gwRuntime.WithMetadata(annotator))
	}

	mux := gwRuntime.NewServeMux(serveMuxOptions...)

	if len(handlers) < 1 {
		return nil, errors.New("no handlers defined")
	}

	for _, handler := range handlers {
		err := handler(ctx, mux, endpoint, dialOptions)
		if err != nil {
			return nil, err
		}
	}

	return mux, nil
}

func Run(ctx context.Context, opts ...GatewayOption) error {
	gate := ApplyGateway(opts...)

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzServer(gate.conn))

	gwmux, err := CreateGateway(
		ctx,
		gate.conn,
		gate.serveMuxOptions,
		gate.annotators,
		gate.registerServiceHandlers...,
	)

	if err != nil {
		return err
	}

	mux.Handle("/", gwmux)

	s := &http.Server{
		Addr:    gate.address,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	log.Infof("Starting listening at %s", gate.address)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}

func (g *Gateway) Endpoint() (*url.URL, error) {
	if err := g.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return g.endpoint, nil
}

// Start start the HTTP server.
func (g *Gateway) Start(ctx context.Context) error {
	if err := g.listenAndEndpoint(); err != nil {
		return err
	}

	log.Infof("[HTTP] server listening on: %s", g.lis.Addr().String())

	// http server and h2c handler
	// create a http mux
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/healthz", healthzServer(g.conn))
	httpMux.Handle("/", g.mux)

	g.Server.Addr = g.address
	g.Server.Handler = httpMux
	g.Server.RegisterOnShutdown(g.shutdownFunc)

	var err error
	if g.tlsConf != nil {
		err = g.ServeTLS(g.lis, "", "")
	} else {
		err = g.Serve(g.lis)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop stop the HTTP server.
func (g *Gateway) Stop(ctx context.Context) error {
	log.Info("[HTTP] server stopping")
	// disable keep-alives on existing connections
	g.Server.SetKeepAlivesEnabled(false)

	return g.Server.Shutdown(ctx)
}

func (g *Gateway) listenAndEndpoint() error {
	if g.lis == nil {
		lis, err := net.Listen(g.network, g.address)
		if err != nil {
			g.err = err
			return err
		}
		g.lis = lis
	}
	if g.endpoint == nil {
		addr, err := host.Extract(g.address, g.lis)
		if err != nil {
			g.err = err
			return err
		}
		g.endpoint = endpoint.NewEndpoint(endpoint.Scheme("http", g.tlsConf != nil), addr)
	}
	return g.err
}
