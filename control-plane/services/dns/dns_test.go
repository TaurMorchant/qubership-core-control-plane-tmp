package dns

import (
	clusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLookupFamily_UnsetEnv(t *testing.T) {
	_ = os.Unsetenv(dnsLookupFamilyEnv)
	DefaultLookupFamily = newLookupFamily()

	assert.Equal(t, clusterV3.Cluster_AUTO, DefaultLookupFamily.Get())
}

func TestLookupFamily_InvalidEnv(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, "invalid")
	defer os.Unsetenv(dnsLookupFamilyEnv)

	verifyPanic(t)
}

func TestLookupFamily_V6Preferred(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, v6Preferred)
	defer os.Unsetenv(dnsLookupFamilyEnv)

	DefaultLookupFamily = newLookupFamily()
	assert.Equal(t, clusterV3.Cluster_AUTO, DefaultLookupFamily.Get())
}

func TestLookupFamily_V4Preferred(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, v4Preferred)
	defer os.Unsetenv(dnsLookupFamilyEnv)

	DefaultLookupFamily = newLookupFamily()
	assert.Equal(t, clusterV3.Cluster_V4_PREFERRED, DefaultLookupFamily.Get())
}

func TestLookupFamily_V6Only(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, v6Only)
	defer os.Unsetenv(dnsLookupFamilyEnv)

	DefaultLookupFamily = newLookupFamily()
	assert.Equal(t, clusterV3.Cluster_V6_ONLY, DefaultLookupFamily.Get())
}

func TestLookupFamily_V4Only(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, v4Only)
	defer os.Unsetenv(dnsLookupFamilyEnv)

	DefaultLookupFamily = newLookupFamily()
	assert.Equal(t, clusterV3.Cluster_V4_ONLY, DefaultLookupFamily.Get())
}

func TestLookupFamily_All(t *testing.T) {
	os.Setenv(dnsLookupFamilyEnv, all)
	defer os.Unsetenv(dnsLookupFamilyEnv)
	
	DefaultLookupFamily = newLookupFamily()
	assert.Equal(t, clusterV3.Cluster_ALL, DefaultLookupFamily.Get())
}

func verifyPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.FailNow()
		}
	}()
	DefaultLookupFamily = newLookupFamily()
}
