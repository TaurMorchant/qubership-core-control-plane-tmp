package common

import (
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGetStringPropFromEnvOrFromYaml(t *testing.T) {
	os.Setenv("envoy.proxy.routes.test1", "100500")
	defer os.Unsetenv("envoy.proxy.routes.test1")

	os.Setenv("envoy.proxy.routes.test2", "500100")
	defer os.Unsetenv("envoy.proxy.routes.test2")

	os.Setenv("envoy.proxy.routes.test3", "")
	defer os.Unsetenv("envoy.proxy.routes.test3")

	configloader.Init(configloader.EnvPropertySource())

	envName := "envoy.proxy.routes.test1"
	localName := "envoy.proxy.routes.test2"
	value, isNotEmpty := getStringPropFromEnvOrFromYaml(envName, localName)
	assert.True(t, isNotEmpty)
	assert.Equal(t, "100500", value)

	envName = "envoy.proxy.routes.test4"
	localName = "envoy.proxy.routes.test2"
	value, isNotEmpty = getStringPropFromEnvOrFromYaml(envName, localName)
	assert.True(t, isNotEmpty)
	assert.Equal(t, "500100", value)

	envName = "envoy.proxy.routes.test3"
	localName = "envoy.proxy.routes.test3"
	value, isNotEmpty = getStringPropFromEnvOrFromYaml(envName, localName)
	assert.False(t, isNotEmpty)
	assert.Equal(t, "", value)
}

func TestReadInt64PropOrDefault(t *testing.T) {
	os.Setenv("envoy.proxy.routes.test1", "100500")
	defer os.Unsetenv("envoy.proxy.routes.test1")

	os.Setenv("envoy.proxy.routes.test2", "1a")
	defer os.Unsetenv("envoy.proxy.routes.test2")

	configloader.Init(configloader.EnvPropertySource())

	propName := "envoy.proxy.routes.test1"
	defaultVal := int64(10)

	result := readInt64PropOrDefault(propName, defaultVal)
	assert.Equal(t, int64(100500), result)

	propName = "envoy.proxy.routes.test2"
	result = readInt64PropOrDefault(propName, defaultVal)
	assert.Equal(t, defaultVal, result)

	propName = "envoy.proxy.routes.test3"
	result = readInt64PropOrDefault(propName, defaultVal)
	assert.Equal(t, defaultVal, result)
}

func TestInitEnvoyProxyProperties(t *testing.T) {
	os.Setenv("ENVOY_PROXY_ROUTES_TIMEOUT", "100500")
	defer os.Unsetenv("ENVOY_PROXY_ROUTES_TIMEOUT")

	os.Setenv("ENVOY_PROXY_COMPRESSION_ENABLED", "true")
	defer os.Unsetenv("ENVOY_PROXY_COMPRESSION_ENABLED")

	os.Setenv("ENVOY_PROXY_GOOGLERE2_WARNSIZE", "1")
	defer os.Unsetenv("ENVOY_PROXY_GOOGLERE2_WARNSIZE")

	os.Setenv("ENVOY_PROXY_GOOGLERE2_MAXSIZE", "2")
	defer os.Unsetenv("ENVOY_PROXY_GOOGLERE2_MAXSIZE")

	os.Setenv("TRACING_ENABLED", "true")
	defer os.Unsetenv("TRACING_ENABLED")

	os.Setenv("zipkin.collector.cluster", "cluster")
	defer os.Unsetenv("zipkin.collector.cluster")

	os.Setenv("zipkin.collector.endpoint", "endpoint")
	defer os.Unsetenv("zipkin.collector.endpoint")

	os.Setenv("envoy.proxy.compression.mime-types", "1,2,3,4,5")
	defer os.Unsetenv("envoy.proxy.compression.mime-types")

	os.Setenv("ENVOY_PROXY_CONNECTION_PER_CONNECTION_BUFFER_LIMIT_MEGABYTES", "100")
	defer os.Unsetenv("ENVOY_PROXY_CONNECTION_PER_CONNECTION_BUFFER_LIMIT_MEGABYTES")

	configloader.Init(configloader.EnvPropertySource())

	expEnvoyProxyProperties := &EnvoyProxyProperties{
		Routes: &RouteProperties{
			Timeout: int64(100500),
		},
		Compression: &CompressionProperties{
			Enabled:         true,
			MimeTypes:       "1,2,3,4,5",
			MinResponseSize: 0,
			MimeTypesList:   []string{"1", "2", "3", "4", "5"},
		},
		Tracing: &TracingProperties{
			Enabled:                 true,
			ZipkinCollectorCluster:  "cluster",
			ZipkinCollectorEndpoint: "endpoint",
		},
		Googlere2: &GoogleRe2Properties{
			Maxsize:  "2",
			WarnSize: "1",
		},
		Connection: &Connection{PerConnectionBufferLimitMegabytes: 100},
	}
	props := NewEnvoyProxyProperties()
	assert.Equal(t, expEnvoyProxyProperties, props)

	result := props.Connection.GetUInt32PerConnectionBufferLimitMBytes()
	assert.Equal(t, uint32(100), result.Value)
}

func TestRouteProperties_GetTimeout(t *testing.T) {
	type fields struct {
		Timeout int64
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{name: "millisecond", fields: fields{Timeout: 500}, want: time.Millisecond * 500},
		{name: "millisecond (default)", fields: fields{Timeout: 120000}, want: time.Second * 120},
		{name: "second", fields: fields{Timeout: 2000}, want: time.Second * 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := &RouteProperties{
				Timeout: tt.fields.Timeout,
			}
			if got := props.GetTimeout(); got != tt.want {
				t.Errorf("GetTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}
