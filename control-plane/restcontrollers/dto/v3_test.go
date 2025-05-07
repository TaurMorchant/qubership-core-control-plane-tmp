package dto

import (
	"database/sql"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	name     string
	endpoint RawEndpoint
	wantHost string
	wantPort int
}

func TestToString(t *testing.T) {
	var emptyInt32Value *int32
	result := toString(emptyInt32Value)
	assert.Equal(t, "<nil>", result)

	var int32Value = int32(0)
	result = toString(&int32Value)
	assert.Equal(t, "0", result)

	var emptyUint32Value *uint32
	result = toString(emptyUint32Value)
	assert.Equal(t, "<nil>", result)

	var uint32Value = uint32(0)
	result = toString(&uint32Value)
	assert.Equal(t, "0", result)

	var emptyInt64Value *int64
	result = toString(emptyInt64Value)
	assert.Equal(t, "<nil>", result)

	var int64Value = int64(0)
	result = toString(&int64Value)
	assert.Equal(t, "0", result)

	result = toString(nil)
	assert.Equal(t, "<nil>", result)

	var stringValue = "test"
	result = toString(stringValue)
	assert.Equal(t, fmt.Sprintf("%v", "test"), result)
}

func TestString_shouldReturnNil_whenEmptyObject(t *testing.T) {
	var activeDCsV3 *ActiveDCsV3 = nil
	result := activeDCsV3.String()
	assert.Equal(t, "<nil>", result)

	var activeDCsHealthCheckV3 *ActiveDCsHealthCheckV3 = nil
	result = activeDCsHealthCheckV3.String()
	assert.Equal(t, "<nil>", result)

	var activeDCsRetryPolicyV3 *ActiveDCsRetryPolicyV3 = nil
	result = activeDCsRetryPolicyV3.String()
	assert.Equal(t, "<nil>", result)

	var activeDCsRetryBackOffV3 *ActiveDCsRetryBackOffV3 = nil
	result = activeDCsRetryBackOffV3.String()
	assert.Equal(t, "<nil>", result)

	var activeDCsCommonLbConfigV3 *ActiveDCsCommonLbConfigV3 = nil
	result = activeDCsCommonLbConfigV3.String()
	assert.Equal(t, "<nil>", result)
}

func TestString_shouldCorrectValue_whenObjectNotEmpty(t *testing.T) {
	activeDCsV3 := &ActiveDCsV3{
		Protocol:  "Protocol",
		HttpPort:  func() *int32 { i := int32(0); return &i }(),
		HttpsPort: func() *int32 { i := int32(0); return &i }(),
	}
	result := activeDCsV3.String()
	assert.Equal(t, "ActiveDCsV3{protocol='Protocol',httpPort=0,httpsPort=0,publicGwHosts=[],privateGwHosts=[],healthCheck=<nil>,retryPolicy=<nil>,commonLbConfig=<nil>}", result)

	activeDCsHealthCheckV3 := &ActiveDCsHealthCheckV3{
		Timeout:  0,
		Interval: 0,
	}
	result = activeDCsHealthCheckV3.String()
	assert.Equal(t, "{timeout=0,interval=0,noTrafficInterval=<nil>,unhealthyThreshold=<nil>,unhealthyInterval=<nil>,healthyThreshold=<nil>}", result)

	activeDCsRetryPolicyV3 := &ActiveDCsRetryPolicyV3{
		RetryOn:    "test",
		NumRetries: uint32(0),
	}
	result = activeDCsRetryPolicyV3.String()
	assert.Equal(t, "{retryOn='test',numRetries=0,perTryTimeout=<nil>,retryBackOff=<nil>,retriableStatusCodes=[]}", result)

	activeDCsRetryBackOffV3 := &ActiveDCsRetryBackOffV3{
		BaseInterval: int64(0),
		MaxInterval:  int64(0),
	}
	result = activeDCsRetryBackOffV3.String()
	assert.Equal(t, "{baseInterval='0',maxInterval=0}", result)

	activeDCsCommonLbConfigV3 := &ActiveDCsCommonLbConfigV3{
		HealthyPanicThreshold: float64(0),
	}
	result = activeDCsCommonLbConfigV3.String()
	assert.Equal(t, "{healthyPanicThreshold='0.000000'}", result)
}

