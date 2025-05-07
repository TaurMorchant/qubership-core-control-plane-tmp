package msaddr

import (
	asrt "github.com/stretchr/testify/assert"
	"testing"
)

func TestGetTargetHost(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("http://some-host:8080", "some-namespace")
	assert.Equal("some-host.some-namespace", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("some-host", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(8080), msAddress.GetPort())
}

func TestCoreInternalHost(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("http://identity-provider:8080", "some-namespace")
	assert.Equal("identity-provider.some-namespace", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("identity-provider", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(8080), msAddress.GetPort())

	msAddress = NewMicroserviceAddress("http://access-control-internal:8080", "some-namespace")
	assert.Equal("access-control-internal.some-namespace", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("access-control-internal", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(8080), msAddress.GetPort())
}

func TestGetTargetHostWithoutPort(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("http://some-host", "default")
	assert.Equal("some-host", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("some-host", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(80), msAddress.GetPort())
}

func TestGetTargetHostWithLocalDevNamespace(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("https://some-host",
		"192.168.56.101"+LocalDevNamespacePostfix)
	assert.Equal("192.168.56.101"+LocalDevNamespacePostfix,
		msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("some-host", msAddress.GetMicroserviceName())
	assert.Equal("https", msAddress.GetProto())
	assert.Equal(int32(443), msAddress.GetPort())
}

func TestPort(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("http://some-host:8080", "some-namespace")
	assert.Equal("some-host.some-namespace", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("some-host", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(8080), msAddress.GetPort())
}

func TestDefaultProto(t *testing.T) {
	assert := asrt.New(t)
	msAddress := NewMicroserviceAddress("some-host:8080", "some-namespace")
	assert.Equal("some-host.some-namespace", msAddress.GetNamespacedMicroserviceHost())
	assert.Equal("some-host", msAddress.GetMicroserviceName())
	assert.Equal("http", msAddress.GetProto())
	assert.Equal(int32(8080), msAddress.GetPort())
}
