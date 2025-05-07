package action

import (
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToTypeURL(t *testing.T) {
	resultType := EnvoyCluster.ToTypeURL()
	assert.Equal(t, resource.ClusterType, resultType)

	resultType = EnvoyListener.ToTypeURL()
	assert.Equal(t, resource.ListenerType, resultType)

	resultType = EnvoyRouteConfig.ToTypeURL()
	assert.Equal(t, resource.RouteType, resultType)
}
