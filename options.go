package gaea

import (
	"context"
	"net/url"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/apus-run/gaea/registry"
	"github.com/apus-run/gaea/server"
)

// Option is an application option.
type Option func(o *options)

// options is an application options.
type options struct {
	id        string
	name      string
	version   string
	metadata  map[string]string
	endpoints []*url.URL

	ctx  context.Context
	sigs []os.Signal

	registry        registry.Registry
	registryTimeout time.Duration
	stopTimeout     time.Duration
	servers         []server.Server

	// Before and After funcs
	beforeStart []func(context.Context) error
	beforeStop  []func(context.Context) error
	afterStart  []func(context.Context) error
	afterStop   []func(context.Context) error
}

// defaultOptions 初始化默认值
func defaultOptions() *options {
	return &options{
		ctx:             context.Background(),
		sigs:            []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		registryTimeout: 10 * time.Second,
		stopTimeout:     10 * time.Second,
	}
}

func Apply(opts ...Option) *options {
	o := defaultOptions()
	if id, err := uuid.NewUUID(); err == nil {
		o.id = id.String()
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// WithID with app id
func WithID(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// WithName .
func WithName(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// WithVersion with a version
func WithVersion(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

// WithContext with a context
func WithContext(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// WithSignal with some system signal
func WithSignal(sigs ...os.Signal) Option {
	return func(o *options) {
		o.sigs = sigs
	}
}

// WithMetadata with service metadata.
func WithMetadata(md map[string]string) Option {
	return func(o *options) {
		o.metadata = md
	}
}

// WithEndpoint with service endpoint.
func WithEndpoint(endpoints ...*url.URL) Option {
	return func(o *options) {
		o.endpoints = endpoints
	}
}

// WithRegistry with service registry.
func WithRegistry(r registry.Registry) Option {
	return func(o *options) {
		o.registry = r
	}
}

// WithServer with servers
func WithServer(srv ...server.Server) Option {
	return func(o *options) {
		o.servers = srv
	}
}

// WithRegistryTimeout with registrar timeout.
func WithRegistryTimeout(t time.Duration) Option {
	return func(o *options) {
		o.registryTimeout = t
	}
}

// WithStopTimeout with app stop timeout.
func WithStopTimeout(t time.Duration) Option {
	return func(o *options) {
		o.stopTimeout = t
	}
}

// Before and Afters

// BeforeStart run funcs before app starts
func BeforeStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStart = append(o.beforeStart, fn)
	}
}

// BeforeStop run funcs before app stops
func BeforeStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.beforeStop = append(o.beforeStop, fn)
	}
}

// AfterStart run funcs after app starts
func AfterStart(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStart = append(o.afterStart, fn)
	}
}

// AfterStop run funcs after app stops
func AfterStop(fn func(context.Context) error) Option {
	return func(o *options) {
		o.afterStop = append(o.afterStop, fn)
	}
}
