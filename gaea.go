package gaea

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/apus-run/gaea/registry"
)

type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

type Gaea struct {
	opts     *options
	ctx      context.Context
	cancel   func()
	mu       sync.Mutex
	instance *registry.ServiceInstance
}

// New create an application lifecycle manager.
func New(opts ...Option) *Gaea {
	o := Apply(opts...)

	ctx, cancel := context.WithCancel(o.ctx)
	return &Gaea{
		ctx:    ctx,
		cancel: cancel,
		opts:   o,
	}
}

// ID returns app instance id.
func (g *Gaea) ID() string { return g.opts.id }

// Name returns service name.
func (g *Gaea) Name() string { return g.opts.name }

// Version returns app version.
func (g *Gaea) Version() string { return g.opts.version }

// Metadata returns service metadata.
func (g *Gaea) Metadata() map[string]string { return g.opts.metadata }

// Endpoint returns endpoints.
func (g *Gaea) Endpoint() []string {
	if g.instance != nil {
		return g.instance.Endpoints
	}
	return nil
}

// Run executes all OnStart hooks registered with the application's Lifecycle.
func (g *Gaea) Run() error {
	// build service instance
	instance, err := g.buildInstance()
	if err != nil {
		return err
	}
	g.mu.Lock()
	g.instance = instance
	g.mu.Unlock()
	eg, ctx := errgroup.WithContext(NewContext(g.ctx, g))
	wg := sync.WaitGroup{}

	for _, fn := range g.opts.beforeStart {
		if err = fn(ctx); err != nil {
			return err
		}
	}

	for _, srv := range g.opts.servers {
		srv := srv
		eg.Go(func() error {
			<-ctx.Done() // wait for stop signal
			stopCtx, cancel := context.WithTimeout(NewContext(g.opts.ctx, g), g.opts.stopTimeout)
			defer cancel()
			return srv.Stop(stopCtx)
		})
		wg.Add(1)
		eg.Go(func() error {
			wg.Done()
			return srv.Start(ctx)
		})
	}
	wg.Wait()

	// register service
	if g.opts.registry != nil {
		c, cancel := context.WithTimeout(ctx, g.opts.registryTimeout)
		defer cancel()
		if err := g.opts.registry.Register(c, instance); err != nil {
			return err
		}
	}

	for _, fn := range g.opts.afterStart {
		if err = fn(ctx); err != nil {
			return err
		}
	}

	// watch signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, g.opts.sigs...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-quit:
			return g.Stop()
		}
	})
	if err := eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}

// Stop gracefully stops the application.
func (a *Gaea) Stop() (err error) {
	ctx := NewContext(a.ctx, a)
	for _, fn := range a.opts.beforeStop {
		err = fn(ctx)
	}

	// deregister instance
	a.mu.Lock()
	instance := a.instance
	a.mu.Unlock()
	if a.opts.registry != nil && instance != nil {
		ctx, cancel := context.WithTimeout(NewContext(a.ctx, a), a.opts.registryTimeout)
		defer cancel()
		if err := a.opts.registry.Deregister(ctx, instance); err != nil {
			return err
		}
	}

	// cancel app
	if a.cancel != nil {
		a.cancel()
	}
	return err
}

func (a *Gaea) buildInstance() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0, len(a.opts.endpoints))
	for _, e := range a.opts.endpoints {
		endpoints = append(endpoints, e.String())
	}
	if len(endpoints) == 0 {
		for _, srv := range a.opts.servers {
			url, err := srv.Endpoint()
			if err == nil {
				endpoints = append(endpoints, url.String())
			}
		}
	}
	return &registry.ServiceInstance{
		ID:        a.opts.id,
		Name:      a.opts.name,
		Version:   a.opts.version,
		Metadata:  a.opts.metadata,
		Endpoints: endpoints,
	}, nil
}

type appKey struct{}

// NewContext returns a new Context that carries value.
func NewContext(ctx context.Context, s AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, s)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (s AppInfo, ok bool) {
	s, ok = ctx.Value(appKey{}).(AppInfo)
	return
}
