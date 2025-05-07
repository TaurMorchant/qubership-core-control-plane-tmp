package routeconfig

import (
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestRouteStatefulSessionPreparer_Prepare_PerCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	routePreparer := NewRouteStatefulSessionPreparer(mockDao, entity.NewService("v1"))

	version := &domain.DeploymentVersion{Version: "v1"}
	routes := []*domain.Route{
		{ClusterName: "test-cluster||test-cluster||8080", DeploymentVersion: "v1", DeploymentVersionVal: version, VirtualHostId: int32(1)},
	}
	clusterName := clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(routes[0].ClusterName)
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(routes[0].ClusterName)

	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(&domain.RouteConfiguration{NodeGroupId: "private-gateway-service"}, nil)
	mockDao.EXPECT().FindStatefulSessionConfigsByClusterAndVersion(clusterName, namespace, version).
		Return([]*domain.StatefulSession{{
			Id:         1,
			Gateways:   []string{"private-gateway-service"},
			Enabled:    true,
			CookieName: "sticky-cookie-v1",
			CookiePath: "/",
		}}, nil)
	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(nil, nil)

	resultRoutes := routePreparer.Prepare(routes)
	assert.NotNil(t, resultRoutes)
	assert.Equal(t, routes[0].ClusterName, resultRoutes[0].ClusterName)
	assert.Equal(t, routes[0].DeploymentVersion, resultRoutes[0].DeploymentVersion)
	assert.Equal(t, routes[0].VirtualHostId, resultRoutes[0].VirtualHostId)
	session := resultRoutes[0].StatefulSession
	assert.NotNil(t, session)
	assert.Equal(t, int32(1), session.Id)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "/", session.CookiePath)
	assert.NotNil(t, session.Gateways)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestRouteStatefulSessionPreparer_Prepare_PerEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	routePreparer := NewRouteStatefulSessionPreparer(mockDao, entity.NewService("v1"))

	version := &domain.DeploymentVersion{Version: "v1"}
	routes := []*domain.Route{
		{ClusterName: "test-cluster||test-cluster||8080", DeploymentVersion: "v1", DeploymentVersionVal: version, VirtualHostId: int32(1)},
	}
	clusterName := clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(routes[0].ClusterName)
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(routes[0].ClusterName)

	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(&domain.RouteConfiguration{NodeGroupId: "private-gateway-service"}, nil)
	mockDao.EXPECT().FindStatefulSessionConfigsByClusterAndVersion(clusterName, namespace, version).
		Return([]*domain.StatefulSession{{
			Id:         1,
			Gateways:   []string{"private-gateway-service"},
			Enabled:    true,
			CookieName: "sticky-cookie-v1",
			CookiePath: "/",
		}}, nil)
	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(&domain.Endpoint{}, nil)

	resultRoutes := routePreparer.Prepare(routes)
	assert.NotNil(t, resultRoutes)
	assert.Equal(t, routes[0].ClusterName, resultRoutes[0].ClusterName)
	assert.Equal(t, routes[0].DeploymentVersion, resultRoutes[0].DeploymentVersion)
	assert.Equal(t, routes[0].VirtualHostId, resultRoutes[0].VirtualHostId)
	session := resultRoutes[0].StatefulSession
	assert.NotNil(t, session)
	assert.Equal(t, int32(1), session.Id)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "/", session.CookiePath)
	assert.NotNil(t, session.Gateways)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestRouteStatefulSessionPreparer_Prepare_PerRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	routePreparer := NewRouteStatefulSessionPreparer(mockDao, entity.NewService("v1"))

	version := &domain.DeploymentVersion{Version: "v1"}
	routes := []*domain.Route{
		{
			ClusterName:          "test-cluster||test-cluster||8080",
			DeploymentVersion:    "v1",
			DeploymentVersionVal: version,
			VirtualHostId:        int32(1),
			StatefulSessionId:    int32(1),
			StatefulSession: &domain.StatefulSession{
				Id:         1,
				Gateways:   []string{"private-gateway-service"},
				Enabled:    true,
				CookieName: "sticky-cookie-v1",
				CookiePath: "/",
			},
		},
	}

	resultRoutes := routePreparer.Prepare(routes)
	assert.NotNil(t, resultRoutes)
	assert.Equal(t, routes[0].ClusterName, resultRoutes[0].ClusterName)
	assert.Equal(t, routes[0].DeploymentVersion, resultRoutes[0].DeploymentVersion)
	assert.Equal(t, routes[0].VirtualHostId, resultRoutes[0].VirtualHostId)
	session := resultRoutes[0].StatefulSession
	assert.NotNil(t, session)
	assert.Equal(t, int32(1), session.Id)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "/", session.CookiePath)
	assert.NotNil(t, session.Gateways)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestRouteMultiVersionPreparer_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	routeTransformer := NewRouteTransformer(
		NewSimpleRouteTransformationRule(),
		NewGenericVersionedRouteTransformRule(),
		NewNoActiveRouteTransformer())
	routeMultiVersionPreparer := NewRouteMultiVersionPreparer(routePreparerMock{}, mockDao, routeTransformer)

	deploymentVersions := []*domain.DeploymentVersion{
		{
			Version: "v1",
			Stage:   domain.ActiveStage,
		},
	}
	mockDao.EXPECT().FindAllDeploymentVersions().Return(deploymentVersions, nil)

	hashPolicies := []*domain.HashPolicy{
		{Id: int32(1), HeaderName: "HeaderName"},
		{Id: int32(1), HeaderName: "HeaderName"},
	}
	mockDao.EXPECT().FindHashPolicyByClusterAndVersions(gomock.Any(), gomock.Any()).Return(hashPolicies, nil)
	mockDao.EXPECT().FindHashPolicyByRouteId(gomock.Any()).Return(hashPolicies, nil)

	routes := []*domain.Route{
		{ClusterName: "ClusterName1", DeploymentVersionVal: deploymentVersions[0]},
	}
	resultRoutes := routeMultiVersionPreparer.Prepare(routes)
	assert.NotNil(t, resultRoutes)
	assert.Equal(t, routes[0].ClusterName, resultRoutes[0].ClusterName)
	assert.Equal(t, 1, len(resultRoutes[0].HashPolicies))
	assert.Equal(t, hashPolicies[0].Id, resultRoutes[0].HashPolicies[0].Id)
	assert.Equal(t, hashPolicies[0].HeaderName, resultRoutes[0].HashPolicies[0].HeaderName)
}

