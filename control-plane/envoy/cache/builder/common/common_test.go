package common

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	stateful_sessionv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/stateful_session/v3"
	cookiev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/http/stateful_session/cookie/v3"
	"github.com/golang/protobuf/ptypes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildCookieBasedSessionFilterForListener(t *testing.T) {
	anyFilter, err := BuildCookieBasedSessionFilterForListener()
	assert.Nil(t, err)
	assert.NotNil(t, anyFilter)
	session := &stateful_sessionv3.StatefulSession{}
	assert.Nil(t, ptypes.UnmarshalAny(anyFilter, session))
	assert.NotNil(t, session)
	assert.Nil(t, session.SessionState)
}

func TestBuildStatefulSessionPerRoute(t *testing.T) {
	ttl := int64(1000)
	anyFilter, err := BuildStatefulSessionPerRoute(&domain.StatefulSession{
		Id:                       1,
		CookieName:               "sticky-cookie-v1",
		CookieTtl:                &ttl,
		CookiePath:               "/",
		Enabled:                  true,
		ClusterName:              "test-cluster",
		Namespace:                "default",
		Gateways:                 []string{"private-gateway-service"},
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, anyFilter)
	session := &stateful_sessionv3.StatefulSessionPerRoute{}
	assert.Nil(t, ptypes.UnmarshalAny(anyFilter, session))
	assert.NotNil(t, session)
	assert.False(t, session.GetDisabled())
	statefulSession := session.GetStatefulSession()
	assert.NotNil(t, statefulSession)
	state := statefulSession.GetSessionState()
	assert.NotNil(t, state)
	assert.Equal(t, "envoy.http.stateful_session.cookie", state.GetName())
	config := state.GetTypedConfig()
	assert.NotNil(t, config)
	cookieBasedFilter := &cookiev3.CookieBasedSessionState{}
	assert.Nil(t, ptypes.UnmarshalAny(config, cookieBasedFilter))
	assert.NotNil(t, cookieBasedFilter)
	cookie := cookieBasedFilter.GetCookie()
	assert.NotNil(t, cookie)
	assert.Equal(t, "sticky-cookie-v1", cookie.Name)
	assert.Equal(t, "/", cookie.Path)
	assert.NotNil(t, cookie.Ttl)
	assert.Equal(t, util.MillisToDuration(ttl).GetSeconds(), cookie.Ttl.GetSeconds())
	assert.Equal(t, util.MillisToDuration(ttl).GetNanos(), cookie.Ttl.GetNanos())
}

func TestBuildSocketAddr(t *testing.T) {
	verifyBuiltAddress(t, false)
}

func TestBuildSocketAddrIPv4(t *testing.T) {
	t.Setenv("IP_STACK", "v4")
	verifyBuiltAddress(t, false)
}

func TestBuildSocketAddrIPv6(t *testing.T) {
	t.Setenv("IP_STACK", "v6")
	verifyBuiltAddress(t, true)
}

func verifyBuiltAddress(t *testing.T, expectIpv4Compat bool) {
	resolveIPVersion()

	address := "address"
	port := uint32(8080)

	expAddress := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Protocol: core.SocketAddress_TCP,
			Address:  address,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: port,
			},
			Ipv4Compat: expectIpv4Compat,
		},
	}}

	resultAddress := BuildSocketAddr(address, port)
	assert.Equal(t, expAddress, resultAddress)
}

func TestBuildFilenameDataSource(t *testing.T) {
	fileName := "test_file_name"
	dataSource := BuildFilenameDataSource(fileName)
	assert.NotNil(t, dataSource.Specifier)
	assert.Equal(t, dataSource.GetFilename(), fileName)
}

func TestBuildInlineStringDataSource(t *testing.T) {
	inlineString := "test_inline_string"
	dataSource := BuildInlineStringDataSource(inlineString)
	assert.NotNil(t, dataSource.Specifier)
	assert.Equal(t, dataSource.GetInlineString(), inlineString)
}
