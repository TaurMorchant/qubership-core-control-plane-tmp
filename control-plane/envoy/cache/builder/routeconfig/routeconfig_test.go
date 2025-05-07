package routeconfig

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	eroute "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/builder/common"
	mock_dao "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/dao"
	mock_routeconfig "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder/routeconfig"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_sortHashPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockRoutePreparer := getMockRoutePreparer(ctrl)
	envoyProxyProperties := getEnvoyProxyProperties()

	routeBuilder := NewRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)
	hashPolicy := []*domain.HashPolicy{
		{Id: int32(1), CookieName: "CookieName"},
		{Id: int32(2), CookieName: "CookieName"},
		{Id: int32(3), CookieName: "CookieName", Terminal: domain.NewNullBool(true)},
		{Id: int32(4), CookieName: "CookieName"},
	}

	result := routeBuilder.sortHashPolicies(hashPolicy)
	assert.Equal(t, int32(3), result[0].Id)
}

func Test_constructors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockRoutePreparer := getMockRoutePreparer(ctrl)
	envoyProxyProperties := getEnvoyProxyProperties()

	facadeRouteBuilder := NewMeshRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)
	assert.Equal(t, true, facadeRouteBuilder.facade)
	assert.Equal(t, false, facadeRouteBuilder.hostRewrite)

	egressRouteBuilder := NewEgressRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)
	assert.Equal(t, true, egressRouteBuilder.facade)
	assert.Equal(t, true, egressRouteBuilder.hostRewrite)
}

func TestBuildRouteHashPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockRoutePreparer := getMockRoutePreparer(ctrl)
	envoyProxyProperties := getEnvoyProxyProperties()
	routeBuilder := NewRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)

	hashPolicy := &domain.HashPolicy{
		Id: int32(1), HeaderName: "HeaderName",
	}
	resultHashPolicy, err := routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, err)
	assert.Equal(t, hashPolicy.HeaderName, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Header_).Header.HeaderName)

	hashPolicy = &domain.HashPolicy{
		Id: int32(1), CookieName: "HeaderName", CookiePath: "CookiePath", CookieTTL: domain.NullInt{},
	}
	resultHashPolicy, err = routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, err)
	assert.Equal(t, hashPolicy.CookieName, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Name)
	assert.Equal(t, hashPolicy.CookiePath, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Path)
	assert.Nil(t, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Ttl)

	cookieTTLValue := domain.NewNullInt(10)

	hashPolicy = &domain.HashPolicy{
		Id: int32(1), CookieName: "HeaderName", CookiePath: "CookiePath", CookieTTL: cookieTTLValue,
	}
	resultHashPolicy, err = routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, err)
	assert.Equal(t, hashPolicy.CookieName, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Name)
	assert.Equal(t, hashPolicy.CookiePath, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Path)
	assert.Equal(t, cookieTTLValue.Int64, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Ttl.Seconds)

	hashPolicy = &domain.HashPolicy{
		Id: int32(1), CookieName: "", QueryParamName: "QueryParamName",
	}
	resultHashPolicy, err = routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, err)
	assert.Equal(t, hashPolicy.QueryParamName, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_QueryParameter_).QueryParameter.Name)

	hashPolicy = &domain.HashPolicy{
		Id: int32(1), CookieName: "", QueryParamSourceIP: domain.NewNullBool(true),
	}
	resultHashPolicy, err = routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, err)
	assert.Equal(t, hashPolicy.QueryParamSourceIP.Bool, resultHashPolicy.PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_ConnectionProperties_).ConnectionProperties.SourceIp)

	hashPolicy = &domain.HashPolicy{
		Id: int32(1),
	}
	resultHashPolicy, err = routeBuilder.buildRouteHashPolicy(hashPolicy)
	assert.Nil(t, resultHashPolicy)
	assert.NotNil(t, err)
}

