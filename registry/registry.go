package registry

import (
	"context"
)

// Registry is service registrar.
type Registry interface {
	Register(ctx context.Context, svc *ServiceInstance) error
	Deregister(ctx context.Context, svc *ServiceInstance) error
}

// Discovery is service discovery.
type Discovery interface {
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	GetServiceList(ctx context.Context) ([]*ServiceInstance, error)
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// Watcher is service watcher.
type Watcher interface {
	// Next returns services in the following two cases:
	// 1.the first time to watch and the service instance list is not empty.
	// 2.any service instance changes found.
	// if the above two conditions are not met, it will block until context deadline exceeded or canceled
	Next() ([]*ServiceInstance, error)
	// Stop close the watcher.
	Stop() error
}

// ServiceInstance is an instance of a service in a discovery system.
type ServiceInstance struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// Endpoints is endpoint addresses of the service instance.
	// schema:
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
}

// NoopRegistry is an empty implement of Registry
var NoopRegistry Registry = &noopRegistry{}

// NoopRegistry
type noopRegistry struct{}

// Deregister implements Registry.
func (*noopRegistry) Deregister(ctx context.Context, svc *ServiceInstance) error {
	return nil
}

// Register implements Registry.
func (*noopRegistry) Register(ctx context.Context, svc *ServiceInstance) error {
	return nil
}
