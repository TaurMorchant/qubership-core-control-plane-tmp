package format

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockRouteInfoProvider struct {
	from string
	to   string
}

func (mock MockRouteInfoProvider) GetTo() string {
	return mock.to
}

func (mock MockRouteInfoProvider) GetFrom() string {
	return mock.from
}

func Test_UrlWithVariables(t *testing.T) {
	mockRouteInfoProvider := MockRouteInfoProvider{"/api/v1/details/{var}/some/{some}", ""}
	routeAddr := NewRouteToAddress(mockRouteInfoProvider)
	assert.Equal(t, "\\/api\\/v1\\/details\\/.*\\/some\\/.*", routeAddr.GetRegexpRewrite())
	mockRouteInfoProvider.from = "/api/v1/details/{var}/some/"
	mockRouteInfoProvider.to = ""
	routeAddr = NewRouteToAddress(mockRouteInfoProvider)
	assert.Equal(t, "\\/api\\/v1\\/details\\/.*\\/some\\/", routeAddr.GetRegexpRewrite())
}