func TestBuildHeaderMatchers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockRoutePreparer := getMockRoutePreparer(ctrl)
	envoyProxyProperties := getEnvoyProxyProperties()
	routeBuilder := NewRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)

	headerMatchers := []*domain.HeaderMatcher{
		{},
		{
			SuffixMatch: "SuffixMatch",
		},
		{
			SafeRegexMatch: "SafeRegexMatch",
		},
		{
			RangeMatch: domain.RangeMatch{
				Start: domain.NewNullInt(1),
				End:   domain.NewNullInt(10),
			},
		},
		{
			PresentMatch: domain.NewNullBool(true),
		},
		{
			PrefixMatch: "PrefixMatch",
		},
		{
			ExactMatch: "ExactMatch",
		},
	}
	resultHeaderMatchers, err := routeBuilder.buildHeaderMatchers(headerMatchers)
	assert.Nil(t, err)
	assert.Equal(t, len(headerMatchers), len(resultHeaderMatchers))

	assert.True(t, resultHeaderMatchers[0].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_PresentMatch).PresentMatch)
	assert.Equal(t, headerMatchers[1].SuffixMatch, resultHeaderMatchers[1].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_SuffixMatch).SuffixMatch)
	assert.Equal(t, headerMatchers[2].SafeRegexMatch, resultHeaderMatchers[2].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_SafeRegexMatch).SafeRegexMatch.Regex)
	assert.Equal(t, headerMatchers[3].RangeMatch.Start.Int64, resultHeaderMatchers[3].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_RangeMatch).RangeMatch.Start)
	assert.Equal(t, headerMatchers[3].RangeMatch.End.Int64, resultHeaderMatchers[3].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_RangeMatch).RangeMatch.End)
	assert.Equal(t, headerMatchers[4].PresentMatch.Bool, resultHeaderMatchers[4].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_PresentMatch).PresentMatch)
	assert.Equal(t, headerMatchers[5].PrefixMatch, resultHeaderMatchers[5].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_PrefixMatch).PrefixMatch)
	assert.Equal(t, headerMatchers[6].ExactMatch, resultHeaderMatchers[6].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_StringMatch).StringMatch.GetExact())
}

func TestRouteBuilderBuildRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockRoutePreparer := getMockRoutePreparer(ctrl)
	envoyProxyProperties := getEnvoyProxyProperties()
	routeBuilder := NewRouteBuilder(mockDao, envoyProxyProperties, mockRoutePreparer)

	routes, headerMatcher, hashPolicy := initTestData(mockDao, mockRoutePreparer)

	resultRoutes, rateLimits, err := routeBuilder.BuildRoutes(routes)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(resultRoutes))

	assert.Equal(t, 1, len(rateLimits))
	rateLimit := rateLimits["testRateLimit"]
	assert.Equal(t, "testRateLimit", rateLimit.Name)
	assert.Equal(t, uint32(10), rateLimit.LimitRequestsPerSecond)
	assert.Equal(t, domain.Project, rateLimit.Priority)

	assert.Equal(t, headerMatcher[0].Name, resultRoutes[0].Match.Headers[0].Name)

	assert.Equal(t, routes[0].Regexp, resultRoutes[0].Match.PathSpecifier.(*envoy_config_route_v3.RouteMatch_SafeRegex).SafeRegex.Regex)
	assert.Equal(t, true, resultRoutes[0].Match.Headers[0].HeaderMatchSpecifier.(*envoy_config_route_v3.HeaderMatcher_PresentMatch).PresentMatch)

	envoyConfigRoute := resultRoutes[0].Action.(*envoy_config_route_v3.Route_Route).Route
	assert.Equal(t, routes[0].PrefixRewrite, envoyConfigRoute.PrefixRewrite)
	assert.Equal(t, routes[0].Regexp, envoyConfigRoute.RegexRewrite.Pattern.Regex)
	assert.Equal(t, routes[0].RegexpRewrite, envoyConfigRoute.RegexRewrite.Substitution)
	assert.Equal(t, routes[0].HostRewrite, envoyConfigRoute.HostRewriteSpecifier.(*envoy_config_route_v3.RouteAction_HostRewriteLiteral).HostRewriteLiteral)

	assert.Equal(t, int32(envoyProxyProperties.Routes.Timeout*1000000), envoyConfigRoute.Timeout.Nanos)
	assert.Equal(t, int32(routes[0].IdleTimeout.Int64*1000000), envoyConfigRoute.IdleTimeout.Nanos)

	assert.Equal(t, hashPolicy[0].CookieName, envoyConfigRoute.HashPolicy[0].PolicySpecifier.(*envoy_config_route_v3.RouteAction_HashPolicy_Cookie_).Cookie.Name)

	assert.Equal(t, routes[0].RequestHeadersToAdd[0].Name, resultRoutes[0].RequestHeadersToAdd[0].Header.Key)
	assert.Equal(t, routes[0].RequestHeadersToAdd[0].Value, resultRoutes[0].RequestHeadersToAdd[0].Header.Value)
	assert.Equal(t, routes[0].RequestHeadersToRemove, resultRoutes[0].RequestHeadersToRemove)
	assert.Equal(t, int32(routes[0].RetryPolicy.RetryBackOff.BaseInterval*1000000), resultRoutes[0].Action.(*eroute.Route_Route).Route.RetryPolicy.RetryBackOff.BaseInterval.Nanos)
	assert.Equal(t, int32(routes[0].RetryPolicy.RetryBackOff.MaxInterval*1000000), resultRoutes[0].Action.(*eroute.Route_Route).Route.RetryPolicy.RetryBackOff.MaxInterval.Nanos)

	assert.Equal(t, routes[1].DirectResponseCode, resultRoutes[1].Action.(*eroute.Route_DirectResponse).DirectResponse.Status)

	assert.Equal(t, 1, len(envoyConfigRoute.RateLimits))
	assert.Equal(t, 1, len(envoyConfigRoute.RateLimits[0].Actions))
	action := envoyConfigRoute.RateLimits[0].Actions[0]
	assert.Equal(t, "rate_limit_name", action.GetGenericKey().DescriptorKey)
	assert.Equal(t, rateLimit.Name, action.GetGenericKey().DescriptorValue)
}

