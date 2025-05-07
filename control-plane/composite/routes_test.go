package composite

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetFallbackAddressWithTls(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	defer os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()

	srv := Service{
		mode:              SatelliteMode,
		coreBaseNamespace: "my-namespace",
	}
	addr := srv.getFallbackAddress("public-gateway-service")
	assert.Equal(t, "https", addr.GetProto())
	assert.Equal(t, int32(8443), addr.GetPort())
	assert.Equal(t, "public-gateway-service.my-namespace", addr.GetNamespacedMicroserviceHost())
}

func TestGetFallbackAddress(t *testing.T) {
	_ = os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()

	srv := Service{
		mode:              SatelliteMode,
		coreBaseNamespace: "my-namespace",
	}
	addr := srv.getFallbackAddress("public-gateway-service")
	assert.Equal(t, "http", addr.GetProto())
	assert.Equal(t, int32(8080), addr.GetPort())
	assert.Equal(t, "public-gateway-service.my-namespace", addr.GetNamespacedMicroserviceHost())
}
