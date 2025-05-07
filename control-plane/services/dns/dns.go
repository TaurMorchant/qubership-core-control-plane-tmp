package dns

import (
	clusterV3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"os"
	"strings"
)

var (
	DefaultLookupFamily = newLookupFamily()
	logger              = logging.GetLogger("dns-service")
)

type lookupFamily struct {
	value clusterV3.Cluster_DnsLookupFamily
}

func newLookupFamily() *lookupFamily {
	return &lookupFamily{value: getFromEnv()}
}

func (f *lookupFamily) Get() clusterV3.Cluster_DnsLookupFamily {
	return f.value
}

const (
	dnsLookupFamilyEnv = "DNS_LOOKUP_FAMILY"

	v4Only      = "V4_ONLY"
	v6Only      = "V6_ONLY"
	v4Preferred = "V4_PREFERRED"
	v6Preferred = "V6_PREFERRED"
	all         = "ALL"
)

func getFromEnv() clusterV3.Cluster_DnsLookupFamily {
	envVal := os.Getenv(dnsLookupFamilyEnv)
	if strings.TrimSpace(envVal) == "" {
		return clusterV3.Cluster_AUTO
	}
	switch envVal {
	case v4Only:
		return clusterV3.Cluster_V4_ONLY
	case v6Only:
		return clusterV3.Cluster_V6_ONLY
	case v4Preferred:
		return clusterV3.Cluster_V4_PREFERRED
	case v6Preferred:
		return clusterV3.Cluster_AUTO
	case all:
		return clusterV3.Cluster_ALL
	}
	logger.Panicf("Unsupported %s environment variable value: '%s'", dnsLookupFamilyEnv, envVal)
	return clusterV3.Cluster_AUTO
}
