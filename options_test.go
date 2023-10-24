package gaea

import (
	"context"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/apus-run/gaea/registry"
)

func TestID(t *testing.T) {
	o := &options{}
	i := "123"
	WithID(i)(o)

	if !reflect.DeepEqual(i, o.id) {
		t.Fatalf("o.id:%s is not equal to v:%s", o.id, i)
	}
}

func TestName(t *testing.T) {
	o := &options{}
	n := "user-srv"
	WithName(n)(o)
	if !reflect.DeepEqual(n, o.name) {
		t.Fatalf("o.name:%s is not equal to v:%s", o.name, n)
	}
}

func TestVersion(t *testing.T) {
	o := &options{}
	v := "v1.0.0"
	WithVersion(v)(o)
	if !reflect.DeepEqual(v, o.version) {
		t.Fatalf("o.version:%s is not equal to v:%s", o.version, v)
	}
}

func TestMetadata(t *testing.T) {
	o := &options{}
	v := map[string]string{
		"a": "1",
		"b": "2",
	}
	WithMetadata(v)(o)
	if !reflect.DeepEqual(v, o.metadata) {
		t.Fatalf("o.metadata:%s is not equal to v:%s", o.metadata, v)
	}
}

func TestEndpoint(t *testing.T) {
	o := &options{}
	v := []*url.URL{
		{Host: "example.com"},
		{Host: "foo.com"},
	}
	WithEndpoint(v...)(o)
	if !reflect.DeepEqual(v, o.endpoints) {
		t.Fatalf("o.endpoints:%s is not equal to v:%s", o.endpoints, v)
	}
}

func TestContext(t *testing.T) {
	type ctxKey = struct{}
	o := &options{}
	v := context.WithValue(context.TODO(), ctxKey{}, "b")
	WithContext(v)(o)
	if !reflect.DeepEqual(v, o.ctx) {
		t.Fatalf("o.ctx:%s is not equal to v:%s", o.ctx, v)
	}
}

func TestNewOptions(t *testing.T) {
	i := "123"
	n := "user-srv"
	v := "v1.0.0"

	opts := Apply(WithID(i), WithName(n), WithVersion(v))
	t.Logf("options: %v \n", opts)
}

type mockSignal struct{}

func (m *mockSignal) String() string { return "sig" }
func (m *mockSignal) Signal()        {}

func TestSignal(t *testing.T) {
	o := &options{}
	v := []os.Signal{
		&mockSignal{}, &mockSignal{},
	}
	WithSignal(v...)(o)
	if !reflect.DeepEqual(v, o.sigs) {
		t.Fatal("o.sigs is not equal to v")
	}
}

type mockRegistrar struct{}

func (m *mockRegistrar) Register(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}

func (m *mockRegistrar) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	return nil
}

func TestRegistrar(t *testing.T) {
	o := &options{}
	v := &mockRegistrar{}
	WithRegistry(v)(o)
	if !reflect.DeepEqual(v, o.registry) {
		t.Fatal("o.registrar is not equal to v")
	}
}

func TestRegistrarTimeout(t *testing.T) {
	o := &options{}
	v := time.Duration(123)
	WithRegistryTimeout(v)(o)
	if !reflect.DeepEqual(v, o.registryTimeout) {
		t.Fatal("o.registrarTimeout is not equal to v")
	}
}

func TestStopTimeout(t *testing.T) {
	o := &options{}
	v := time.Duration(123)
	WithStopTimeout(v)(o)
	if !reflect.DeepEqual(v, o.stopTimeout) {
		t.Fatal("o.stopTimeout is not equal to v")
	}
}

func TestBeforeStart(t *testing.T) {
	o := &options{}
	v := func(_ context.Context) error {
		t.Log("BeforeStart...")
		return nil
	}
	BeforeStart(v)(o)
}

func TestBeforeStop(t *testing.T) {
	o := &options{}
	v := func(_ context.Context) error {
		t.Log("BeforeStop...")
		return nil
	}
	BeforeStop(v)(o)
}

func TestAfterStart(t *testing.T) {
	o := &options{}
	v := func(_ context.Context) error {
		t.Log("AfterStart...")
		return nil
	}
	AfterStart(v)(o)
}

func TestAfterStop(t *testing.T) {
	o := &options{}
	v := func(_ context.Context) error {
		t.Log("AfterStop...")
		return nil
	}
	AfterStop(v)(o)
}