func TestSortRoutePreparer_Prepare(t *testing.T) {
	type args struct {
		routes []*domain.Route
	}
	tests := []struct {
		name string
		args args
		want []*domain.Route
	}{
		{
			name: "Sort prefix routes",
			args: struct{ routes []*domain.Route }{routes: []*domain.Route{
				{Prefix: "/short"},
				{Prefix: "/api/v2/another-service/long/narrow/endpoint"},
				{Prefix: "/api/v1/some-service/endpoint"},
			}},
			want: []*domain.Route{
				{Prefix: "/api/v2/another-service/long/narrow/endpoint"},
				{Prefix: "/api/v1/some-service/endpoint"},
				{Prefix: "/short"},
			},
		},
		{
			name: "Sort regexp routes",
			args: struct{ routes []*domain.Route }{routes: []*domain.Route{
				{Regexp: "/short(/.*)?"},
				{Regexp: "/api/v2/another-service/long/([^/]+)/endpoint(/.*)?"},
				{Regexp: "/api/v1/some-service/endpoint/([^/]+)(/.*)?"},
			}},
			want: []*domain.Route{
				{Regexp: "/api/v2/another-service/long/([^/]+)/endpoint(/.*)?"},
				{Regexp: "/api/v1/some-service/endpoint/([^/]+)(/.*)?"},
				{Regexp: "/short(/.*)?"},
			},
		},
		{
			name: "Sort mixed routes",
			args: struct{ routes []*domain.Route }{routes: []*domain.Route{
				{Regexp: "/short(/.*)?"},
				{Prefix: "/api/v2/another-service/endpoint/operation/data"},
				{Regexp: "/api/v2/another-service/endpoint/([^/]+)/details(/.*)?"},
			}},
			want: []*domain.Route{
				{Prefix: "/api/v2/another-service/endpoint/operation/data"},
				{Regexp: "/api/v2/another-service/endpoint/([^/]+)/details(/.*)?"},
				{Regexp: "/short(/.*)?"},
			},
		},
		{
			name: "Sort blue-green routes",
			args: struct{ routes []*domain.Route }{routes: []*domain.Route{
				{Regexp: "/short(/.*)?"},
				{Prefix: "/api/v2/another-service/endpoint/operation/data"},
				{Regexp: "/api/v2/another-service/endpoint/([^/]+)/details(/.*)?"},
				{Prefix: "/short", HeaderMatchers: []*domain.HeaderMatcher{{Name: "x-version"}}},
			}},
			want: []*domain.Route{
				{Prefix: "/short", HeaderMatchers: []*domain.HeaderMatcher{{Name: "x-version"}}},
				{Prefix: "/api/v2/another-service/endpoint/operation/data"},
				{Regexp: "/api/v2/another-service/endpoint/([^/]+)/details(/.*)?"},
				{Regexp: "/short(/.*)?"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preparer := &SortRoutePreparer{
				beforePreparer: routePreparerMock{},
			}
			if got := preparer.Prepare(tt.args.routes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Prepare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortRoutePreparer_WithExtraSlashPreparer(t *testing.T) {
	extraSlashPreparer := NewExtraSlashRoutePreparer(routePreparerMock{})
	sortPreparer := NewSortRoutePreparer(extraSlashPreparer)

	routes := []*domain.Route{
		{
			Prefix:        "/api/v1/tenants/6b741a74-5d2e-4dfe-90ea-e44f38f2bf10/shopping-frontend/",
			PrefixRewrite: "/",
		},
	}
	expected := []*domain.Route{
		{
			Prefix:        "/api/v1/tenants/6b741a74-5d2e-4dfe-90ea-e44f38f2bf10/shopping-frontend/",
			PrefixRewrite: "/",
		},
		{
			Prefix: "/api/v1/tenants/6b741a74-5d2e-4dfe-90ea-e44f38f2bf10/shopping-frontend",
		},
	}

	for i := 0; i < 5; i++ {
		sortedWithClonedRoute := sortPreparer.Prepare(routes)
		assert.Equal(t, expected[0], sortedWithClonedRoute[0])
		assert.Equal(t, expected[1], sortedWithClonedRoute[1])
	}
}

func TestExtraSlashRoutePreparer_PrefixWithoutSlash_PrefixRewriteOnlySlash(t *testing.T) {
	extraSlashPreparer := NewExtraSlashRoutePreparer(routePreparerMock{})

	routes := []*domain.Route{
		{
			Prefix:        "/operation-manager",
			PrefixRewrite: "/",
		},
	}
	expected := []*domain.Route{
		{
			Prefix:        "/operation-manager/",
			PrefixRewrite: "/",
		},
		{
			Prefix:        "/operation-manager",
			PrefixRewrite: "/",
		},
	}

	withClonedRoute := extraSlashPreparer.Prepare(routes)

	assert.Equal(t, 2, len(withClonedRoute))
	assert.True(t, contains(withClonedRoute, expected[0]))
	assert.True(t, contains(withClonedRoute, expected[1]))
}

func contains(routes []*domain.Route, route *domain.Route) bool {
	for _, v := range routes {
		if reflect.DeepEqual(v, route) {
			return true
		}
	}

	return false
}

type routePreparerMock struct{}

func (r routePreparerMock) Prepare(routes []*domain.Route) []*domain.Route {
	return routes
}
