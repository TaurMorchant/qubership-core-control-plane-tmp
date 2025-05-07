package tlsmode

import (
	"fmt"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/netcracker/qubership-core-lib-go/v3/utils"
	"strconv"
	"strings"
)

type Mode int

const (
	Disabled Mode = iota
	Preferred
)

var (
	log logging.Logger

	mode                        = Disabled
	gatewayCertificatesFilePath = utils.DefaultTlsPath

	internalSuffixLen = len("-internal")

	staticCoreServices = map[string]bool{
		"identity-provider": true,
		"tenant-manager":    true,
		"config-server":     true,
		"control-plane":     true,
		"site-management":   true,
		"paas-mediation":    true,
		"dbaas-agent":       true,
		"maas-agent":        true,
		"key-manager":       true,
	}
)

func init() {
	log = logging.GetLogger("tls")
}

// SetUpTlsProperties initializes this package variables (mode, gatewayCertificatesFilePath, etc).
// Must always be called before using any exported functions of the `tlsmode` package.
func SetUpTlsProperties() {
	internalTlsEnabled := configloader.GetOrDefaultString("internal.tls.enabled", "false")
	if strings.EqualFold("true", internalTlsEnabled) {
		mode = Preferred
		log.Infof("TLS mode is set to Preferred")
	} else {
		mode = Disabled
		log.Infof("TLS mode is set to Disabled")
	}
	gatewayCertificatesFilePath = configloader.GetOrDefaultString("gateway.certificate.file.path", utils.DefaultTlsPath)
}

func GatewayCertificatesFilePath() string {
	return gatewayCertificatesFilePath
}

func GetMode() Mode {
	return mode
}

type Proto int

const (
	Websocket Proto = iota
	Http
)

func UrlFromProperty(proto Proto, propertyName, defaultHostname string) string {
	propertyVal := configloader.GetOrDefaultString(propertyName, "")
	if propertyVal == "" {
		return BuildUrl(proto, defaultHostname)
	}
	if idx := strings.Index(propertyVal, "://"); idx != -1 {
		propertyVal = propertyVal[idx+3:]
	}
	hostAndPort := strings.Split(propertyVal, ":")
	if len(hostAndPort) == 1 {
		return BuildUrl(proto, hostAndPort[0])
	}
	port, err := strconv.Atoi(hostAndPort[1])
	if err != nil {
		log.Panicf("Could not parse custom port from property %s (%v):\n %v", propertyName, propertyVal, err)
	}
	return BuildUrl(proto, hostAndPort[0], port)
}

func BuildUrl(proto Proto, hostname string, customPort ...int) string {
	if proto == Websocket {
		return SelectByMode("ws://", "wss://") + hostname + ResolvePort(customPort...)
	} else if proto == Http {
		return SelectByMode("http://", "https://") + hostname + ResolvePort(customPort...)
	} else {
		log.Panicf("Unsupported proto for BuildUrl function: %v", proto)
		panic("unsupported proto for BuildUrl function")
	}
}

func ResolvePort(customPort ...int) string {
	if len(customPort) == 0 {
		return SelectByMode(":8080", ":8443")
	}
	return fmt.Sprintf(":%d", customPort[0])
}

func SelectByMode[T any](nonTlsValue, tlsValue T) T {
	if mode == Disabled {
		return nonTlsValue
	} else {
		return tlsValue
	}
}

// IsStaticCoreService indicates if service is hidden behind static-core-gateway,
// so its hostname needs to be modified in route configuration.
// Accepts endpoint address as an argument.
func IsStaticCoreService(hostname string) bool {
	_, found := staticCoreServices[hostname]
	return found
}

// AdaptHostname adapts new cloud-core microservices hostnames (e.g. control-plane-internal) to the old ones (e.g. control-plane),
// to avoid endpoint duplicates. For non-core services AdaptHostname returns hostname without changes.
func AdaptHostname(hostname string) string {
	if strings.HasSuffix(hostname, "-internal") {
		hostnameWithoutSuffix := hostname[:len(hostname)-internalSuffixLen]
		if IsStaticCoreService(hostnameWithoutSuffix) {
			return hostnameWithoutSuffix
		}
	}
	return hostname
}

// TransformHostRewrite transforms domain.Route#HostRewrite field for services hidden behind static-core-gateway,
// since envoy cache builder overrides upstream host for these services to avoid extra hops.
//
// For all the other services this is no-op.
func TransformHostRewrite(address string) string {
	if address == "" {
		return ""
	}
	parts := strings.Split(address, ":")
	if len(parts) == 1 { // address contains only hostname
		if IsStaticCoreService(address) {
			return address + "-internal"
		} else {
			return address
		}
	} else { // address contains both hostname and port
		if IsStaticCoreService(parts[0]) {
			return fmt.Sprintf("%s-internal:%s", parts[0], parts[1])
		} else {
			return address
		}
	}
}

func (m Mode) String() string {
	switch m {
	case Preferred:
		return "Preferred"
	case Disabled:
		return "Disabled"
	default:
		return fmt.Sprintf("<Unknown TLS Mode: %v>", int(mode))
	}
}
