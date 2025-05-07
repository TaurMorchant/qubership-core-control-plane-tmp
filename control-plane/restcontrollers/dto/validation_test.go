package dto

import (
	"database/sql"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	asrt "github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRoutingConfigRequestV1_Validate(t *testing.T) {
	disableIPRouteRegistration()
	routingV1RequestValidator := RoutingV1RequestValidator{}

	request := RouteEntityRequest{
		MicroserviceUrl: "172.30.156.124:8080",
	}

	valid, err := routingV1RequestValidator.Validate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Invalid address at RouteEntityRequest: 172.30.156.124:8080 and nodeGroup: public-gateway-service", err)

	request = RouteEntityRequest{
		MicroserviceUrl: "172.30.156.124",
	}

	valid, err = routingV1RequestValidator.Validate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Invalid address at RouteEntityRequest: 172.30.156.124 and nodeGroup: public-gateway-service", err)

	request = RouteEntityRequest{
		MicroserviceUrl: "http://control-plane:8080",
		Routes: &[]RouteEntry{
			{
				From: "/",
			},
		},
	}
	valid, err = routingV1RequestValidator.Validate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route / is forbidden for registration", err)

	request = RouteEntityRequest{
		MicroserviceUrl: "http://control-plane:8080",
	}

	valid, err = routingV1RequestValidator.Validate(request, domain.PublicGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)

	request = RouteEntityRequest{
		MicroserviceUrl: "172.30.156.124:8080",
	}

	valid, err = routingV1RequestValidator.Validate(request, cache.EgressGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)

	enableIPRouteRegistration()
	request = RouteEntityRequest{
		MicroserviceUrl: "172.30.156.124:8080",
	}

	valid, err = routingV1RequestValidator.Validate(request, domain.PublicGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)

}

