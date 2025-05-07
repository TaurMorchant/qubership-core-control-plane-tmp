package common

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"strconv"
	"strings"
	"time"
)

type EnvoyProxyProperties struct {
	Routes      *RouteProperties
	Compression *CompressionProperties
	Tracing     *TracingProperties
	Googlere2   *GoogleRe2Properties
	Connection  *Connection
}

func NewEnvoyProxyProperties() *EnvoyProxyProperties {
	props := EnvoyProxyProperties{
		Routes: &RouteProperties{
			Timeout: readInt64PropOrDefault("envoy.proxy.routes.timeout", 120000),
		},
		Compression: &CompressionProperties{
			Enabled:         strings.EqualFold("true", configloader.GetKoanf().String("envoy.proxy.compression.enabled")),
			MinResponseSize: configloader.GetOrDefault("envoy.proxy.compression.min-response-size", 0).(int),
		},
		Tracing: &TracingProperties{},
		Googlere2: &GoogleRe2Properties{
			Maxsize:  configloader.GetOrDefaultString("envoy.proxy.googlere2.maxsize", "200"),
			WarnSize: configloader.GetOrDefaultString("envoy.proxy.googlere2.warnsize", "150"),
		},
		Connection: &Connection{
			PerConnectionBufferLimitMegabytes: int(readInt64PropOrDefault("envoy.proxy.connection.per.connection.buffer.limit.megabytes", 10)),
		},
	}

	if tracingEnabledRaw, found := getStringPropFromEnvOrFromYaml("tracing.enabled", "envoy.proxy.tracing.enabled"); found {
		props.Tracing.Enabled = strings.EqualFold("true", tracingEnabledRaw)
	}
	if zipkinCluster, found := getStringPropFromEnvOrFromYaml("zipkin.collector.cluster", "envoy.proxy.tracing.zipkin.collector_cluster"); found {
		props.Tracing.ZipkinCollectorCluster = zipkinCluster
	}
	if zipkinEndpoint, found := getStringPropFromEnvOrFromYaml("zipkin.collector.endpoint", "envoy.proxy.tracing.zipkin.collector_endpoint"); found {
		props.Tracing.ZipkinCollectorEndpoint = zipkinEndpoint
	}
	if tracingSamplingPercent, found := getStringPropFromEnvOrFromYaml("tracing.sampler.probabilistic", "envoy.proxy.tracing.overall_sampling"); found {
		floatTracingSamplingPercent, err := strconv.ParseFloat(tracingSamplingPercent, 64)
		if err != nil {
			logger.Errorf("Error parsing float64 property TRACING_SAMPLER_PROBABILISTIC from config: %v; default value will be used: 0.01", err)
		} else {
			props.Tracing.TracingSamplerProbabilisticValue = floatTracingSamplingPercent * 100.0
		}
	}

	mimeTypes := configloader.GetKoanf().String("envoy.proxy.compression.mime-types")
	props.Compression.SetMimeTypes(mimeTypes)
	return &props
}

func readInt64PropOrDefault(name string, defaultVal int64) int64 {
	if prop := configloader.GetOrDefaultString(name, ""); prop != "" {
		val, err := strconv.ParseInt(prop, 10, 64)
		if err != nil {
			logger.Errorf("Error parsing int64 property %v from config: %v; default value will be used: %v", name, err, defaultVal)
			return defaultVal
		}
		return val
	} else {
		return defaultVal
	}
}

func getStringPropFromEnvOrFromYaml(envName, localName string) (string, bool) {
	value := ""
	if envName != "" {
		value = configloader.GetOrDefaultString(envName, "")
	}
	if value == "" && localName != "" {
		value = configloader.GetOrDefaultString(localName, "")
	}
	return value, value != ""
}

type RouteProperties struct {
	Timeout int64
}

func (props *RouteProperties) GetTimeout() time.Duration {
	return time.Duration(props.Timeout) * time.Millisecond
}

type CompressionProperties struct {
	Enabled         bool
	MimeTypes       string
	MinResponseSize int
	MimeTypesList   []string
}

func (props *CompressionProperties) SetMimeTypes(mimeTypes string) {
	props.MimeTypes = mimeTypes
	props.MimeTypesList = strings.Split(mimeTypes, ",")
}

type TracingProperties struct {
	Enabled                          bool
	ZipkinCollectorCluster           string
	ZipkinCollectorEndpoint          string
	TracingSamplerProbabilisticValue float64
}

type GoogleRe2Properties struct {
	Maxsize  string
	WarnSize string
}

type Connection struct {
	PerConnectionBufferLimitMegabytes int
}

func (c *Connection) GetUInt32PerConnectionBufferLimitMBytes() *wrappers.UInt32Value {
	return &wrappers.UInt32Value{Value: uint32(c.PerConnectionBufferLimitMegabytes)}
}
