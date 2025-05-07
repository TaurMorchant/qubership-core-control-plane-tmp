package routes

import (
	asrt "github.com/stretchr/testify/assert"
	"testing"
)

func TestIsRouteAllowed(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal(false, IsRouteAllowed(""))
	assert.Equal(true, IsRouteAllowed("/some/route"))
}

func TestGetRoutePropertyKey(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal("/api/v1/**", getRoutePropertyKey("/api/v1"))
	assert.Equal("/api/v1/*/**", getRoutePropertyKey("/api/v1/{tenantId}"))
	assert.Equal("/api/v1/*/**", getRoutePropertyKey("/api/v1/{tenantId}/**"))
	assert.Equal("/api/v1/*/*/**", getRoutePropertyKey("/api/v1/{microserviceName}/{dbClassifier}"))
	assert.Equal("/api/v1/*/activate/**", getRoutePropertyKey("/api/v1/{tenantId}/activate"))
}

/*
func TestGenerateRoutes(t *testing.T) {
	assert := asrt.New(t)
	testSuite := []string{
		"/api/v1", "", "/api/v1",
		"", "\\/api\\/v1\\/.*", "/api/v1/{tenantId}",
		"", "\\/api\\/v1\\/.*", "/api/v1/{tenantId}/**",
		"", "\\/api\\/v2\\/.*\\/.*", "/api/v2/{microserviceName}/{dbClassifier}",
		"", "\\/api\\/v1\\/.*\\/activate", "/api/v1/{tenantId}/activate",
		"/api/v2/some/path", "", "/api/v2/some/path",
		"/", "", "/",
	}
	for i := 0; i < len(testSuite); i += 3 {
		actualPrefix, actualRegex := FormatRoutes(testSuite[i+2])
		assert.Equal(testSuite[i], actualPrefix)
		assert.Equal(testSuite[i+1], actualRegex)
	}
}
*/
