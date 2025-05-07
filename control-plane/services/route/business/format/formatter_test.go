package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouteFormatter_IsRouteAllowedIsRouteAllowed(t *testing.T) {
	assert.False(t, DefaultRouteFormatter.IsRouteAllowed(""))
	assert.True(t, DefaultRouteFormatter.IsRouteAllowed("/api"))
}

func TestRouteFormatter_GetRoutePropertyKey(t *testing.T) {
	assert.Equal(t, "/api/v1/**", DefaultRouteFormatter.GetRoutePropertyKey("/api/v1"))
	assert.Equal(t, "/api/v1/*/**", DefaultRouteFormatter.GetRoutePropertyKey("/api/v1/{tenantId}"))
	assert.Equal(t, "/api/v1/*/*/**", DefaultRouteFormatter.GetRoutePropertyKey("/api/v1/{microserviceName}/{dbClassifier}"))
	assert.Equal(t, "/api/v1/*/activate/**", DefaultRouteFormatter.GetRoutePropertyKey("/api/v1/{tenantId}/activate"))
	assert.Equal(t, "/api/v1/*/*/**", DefaultRouteFormatter.GetRoutePropertyKey("/api/v1/{microserviceName}/{dbClassifier}"))
}

func TestRouteFormatter_GetFromWithoutVariable(t *testing.T) {
	assert.Equal(t, "/api", DefaultRouteFormatter.GetFromWithoutVariable("/api/{var}/tenant-manager/{tenant}"))
	assert.Equal(t, "/api/var/tenant-manager", DefaultRouteFormatter.GetFromWithoutVariable("/api/var/tenant-manager"))
}
