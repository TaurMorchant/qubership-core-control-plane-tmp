package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRouteFromAddress_UrlWithVariables(t *testing.T) {
	testRouteFromRegex(t, "/api/v1/details", "")
	testRouteFromRegex(t, "/api/v1/{variable}/details", "/api/v1/([^/]+)/details(/.*)?")
	testRouteFromRegex(t, "/api/v1/{variable}/details/", "/api/v1/([^/]+)/details(/.*)?")
	testRouteFromRegex(t, "/api/v1/{variable}/path/{variable}", "/api/v1/([^/]+)/path/([^/]+)(/.*)?")
	testRouteFromRegex(t, "/api/v1/{variable}/path/{variable}/details", "/api/v1/([^/]+)/path/([^/]+)/details(/.*)?")
	testRouteFromRegex(t, "/api/v1/{tenantId}/**", "/api/v1/([^/]+)(/.*)?")
	testRouteFromRegex(t, "/api/v2/{microserviceName}/{dbClassifier}", "/api/v2/([^/]+)/([^/]+)(/.*)?")
}

func TestNewRouteFromAddress_UrlWithoutVariables(t *testing.T) {
	testRouteFromPrefix(t, "/api/v1/details", "/api/v1/details")
	testRouteFromPrefix(t, "/api/v1/details/some", "/api/v1/details/some")
	testRouteFromPrefix(t, "/api/v2", "/api/v2")
}

func testRouteFromRegex(t *testing.T, sourceUrl, expectedRegex string) {
	routeFromAddress := createFromAddress(sourceUrl)
	assert.Equal(t, expectedRegex, routeFromAddress.RouteFromRegex)
	assert.False(t, routeFromAddress.IsListedInForbiddenRoutes())
	assert.True(t, routeFromAddress.IsValidUrlPath())
}

func testRouteFromPrefix(t *testing.T, sourceUrl, expectedRegex string) {
	routeFromAddress := createFromAddress(sourceUrl)
	assert.Equal(t, expectedRegex, routeFromAddress.RouteFromPrefix)
	assert.False(t, routeFromAddress.IsListedInForbiddenRoutes())
	assert.True(t, routeFromAddress.IsValidUrlPath())
}

func createFromAddress(routeFromRaw string) *RouteFromAddress {
	return NewRouteFromAddress(routeFromRaw)
}
