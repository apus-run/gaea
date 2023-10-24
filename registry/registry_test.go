package registry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoopRegistry(t *testing.T) {
	reg := noopRegistry{}

	assert.Nil(t, reg.Deregister(context.Background(), &ServiceInstance{}))
	assert.Nil(t, reg.Register(context.Background(), &ServiceInstance{}))
}