func initTestData(mockDao *mock_dao.MockDao, mockRoutePreparer *mock_routeconfig.MockRoutePreparer) ([]*domain.Route, []*domain.HeaderMatcher, []*domain.HashPolicy) {
	rateLimit := &domain.RateLimit{
		Name:                   "testRateLimit",
		LimitRequestsPerSecond: 10,
		Priority:               domain.Project,
	}
	routes := []*domain.Route{
		{
			Id:                     int32(1),
			Uuid:                   "d64a9674-96ae-4a1a-b168-9a55afe6d6c8",
			DeploymentVersion:      "dv1",
			Prefix:                 "pref",
			Path:                   "path",
			Regexp:                 "Regexp",
			RequestHeadersToAdd:    []domain.Header{{Name: "name1", Value: "value1"}},
			RequestHeadersToRemove: []string{"remove"},
			PrefixRewrite:          "PrefixRewrite",
			RegexpRewrite:          "RegexpRewrite",
			HostRewrite:            "HostRewrite",
			HostAutoRewrite:        domain.NewNullBool(true),
			RateLimitId:            "testRateLimit",
			RateLimit:              rateLimit,
			RetryPolicy:            &domain.RetryPolicy{RetryBackOff: &domain.RetryBackOff{1, 2}},
			IdleTimeout:            domain.NewNullInt(10),
		},
		{
			Id:                 int32(2),
			DeploymentVersion:  "dv2",
			DirectResponseCode: uint32(1),
			Prefix:             "pref",
		},
	}
	deploymentVersion := &domain.DeploymentVersion{
		Version: "v1", Stage: domain.ActiveStage,
	}
	mockDao.EXPECT().FindDeploymentVersion(routes[0].DeploymentVersion).Return(deploymentVersion, nil)

	deploymentVersion2 := &domain.DeploymentVersion{
		Version: "v2", Stage: domain.ActiveStage,
	}
	mockDao.EXPECT().FindDeploymentVersion(routes[1].DeploymentVersion).Return(deploymentVersion2, nil)

	headerMatcher := []*domain.HeaderMatcher{
		{Id: int32(1), Name: "headerName"},
	}
	mockDao.EXPECT().FindHeaderMatcherByRouteId(routes[0].Id).Return(headerMatcher, nil)

	mockDao.EXPECT().FindHeaderMatcherByRouteId(routes[1].Id).Return([]*domain.HeaderMatcher{}, nil)

	hashPolicy := []*domain.HashPolicy{
		{Id: int32(1), CookieName: "CookieName"},
	}
	mockDao.EXPECT().FindHashPolicyByRouteId(routes[0].Id).Return(hashPolicy, nil)

	mockDao.EXPECT().FindHashPolicyByRouteId(routes[1].Id).Return([]*domain.HashPolicy{}, nil)

	//mockDao.EXPECT().FindStatefulSessionConfigById(routes[0].StatefulSessionId).Return(&domain.StatefulSession{Id: int32(1), CookieName: "sticky-cookie"}, nil)

	retryPolicy := &domain.RetryPolicy{
		Id:           int32(1),
		RetryBackOff: &domain.RetryBackOff{1, 2},
	}
	mockDao.EXPECT().FindRetryPolicyByRouteId(routes[0].Id).Return(retryPolicy, nil)

	mockDao.EXPECT().FindRetryPolicyByRouteId(routes[1].Id).Return(&domain.RetryPolicy{Id: int32(2)}, nil)

	mockDao.EXPECT().FindRateLimitByNameWithHighestPriority(routes[0].RateLimitId).Return(rateLimit, nil)

	mockRoutePreparer.EXPECT().Prepare(gomock.Any()).Return(routes)

	return routes, headerMatcher, hashPolicy
}

func getEnvoyProxyProperties() *common.EnvoyProxyProperties {
	return &common.EnvoyProxyProperties{
		Routes: &common.RouteProperties{
			Timeout: int64(10),
		},
		Compression: &common.CompressionProperties{
			Enabled:         false,
			MimeTypes:       "string",
			MinResponseSize: 0,
			MimeTypesList:   []string{"MimeTypesList"},
		},
		Tracing: &common.TracingProperties{
			Enabled:                 false,
			ZipkinCollectorCluster:  "string",
			ZipkinCollectorEndpoint: "string",
		},
		Googlere2: &common.GoogleRe2Properties{
			Maxsize:  "string",
			WarnSize: "string",
		},
		Connection: &common.Connection{PerConnectionBufferLimitMegabytes: 0},
	}
}