func TestRoutingConfigRequestV2_Validate(t *testing.T) {
	disableIPRouteRegistration()
	routingV2RequestValidator := RoutingV2RequestValidator{}

	requests := []RouteRegistrationRequest{
		{
			Endpoint: "172.30.156.124:8080",
		},
	}

	valid, err := routingV2RequestValidator.Validate(requests, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Registration of routes with ip address is forbidden for cluster:  at namespace:  in node group: public-gateway-service", err)

	requests = []RouteRegistrationRequest{
		{
			Endpoint: "172.30.156.124",
		},
	}

	valid, err = routingV2RequestValidator.Validate(requests, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Registration of routes with ip address is forbidden for cluster:  at namespace:  in node group: public-gateway-service", err)

	requests = []RouteRegistrationRequest{
		{
			Endpoint: "http://control-plane:8080",
			Routes: []RouteItem{
				{
					Prefix: "/",
				},
			},
		},
	}

	valid, err = routingV2RequestValidator.Validate(requests, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route / is forbidden for registration", err)

	requests = []RouteRegistrationRequest{
		{
			Endpoint: "http://control-plane:8080",
		},
	}

	valid, err = routingV2RequestValidator.Validate(requests, domain.PublicGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)

	requests = []RouteRegistrationRequest{
		{
			Endpoint: "172.30.156.124",
		},
	}

	valid, err = routingV2RequestValidator.Validate(requests, cache.EgressGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)

	enableIPRouteRegistration()
	requests = []RouteRegistrationRequest{
		{
			Endpoint: "172.30.156.124:8080",
		},
	}

	valid, err = routingV2RequestValidator.Validate(requests, domain.PublicGateway)
	asrt.True(t, valid)
	asrt.Equal(t, "", err)
}

func TestRoutingConfigRequestV3_Validate(t *testing.T) {
	disableIPRouteRegistration()
	routingV3RequestValidator := RoutingV3RequestValidator{}

	request := getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
		[]string{domain.PublicGateway},
	)
	valid, err := routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Incorrect virtual service with index 0. Msg: Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124",
		},
		[]string{domain.PublicGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Incorrect virtual service with index 0. Msg: Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	request = getRoutingConfigRequestV3ForRouteMatch(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "http://control-plane:8080",
		},
		[]string{domain.PublicGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.Equal(t, "Route / is forbidden for registration", err)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "http://control-plane",
		},
		[]string{domain.PublicGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
		[]string{cache.EgressGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
		[]string{cache.EgressGateway, domain.PublicGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Incorrect virtual service with index 0. Msg: Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	enableIPRouteRegistration()
	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
		[]string{domain.PublicGateway},
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{domain.PublicGateway},
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{"facade-gateway"},
		false,
		1234,
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{domain.PublicGateway},
		false,
		1234,
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Public, private and internal gateways do not allow custom listener ports", err)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{domain.PrivateGateway},
		false,
		1234,
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Public, private and internal gateways do not allow custom listener ports", err)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{domain.InternalGateway},
		false,
		1234,
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Public, private and internal gateways do not allow custom listener ports", err)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{"facade-gateway", domain.PublicGateway},
		false,
		1234,
	)
	request.VirtualServices[0].RateLimit = "test-ratelimit"
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
	asrt.Equal(t, "Public, private and internal gateways do not allow custom listener ports", err)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{"facade-gateway", domain.PublicGateway},
		false,
		0,
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{"facade-gateway"},
		true,
		0,
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRoutingConfigRequestV3WithListenerPortAndProto(
		RouteDestination{
			Cluster:     "control-plane",
			Endpoint:    "control-plane:8080",
			TlsEndpoint: "https://control-plane:8443",
		},
		[]string{"facade-gateway"},
		true,
		0,
	)
	valid, err = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)
}

func TestRoutingConfigRequestV3_ValidateVirtualService(t *testing.T) {
	disableIPRouteRegistration()
	enableTls()
	routingV3RequestValidator := RoutingV3RequestValidator{}

	request := getVirtualService(
		RouteDestination{
			Cluster:     "control-plane",
			Endpoint:    "control-plane:8080",
			TlsEndpoint: "https://control-plane:8443",
		},
	)
	valid, err := routingV3RequestValidator.ValidateVirtualService(request, []string{domain.PublicGateway})
	asrt.False(t, valid)
	asrt.Equal(t, "TlsEndpoint not supported in Route with index 0 in virtual service with name public-gateway-service", err)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualService(request, []string{domain.PublicGateway})
	asrt.True(t, valid)
}

func TestRoutingConfigRequestV3_ValidateVirtualServiceUpdate(t *testing.T) {
	disableIPRouteRegistration()
	routingV3RequestValidator := RoutingV3RequestValidator{}

	request := getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
	)
	valid, err := routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "http://control-plane:8080",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "http://control-plane",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, cache.EgressGateway)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route with index 0 in virtual service with name public-gateway-service has wrong endpoint field", err)

	enableTls()
	request = getVirtualService(
		RouteDestination{
			Cluster:     "control-plane",
			Endpoint:    "control-plane:8080",
			TlsEndpoint: "172.30.156.124:8443",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.False(t, valid)
	asrt.Equal(t, "Route with index 0 in virtual service with name public-gateway-service has wrong tls endpoint field", err)

	enableIPRouteRegistration()
	request = getVirtualService(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "172.30.156.124:8080",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.Equal(t, "", err)
	asrt.True(t, valid)

	request = getVirtualService(
		RouteDestination{
			Cluster:     "control-plane",
			Endpoint:    "control-plane:8080",
			TlsEndpoint: "172.30.156.124:8443",
		},
	)
	valid, err = routingV3RequestValidator.ValidateVirtualServiceUpdate(request, domain.PublicGateway)
	asrt.True(t, valid)
}

func getRoutingConfigRequestV3(destination RouteDestination, gateways []string) RoutingConfigRequestV3 {
	virtualServices := []VirtualService{getVirtualService(destination)}
	return RoutingConfigRequestV3{
		Namespace:       "namespace",
		Gateways:        gateways,
		VirtualServices: virtualServices,
	}
}

func getRoutingConfigRequestV3WithListenerPortAndProto(destination RouteDestination, gateways []string, tlsSupported bool, listenerPort int) RoutingConfigRequestV3 {
	virtualServices := []VirtualService{getVirtualService(destination)}
	return RoutingConfigRequestV3{
		Namespace:       "namespace",
		TlsSupported:    tlsSupported,
		ListenerPort:    listenerPort,
		Gateways:        gateways,
		VirtualServices: virtualServices,
	}
}

func getVirtualService(destination RouteDestination) VirtualService {
	return VirtualService{
		Name:          domain.PublicGateway,
		Hosts:         []string{"*"},
		AddHeaders:    []HeaderDefinition{},
		RemoveHeaders: []string{},
		RouteConfiguration: RouteConfig{
			Version: "v1",
			Routes: []RouteV3{
				{
					Destination: destination,
					Rules: []Rule{
						{
							Match: RouteMatch{
								Prefix: "prefix",
							},
							PrefixRewrite: "prefixRewrite",
						},
					},
				},
			},
		},
	}
}

func getRoutingConfigRequestV3ForRouteMatch(destination RouteDestination, gateways []string) RoutingConfigRequestV3 {
	virtualServices := []VirtualService{getVirtualServiceForRouteMatch(destination)}
	return RoutingConfigRequestV3{
		Namespace:       "namespace",
		Gateways:        gateways,
		VirtualServices: virtualServices,
	}
}

func getVirtualServiceForRouteMatch(destination RouteDestination) VirtualService {
	return VirtualService{
		Name:          domain.PublicGateway,
		Hosts:         []string{"*"},
		AddHeaders:    []HeaderDefinition{},
		RemoveHeaders: []string{},
		RouteConfiguration: RouteConfig{
			Version: "v1",
			Routes: []RouteV3{
				{
					Destination: destination,
					Rules: []Rule{
						{
							Match: RouteMatch{
								Prefix: "/",
							},
							PrefixRewrite: "prefixRewrite",
						},
					},
				},
			},
		},
	}
}

func TestIsValidDeploymentVersion(t *testing.T) {
	assert := asrt.New(t)

	assert.True(isValidDeploymentVersion("v1"))
	assert.True(isValidDeploymentVersion("v2"))
	assert.True(isValidDeploymentVersion("v3"))
	assert.True(isValidDeploymentVersion("v77"))
	assert.True(isValidDeploymentVersion("v12312"))

	assert.False(isValidDeploymentVersion("-v1"))
	assert.False(isValidDeploymentVersion("+v1"))
	assert.False(isValidDeploymentVersion("v-122"))
	assert.False(isValidDeploymentVersion("v12+5"))
	assert.False(isValidDeploymentVersion("v12-55"))
	assert.False(isValidDeploymentVersion("v12.54"))
	assert.False(isValidDeploymentVersion("v12,53"))
	assert.False(isValidDeploymentVersion("v12_52"))
	assert.False(isValidDeploymentVersion(" v1"))
	assert.False(isValidDeploymentVersion("	v1"))
	assert.False(isValidDeploymentVersion(" v1 "))
	assert.False(isValidDeploymentVersion("v2    "))
	assert.False(isValidDeploymentVersion("v3	"))
	assert.False(isValidDeploymentVersion("v1 5"))
	assert.False(isValidDeploymentVersion("v1 v5"))
	assert.False(isValidDeploymentVersion("a1"))
	assert.False(isValidDeploymentVersion("b2"))
	assert.False(isValidDeploymentVersion("zz2"))
	assert.False(isValidDeploymentVersion("1"))
	assert.False(isValidDeploymentVersion("2"))
	assert.False(isValidDeploymentVersion("22"))
	assert.False(isValidDeploymentVersion("-33"))

	assert.False(isValidDeploymentVersion("{{DEPLOYMENT_VERSION}}"))
	assert.False(isValidDeploymentVersion("{{ENV_DEPLOYMENT_VERSION}}"))
	assert.False(isValidDeploymentVersion("${ENV_DEPLOYMENT_VERSION}"))
	assert.False(isValidDeploymentVersion("${DEPLOYMENT_VERSION}"))
	assert.False(isValidDeploymentVersion("${SOME_OTHER_VAR}"))
}

func TestIsValidAddress(t *testing.T) {
	assert := asrt.New(t)
	disableIPRouteRegistration()

	assert.True(isValidAddress("http://test.t"))
	assert.True(isValidAddress("https://test.t"))
	assert.True(isValidAddress("test.t:8080"))
	assert.True(isValidAddress("http://test.t:8080"))
	assert.True(isValidAddress("http://test.t:8080/cde?foo=bar"))
	assert.True(isValidAddress("test.t:8080/cde?foo=bar"))
	assert.True(isValidAddress("test.t:65535/cde?foo=bar"))

	assert.False(isValidAddress("http://test.t:-1"))
	assert.False(isValidAddress("https://test.t:1q00"))
	assert.False(isValidAddress("https://test.t:65536"))
}

func testValidateStatefulSessionRequest(assert *asrt.Assertions, req StatefulSession, positive bool) {
	routingV3RequestValidator := RoutingV3RequestValidator{}
	trueVal := true
	falseVal := false

	valid, errMsg := routingV3RequestValidator.ValidateStatefulSession(req)
	if positive {
		assert.True(valid)
		assert.Empty(errMsg)
	} else {
		assert.False(valid)
		assert.NotEmpty(errMsg)
	}

	if positive {
		// test disable
		req.Enabled = &falseVal
		valid, errMsg := routingV3RequestValidator.ValidateStatefulSession(req)
		assert.True(valid)
		assert.Empty(errMsg)

		req.Cookie = &Cookie{
			Name: "cookie",
			Ttl:  nil,
			Path: domain.NullString{NullString: sql.NullString{String: "/", Valid: true}},
		}
		valid, errMsg = routingV3RequestValidator.ValidateStatefulSession(req)
		assert.False(valid)
		assert.NotEmpty(errMsg)

		// test apply
		req.Enabled = nil
		valid, errMsg = routingV3RequestValidator.ValidateStatefulSession(req)
		assert.True(valid)
		assert.Empty(errMsg)

		req.Enabled = &trueVal
		valid, errMsg = routingV3RequestValidator.ValidateStatefulSession(req)
		assert.True(valid)
		assert.Empty(errMsg)

		// test delete
		req.Cookie = nil
		valid, errMsg = routingV3RequestValidator.ValidateStatefulSession(req)
		assert.False(valid)
		assert.NotEmpty(errMsg)

		req.Enabled = nil
		valid, errMsg = routingV3RequestValidator.ValidateStatefulSession(req)
		assert.True(valid)
		assert.Empty(errMsg)
	}

}

func TestValidateStatefulSessionPerCluster(t *testing.T) {
	assert := asrt.New(t)
	port := 8080

	request := StatefulSession{
		Namespace: "namespace",
		Cluster:   "",
		Version:   "",
		Port:      nil,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "",
		Version:   "",
		Port:      nil,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "cluster",
		Version:   "",
		Port:      nil,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "",
		Cluster:   "cluster",
		Version:   "",
		Port:      nil,
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "",
		Version:   "v1",
		Port:      nil,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "",
		Version:   "",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "some-cluster",
		Version:   "v1",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "",
		Cluster:   "some-cluster",
		Version:   "v1",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "my-namespace",
		Cluster:   "some-cluster",
		Version:   "v1",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "my-namespace",
		Cluster:   "some-cluster",
		Version:   "v1",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "some-cluster",
		Version:   "v1",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "",
		Cluster:   "some-cluster",
		Version:   "v1",
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "my-namespace",
		Cluster:   "some-cluster",
		Version:   "v1",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "my-namespace",
		Cluster:   "some-cluster",
		Version:   "v1",
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "",
		Cluster:   "some-cluster",
		Hostname:  "test-service",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "namespace",
		Cluster:   "some-cluster",
		Hostname:  "test-service",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "namespace",
		Cluster:   "some-cluster",
		Hostname:  "test-service",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Gateways:  []string{"private-gateway-service"},
		Namespace: "namespace",
		Cluster:   "some-cluster",
		Hostname:  "test-service",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, true)

	request = StatefulSession{
		Namespace: "",
		Hostname:  "test-service",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "namespace",
		Hostname:  "test-service",
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "",
		Hostname:  "test-service",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)

	request = StatefulSession{
		Namespace: "namespace",
		Hostname:  "test-service",
		Port:      &port,
	}
	testValidateStatefulSessionRequest(assert, request, false)
}

func TestValidateStatefulSessionForRoute(t *testing.T) {
	routingV3RequestValidator := RoutingV3RequestValidator{}
	falseVal := false
	trueVal := true
	port := 8080

	request := getRouteConfigWithStatefulSession(nil)
	valid, _ := routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Enabled: &falseVal})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Cluster: "service"})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Gateways: []string{"egress-gateway"}})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Port: &port})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Version: "v1"})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Cookie: &Cookie{Name: "cookie"}})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Cookie: &Cookie{Name: "cookie"}, Enabled: &trueVal})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.True(t, valid)

	request = getRouteConfigWithStatefulSession(&StatefulSession{Cookie: &Cookie{Name: "cookie"}, Enabled: &falseVal})
	valid, _ = routingV3RequestValidator.Validate(request)
	asrt.False(t, valid)
}

func getRouteConfigWithStatefulSession(session *StatefulSession) RoutingConfigRequestV3 {
	request := getRoutingConfigRequestV3(
		RouteDestination{
			Cluster:  "control-plane",
			Endpoint: "control-plane:8080",
		},
		[]string{domain.PublicGateway},
	)
	request.VirtualServices[0].RouteConfiguration.Routes[0].Rules[0].StatefulSession = session
	return request
}

func disableIPRouteRegistration() {
	os.Setenv("DISABLE_IP_ROUTE_REGISTRATION", "true")
	defer os.Unsetenv("DISABLE_IP_ROUTE_REGISTRATION")
	reloadIpRouteRegistrationValidation()
}

func enableIPRouteRegistration() {
	os.Setenv("DISABLE_IP_ROUTE_REGISTRATION", "false")
	defer os.Unsetenv("DISABLE_IP_ROUTE_REGISTRATION")
	reloadIpRouteRegistrationValidation()
}

func disableTls() {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "false")
	defer os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()
}

func enableTls() {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	defer os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()
}