func TestRawEndpoint_HostPort(t *testing.T) {
	tests := []testStruct{
		{name: "http no port", endpoint: "http://test.t", wantHost: "test.t", wantPort: 80},
		{name: "https no port", endpoint: "https://test.t", wantHost: "test.t", wantPort: 443},
		{name: "with port no scheme", endpoint: "test.t:8080", wantHost: "test.t", wantPort: 8080},
		{name: "with port with scheme", endpoint: "http://test.t:8080", wantHost: "test.t", wantPort: 8080},
		{name: "with port with scheme with path", endpoint: "http://test.t:8080/cde?foo=bar", wantHost: "test.t", wantPort: 8080},
		{name: "with port no scheme with path", endpoint: "test.t:8080/cde?foo=bar", wantHost: "test.t", wantPort: 8080},
		{name: "with port 65535 no scheme with path", endpoint: "test.t:65535/cde?foo=bar", wantHost: "test.t", wantPort: 65535},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, _ := tt.endpoint.HostPort()
			if gotHost != tt.wantHost {
				t.Errorf("HostPort() got host = %v, want host %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("HostPort() got port = %v, want port %v", gotPort, tt.wantPort)
			}
		})
	}
}

func TestRawEndpoint_HostPort_Error(t *testing.T) {
	tests := []testStruct{
		{name: "negative port", endpoint: "http://test.t:-1", wantHost: "", wantPort: 0},
		{name: "port with letter", endpoint: "https://test.t:1q00", wantHost: "", wantPort: 0},
		{name: "port exceeds the maximum value", endpoint: "https://test.t:65536", wantHost: "", wantPort: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotPort, err := tt.endpoint.HostPort()
			if err == nil {
				t.Error("Must be a error, but error is nil")
			}
			fmt.Printf("Got error %v \n", err)
			if gotHost != tt.wantHost {
				t.Errorf("HostPort() got host = %v, want host %v", gotHost, tt.wantHost)
			}
			if gotPort != tt.wantPort {
				t.Errorf("HostPort() got port = %v, want port %v", gotPort, tt.wantPort)
			}
		})
	}
}

func TestStatefulSession_IsDeleteRequest(t *testing.T) {
	trueVal := true
	falseVal := false
	spec := StatefulSession{}
	assert.True(t, spec.IsDeleteRequest())
	spec = StatefulSession{Enabled: &falseVal}
	assert.False(t, spec.IsDeleteRequest())
	spec = StatefulSession{Cookie: &Cookie{}}
	assert.False(t, spec.IsDeleteRequest())
	spec = StatefulSession{Cookie: &Cookie{}}
	assert.False(t, spec.IsDeleteRequest())
	spec = StatefulSession{Cookie: &Cookie{}, Enabled: &trueVal}
	assert.False(t, spec.IsDeleteRequest())
}

func TestStatefulSession_ToRouteStatefulSession(t *testing.T) {
	trueVal := true
	falseVal := false
	ttlVal := int64(1000)

	spec := StatefulSession{}
	session := spec.ToRouteStatefulSession("internal-gateway-service")
	assert.Nil(t, session)

	spec = StatefulSession{Enabled: &falseVal}
	session = spec.ToRouteStatefulSession("internal-gateway-service")
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "internal-gateway-service", session.Gateways[0])
	assert.Empty(t, session.CookieName)
	assert.Empty(t, session.CookiePath)
	assert.Nil(t, session.CookieTtl)
	assert.False(t, session.Enabled)

	spec = StatefulSession{Enabled: &trueVal}
	session = spec.ToRouteStatefulSession("internal-gateway-service")
	assert.Nil(t, session)

	spec = StatefulSession{Cookie: &Cookie{
		Name: "sticky-cookie",
		Ttl:  &ttlVal,
		Path: domain.NullString{NullString: sql.NullString{
			String: "/",
			Valid:  true,
		}},
	}}
	session = spec.ToRouteStatefulSession("internal-gateway-service")
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "internal-gateway-service", session.Gateways[0])
	assert.Equal(t, "sticky-cookie", session.CookieName)
	assert.Equal(t, "/", session.CookiePath)
	assert.NotNil(t, session.CookieTtl)
	assert.Equal(t, ttlVal, *session.CookieTtl)
	assert.True(t, session.Enabled)
}
