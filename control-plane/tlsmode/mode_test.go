package tlsmode

import (
	"fmt"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDefaultTlsProperties(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	assert.Equal(t, Disabled, GetMode())
	assert.Equal(t, "/etc/tls", GatewayCertificatesFilePath())
}

func TestSetUpTlsProperties(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	defer os.Unsetenv("INTERNAL_TLS_ENABLED")
	_ = os.Setenv("GATEWAY_CERTIFICATE_FILE_PATH", "/custom-path")
	defer os.Unsetenv("GATEWAY_CERTIFICATE_FILE_PATH")

	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	assert.Equal(t, Preferred, GetMode())
	assert.Equal(t, "/custom-path", GatewayCertificatesFilePath())
}

func TestUrlFromProperty_TlsEnabled(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	_ = os.Setenv("TEST_MS_URL", "http://some-service:8888")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	defer func() {
		_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
		_ = os.Unsetenv("TEST_MS_URL")
		configloader.Init(configloader.EnvPropertySource())
		SetUpTlsProperties()
	}()

	assert.Equal(t, "https://some-service:8888", UrlFromProperty(Http, "test.ms.url", "my-service.ns"))
	assert.Equal(t, "https://my-service.ns:8443", UrlFromProperty(Http, "test.ms.url.nonexistent", "my-service.ns"))
	assert.Equal(t, "wss://some-service:8888", UrlFromProperty(Websocket, "test.ms.url", "my-service.ns"))
	assert.Equal(t, "wss://my-service.ns:8443", UrlFromProperty(Websocket, "test.ms.url.nonexistent", "my-service.ns"))
}

func TestUrlFromProperty_TlsDisabled(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	_ = os.Setenv("TEST_MS_URL", "http://some-service:8888")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	defer func() {
		_ = os.Unsetenv("TEST_MS_URL")
		configloader.Init(configloader.EnvPropertySource())
	}()

	assert.Equal(t, "http://some-service:8888", UrlFromProperty(Http, "test.ms.url", "my-service.ns"))
	assert.Equal(t, "http://my-service.ns:8080", UrlFromProperty(Http, "test.ms.url.nonexistent", "my-service.ns"))
	assert.Equal(t, "ws://some-service:8888", UrlFromProperty(Websocket, "test.ms.url", "my-service.ns"))
	assert.Equal(t, "ws://my-service.ns:8080", UrlFromProperty(Websocket, "test.ms.url.nonexistent", "my-service.ns"))
}

func TestBuildUrl_TlsEnabled(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	defer func() {
		_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
		configloader.Init(configloader.EnvPropertySource())
		SetUpTlsProperties()
	}()

	assert.Equal(t, "https://my-service.ns:12345", BuildUrl(Http, "my-service.ns", 12345))
	assert.Equal(t, "https://my-service.ns:8443", BuildUrl(Http, "my-service.ns"))
	assert.Equal(t, "wss://my-service.ns:12345", BuildUrl(Websocket, "my-service.ns", 12345))
	assert.Equal(t, "wss://my-service.ns:8443", BuildUrl(Websocket, "my-service.ns"))
}

func TestBuildUrl_TlsDisabled(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()

	assert.Equal(t, "http://my-service.ns:12345", BuildUrl(Http, "my-service.ns", 12345))
	assert.Equal(t, "http://my-service.ns:8080", BuildUrl(Http, "my-service.ns"))
	assert.Equal(t, "ws://my-service.ns:12345", BuildUrl(Websocket, "my-service.ns", 12345))
	assert.Equal(t, "ws://my-service.ns:8080", BuildUrl(Websocket, "my-service.ns"))
}

func TestResolvePort_TlsEnabled(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	defer func() {
		_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
		configloader.Init(configloader.EnvPropertySource())
		SetUpTlsProperties()
	}()

	assert.Equal(t, ":12345", ResolvePort(12345))
	assert.Equal(t, ":8443", ResolvePort())
}

func TestResolvePort_TlsDisabled(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()

	assert.Equal(t, ":12345", ResolvePort(12345))
	assert.Equal(t, ":8080", ResolvePort())
}

func TestSelectByMode_TlsEnabled(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()
	defer func() {
		_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
		configloader.Init(configloader.EnvPropertySource())
		SetUpTlsProperties()
	}()

	assert.Equal(t, "tls-val", SelectByMode("non-tls-val", "tls-val"))
}

func TestSelectByMode_TlsDisabled(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	SetUpTlsProperties()

	assert.Equal(t, "non-tls-val", SelectByMode("non-tls-val", "tls-val"))
}

func TestTlsModeString(t *testing.T) {
	val := fmt.Sprintf("%v", Preferred)
	assert.Equal(t, "Preferred", val)
	val = fmt.Sprintf("%v", Disabled)
	assert.Equal(t, "Disabled", val)
}

func TestIsStaticCoreService(t *testing.T) {
	staticGwServices := []string{
		"identity-provider",
		"key-manager",
		"tenant-manager",
		"config-server",
		"control-plane",
		"site-management",
		"paas-mediation",
		"dbaas-agent",
		"maas-agent",
	}
	for _, srv := range staticGwServices {
		assert.True(t, IsStaticCoreService(srv))
	}
	nonStaticGwServices := []string{
		"gateway-auth-extension",
		"domain-resolver-frontend",
		"cloud-administrator-frontend",
		"order-manager",
		"order-management",
		"customer-management",
		"static-core-gateway",
		"install-base",
		"samples-repository",
	}
	for _, srv := range nonStaticGwServices {
		assert.False(t, IsStaticCoreService(srv))
	}
}

func TestTransformHostRewrite(t *testing.T) {
	argumentToResultMap := map[string]string{
		"":                             "",
		"identity-provider:8080":       "identity-provider-internal:8080",
		"key-manager:8443":             "key-manager-internal:8443",
		"tenant-manager":               "tenant-manager-internal",
		"config-server":                "config-server-internal",
		"control-plane:8443":           "control-plane-internal:8443",
		"site-management":              "site-management-internal",
		"paas-mediation:8080":          "paas-mediation-internal:8080",
		"dbaas-agent":                  "dbaas-agent-internal",
		"maas-agent":                   "maas-agent-internal",
		"gateway-auth-extension":       "gateway-auth-extension",
		"domain-resolver-frontend":     "domain-resolver-frontend",
		"cloud-administrator-frontend": "cloud-administrator-frontend",
		"order-manager:8080":           "order-manager:8080",
		"order-management:8443":        "order-management:8443",
		"customer-management:8080":     "customer-management:8080",
		"static-core-gateway:8443":     "static-core-gateway:8443",
		"install-base":                 "install-base",
		"samples-repository":           "samples-repository",
	}
	for originalAddr, result := range argumentToResultMap {
		assert.Equal(t, result, TransformHostRewrite(originalAddr))
	}
}

func TestAdaptHostname(t *testing.T) {
	assert.Equal(t, "test-service", AdaptHostname("test-service"))
	assert.Equal(t, "test-service-internal", AdaptHostname("test-service-internal"))

	assert.Equal(t, "control-plane", AdaptHostname("control-plane"))
	assert.Equal(t, "control-plane", AdaptHostname("control-plane-internal"))
	assert.Equal(t, "paas-mediation", AdaptHostname("paas-mediation"))
	assert.Equal(t, "paas-mediation", AdaptHostname("paas-mediation-internal"))
	assert.Equal(t, "tenant-manager", AdaptHostname("tenant-manager"))
	assert.Equal(t, "tenant-manager", AdaptHostname("tenant-manager-internal"))
	assert.Equal(t, "config-server", AdaptHostname("config-server"))
	assert.Equal(t, "config-server", AdaptHostname("config-server-internal"))
	assert.Equal(t, "site-management", AdaptHostname("site-management"))
	assert.Equal(t, "site-management", AdaptHostname("site-management-internal"))
	assert.Equal(t, "dbaas-agent", AdaptHostname("dbaas-agent"))
	assert.Equal(t, "dbaas-agent", AdaptHostname("dbaas-agent-internal"))
	assert.Equal(t, "maas-agent", AdaptHostname("maas-agent"))
	assert.Equal(t, "maas-agent", AdaptHostname("maas-agent-internal"))
	assert.Equal(t, "key-manager", AdaptHostname("key-manager"))
	assert.Equal(t, "key-manager", AdaptHostname("key-manager-internal"))
	assert.Equal(t, "identity-provider", AdaptHostname("identity-provider"))
	assert.Equal(t, "identity-provider", AdaptHostname("identity-provider-internal"))
}
