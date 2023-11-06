package grpc

import (
	"context"
	"net/http"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	log "google.golang.org/grpc/grpclog"
)

type Gateway struct {
	endpoint string

	conn                    *grpc.ClientConn
	mux                     []gwruntime.ServeMuxOption
	registerServiceHandlers []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error
}

// GatewayOption is gRPC Gateway option.
type GatewayOption func(o *Gateway)

// WithGatewayConn with gateway network.
func WithGatewayConn(conn *grpc.ClientConn) GatewayOption {
	return func(g *Gateway) {
		g.conn = conn
	}
}

// WithGatewayEndpoint with gateway address.
func WithGatewayEndpoint(endpoint string) GatewayOption {
	return func(g *Gateway) {
		g.endpoint = endpoint
	}
}

// defaultGateway return a default config server
func defaultGateway() *Gateway {
	return &Gateway{
		endpoint: ":0",
	}
}

func ApplyGateway(opts ...GatewayOption) *Gateway {
	c := defaultGateway()
	for _, o := range opts {
		o(c)
	}
	return c
}

func (g *Gateway) Run(ctx context.Context, opts ...GatewayOption) error {
	gate := ApplyGateway(opts...)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		if err := gate.conn.Close(); err != nil {
			log.Errorf("Failed to close a client connection to the gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	gwmux, err := gate.newGateway(ctx, gate.conn)

	if err != nil {
		return err
	}

	mux.Handle("/", gwmux)

	s := &http.Server{
		Addr:    g.endpoint,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	log.Infof("Starting listening at %s", gate.endpoint)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}

// newGateway returns a new gateway server which translates HTTP into gRPC.
func (g *Gateway) newGateway(ctx context.Context, conn *grpc.ClientConn) (http.Handler, error) {
	mux := gwruntime.NewServeMux(g.mux...)

	for _, f := range g.registerServiceHandlers {
		if err := f(ctx, mux, conn); err != nil {
			return nil, err
		}
	}

	return mux, nil
}
