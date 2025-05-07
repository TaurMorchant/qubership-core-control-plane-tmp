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
