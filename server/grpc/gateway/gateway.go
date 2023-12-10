package gateway

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	log "google.golang.org/grpc/grpclog"

	"github.com/apus-run/gaea/internal/endpoint"
	"github.com/apus-run/gaea/internal/host"
)

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
