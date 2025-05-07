package routes

import (
	asrt "github.com/stretchr/testify/assert"
	"testing"
)

/*
func TestGetRouteFromRegex(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal("\\/api\\/v2\\/user-management\\/tenants\\/.*\\/disable",
		GetRouteFromRegex("/api/v2/user-management/tenants/{realmName}/disable"))
}
*/

func TestIsValidFromUrlPath(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal(true, IsValidFromUrlPath(""))
	assert.Equal(false, IsValidFromUrlPath("/pathFrom /"))
	assert.Equal(true, IsValidFromUrlPath("/pathFrom/path"))
}

func TestIsValidToUrlPath(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal(false, IsValidToUrlPath(""))
	assert.Equal(false, IsValidToUrlPath("/pathTo /"))
	assert.Equal(true, IsValidToUrlPath("/pathTo/path"))
}

func TestIsListedInForbiddenRoutes(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal(true, IsListedInForbiddenRoutes(""))
	assert.Equal(false, IsListedInForbiddenRoutes("/"))
}

/*
func TestGetRouteFromPrefix(t *testing.T) {
	assert := asrt.New(t)
	assert.Equal("/api/v1/details/", GetRouteFromPrefix("/api/v1/details/"))
	assert.Equal("/api/v1/details", GetRouteFromPrefix("/api/v1/details/{var}/some"))
}
*/
