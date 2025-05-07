package v1

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/creator"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/routekey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync/atomic"
	"testing"
)

const (
	pathFrom                   = "/foo/from"
	pathTo                     = "/foo/to"
	pathWithVariable           = "/api/v1/test/{variable}/details"
	microserviceUrl            = "http://foo-bar:8080"
	defaultNamespace           = "default"
	localDevNamespace          = "192.168.0.1.nip.io"
	defaultCluster             = "foo-bar||foo-bar||8080"
	localDevCluster            = "foo-bar||foo-bar.192.168.0.1.nip.io||8080"
	testNodeGroup              = "test-nodegroup"
	testTimeout                = int64(150000)
	routeKeyFormat             = "%s||%s||%s"
	testVirtualHostName        = "test-nodegroup"
	testRouteConfigurationName = "test-nodegroup-routes"
	testNodeGroupName          = "test-nodegroup"
)

var tCtx = context.Background()

// --------------route registration

func TestRegisterRoutes_MustBeSetDefaultDeploymentVersionIfNotPresent(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, defaultNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, "v1", routes[0].DeploymentVersion)
}

func TestRegisterRoutes_WithVariableAndRegExpMatcher(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	path1 := "/api/v1/ext-frontend-api/customers/{id}/subscriptions"
	routeKey1 := routekey.Generate(routekey.RouteMatch{
		Regexp:  "/api/v1/ext-frontend-api/customers/([^/]+)/subscriptions(/.*)?",
		Version: "v1",
	})
	request := createV1Request(microserviceUrl, true, path1, path1, "")
	err := v1Service.RegisterRoutes(nil, testNodeGroupName, request)
	assert.Nil(t, err)

	path2 := "/api/v1/ext-frontend-api/customers/{id}"
	routeKey2 := routekey.Generate(routekey.RouteMatch{
		Regexp:  "/api/v1/ext-frontend-api/customers/([^/]+)(/.*)?",
		Version: "v1",
	})
	request = createV1Request(microserviceUrl, false, path2, path2, "")
	err = v1Service.RegisterRoutes(nil, testNodeGroupName, request)
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	routeAllowedWithVariable := routes[0]
	routeForbiddenWithVariable := routes[1]

	assert.Equal(t, routeKey1, routeAllowedWithVariable.RouteKey)
	assert.Equal(t, "/api/v1/ext-frontend-api/customers/([^/]+)/subscriptions(/.*)?", routeAllowedWithVariable.Regexp)
	assert.Equal(t, "", routeAllowedWithVariable.Prefix)
	assert.Equal(t, uint32(0), routeAllowedWithVariable.DirectResponseCode)

	assert.Equal(t, routeKey2, routeForbiddenWithVariable.RouteKey)
	assert.Equal(t, "/api/v1/ext-frontend-api/customers/([^/]+)(/.*)?", routeForbiddenWithVariable.Regexp)
	assert.Equal(t, "", routeForbiddenWithVariable.Prefix)
	assert.Equal(t, uint32(http.StatusNotFound), routeForbiddenWithVariable.DirectResponseCode)
}

func TestRegisterRoutes_HaveToCreateLocalClusterWhenNamespaceSet(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, localDevNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, localDevCluster, clusters[0].Name)
}

func TestRegisterRoutes_HaveToCreatePlainClusterWhenNamespaceSet(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, defaultNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, defaultCluster, clusters[0].Name)
}

func TestRegisterRoutes_HaveToCreateRouteWithPlainClusterWhenNamespaceNotSet(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, "")
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, defaultCluster, clusters[0].Name)
}

func TestRegisterRoutes_HaveToAddNormalRouteWhenRouteIsAllowed(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, defaultNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, pathFrom, routes[0].Prefix)
	assert.Equal(t, pathTo, routes[0].PrefixRewrite)
}

func TestRegisterRoutes_HaveToAddNormalRouteWhenRouteIsForbidden(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, false, pathFrom, pathTo, defaultNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, pathFrom, routes[0].Prefix)
	assert.NotEmpty(t, routes[0].ClusterName)
	assert.NotEmpty(t, routes[0].HostRewrite)
	assert.Equal(t, uint32(http.StatusNotFound), routes[0].DirectResponseCode)
}

func TestRegisterRoutes_HaveToAddNormalRouteWithTimeoutWhenTimeoutNotNull(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	request := createV1Request(microserviceUrl, true, pathFrom, pathTo, defaultNamespace)
	assert.Equal(t, int64(-1), request.GetRoutes()[0].GetTimeout())
	request = createRequestWithTimeout(microserviceUrl, true, pathFrom, pathTo, defaultNamespace, testTimeout)
	assert.Equal(t, testTimeout, request.GetRoutes()[0].GetTimeout())
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, pathFrom, routes[0].Prefix)
	assert.Equal(t, pathTo, routes[0].PrefixRewrite)
	assert.True(t, routes[0].Timeout.Valid)
	assert.Equal(t, testTimeout, routes[0].Timeout.Int64)
}

// --------------route delete

func TestDeleteRoutes_UsingRoutePrefix(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	saveDeploymentVersions(t, inMemDao, v1, v2)
	vHost := createAndSaveRouteConfig(t, inMemDao)
	route1 := createRoute(vHost.Id, "/api/v1/test", "", v1)
	route2 := createRoute(vHost.Id, "/api/v2/new-test", "", v1)
	route3 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v1)
	route4 := createRoute(vHost.Id, "/api/v1/test", "", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)
	createAndSaveDefault(t, inMemDao, route1)
	deletedRoutes, err := v1Service.DeleteRoutes(tCtx, testNodeGroupName, "/api/v1/test", "")
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route1)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.NotContains(t, actualRoutes, route1)
}

func TestDeleteRoutes_UsingNamespace(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	saveDeploymentVersions(t, inMemDao, v1, v2)
	vHost := createAndSaveRouteConfig(t, inMemDao)
	route1 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v1)
	route2 := createRoute(vHost.Id, "/api/v1/test", "another-namespace", v1)
	route3 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3)
	deletedRoutes, err := v1Service.DeleteRoutes(nil, testNodeGroupName, "/api/v1/test", "test-namespace")
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route1)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 2, len(actualRoutes))
	assert.NotContains(t, actualRoutes, route1)
}

func TestDeleteRoutes_UsingOnlyVersion(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	saveDeploymentVersions(t, inMemDao, v1, v2)
	vHost := createAndSaveRouteConfig(t, inMemDao)
	route1 := createRoute(vHost.Id, "/api/v1/test", "", v1)
	route2 := createRoute(vHost.Id, "/api/v1/test", "", v2)
	route3 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3)
	deletedRoutes, err := v1Service.DeleteRoutes(tCtx, testNodeGroupName, "/api/v1/test", "")
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route1)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 2, len(actualRoutes))
	assert.NotContains(t, actualRoutes, route1)
}

func TestDeleteNonExistentCluster(t *testing.T) {
	v1Service, _ := getV1Service()
	err := v1Service.DeleteCluster(99999)
	assert.NotNil(t, err)
	assert.Equal(t, "Cluster does not exist for passed clusterId", err.Error())
}

func TestDeleteRoutes_NarrowInput(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	saveDeploymentVersions(t, inMemDao, v1, v2)
	vHost := createAndSaveRouteConfig(t, inMemDao)
	route1 := createRoute(vHost.Id, "/api/v1/test", "", v1)
	route2 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v1)
	route3 := createRoute(vHost.Id, "/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3)
	deletedRoutes, err := v1Service.DeleteRoutes(tCtx, testNodeGroupName, "/api/v1/test", "test-namespace")
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route2)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 2, len(actualRoutes))
	assert.NotContains(t, actualRoutes, route2)
}

func TestDeleteRoutes_AllParametersAreEmpty(t *testing.T) {
	v1Service, _ := getV1Service()
	assert.NotPanics(t, func() {
		deletedRoutes, err := v1Service.DeleteRoutes(tCtx, "", "", "")
		assert.NotNil(t, err)
		assert.Empty(t, deletedRoutes)
	})
}

func TestDeleteRoutes_PathVariablePrefix(t *testing.T) {
	v1Service, inMemDao := getV1Service()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	saveDeploymentVersions(t, inMemDao, v1)
	request := createV1Request(microserviceUrl, true, pathWithVariable, pathWithVariable, defaultNamespace)
	err := v1Service.RegisterRoutes(nil, testNodeGroup, request)
	assert.Nil(t, err)

	deletedRoutes, err := v1Service.DeleteRoutes(tCtx, testNodeGroupName, pathWithVariable, defaultNamespace)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Equal(t, defaultCluster, deletedRoutes[0].ClusterName)
	assert.Equal(t, "/api/v1/test/([^/]+)/details(/.*)?", deletedRoutes[0].Regexp)
	assert.Equal(t, "foo-bar:8080", deletedRoutes[0].HostRewrite)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.Empty(t, actualRoutes)
}

//-----PRIVATE

func createAndSaveDefault(t *testing.T, memStorage *dao.InMemDao, routes ...*domain.Route) {
	_, err := memStorage.WithWTx(func(dao dao.Repository) error {
		virtualHost := createAndSaveRouteConfig(t, dao)
		for _, route := range routes {
			route.VirtualHostId = virtualHost.Id
			assert.Nil(t, dao.SaveRoute(route))
			if len(route.HeaderMatchers) > 0 {
				for _, headerMatcher := range route.HeaderMatchers {
					headerMatcher.RouteId = route.Id
					assert.Nil(t, dao.SaveHeaderMatcher(headerMatcher))
				}
			}
		}
		return nil
	})
	assert.Empty(t, err)
}

func createAndSaveRouteConfig(t *testing.T, repository dao.Repository) *domain.VirtualHost {
	nodeGroup := domain.NewNodeGroup(testNodeGroupName)
	assert.Nil(t, repository.SaveNodeGroup(nodeGroup))
	routeConfig := domain.NewRouteConfiguration(testRouteConfigurationName, testNodeGroupName)
	assert.Nil(t, repository.SaveRouteConfig(routeConfig))
	virtualHost := domain.NewVirtualHost(testVirtualHostName, routeConfig.Id)
	assert.Nil(t, repository.SaveVirtualHost(virtualHost))
	return virtualHost
}

func saveDeploymentVersions(t *testing.T, storage *dao.InMemDao, dVs ...*domain.DeploymentVersion) {
	_, err := storage.WithWTx(func(dao dao.Repository) error {
		for _, dV := range dVs {
			assert.Nil(t, dao.SaveDeploymentVersion(dV))
		}
		return nil
	})
	assert.Nil(t, err)
}

func createRoute(vHostId int32, prefix, namespace string, version *domain.DeploymentVersion) *domain.Route {
	headerMatchers := make([]*domain.HeaderMatcher, 0)
	if namespace != "" {
		headerMatcher := &domain.HeaderMatcher{
			Name:       "namespace",
			ExactMatch: namespace,
		}
		headerMatchers = append(headerMatchers, headerMatcher)
	}
	return &domain.Route{
		RouteKey:                 fmt.Sprintf(routeKeyFormat, namespace, prefix, version.Version),
		Prefix:                   prefix,
		DeploymentVersion:        version.Version,
		InitialDeploymentVersion: version.Version,
		Uuid:                     uuid.New().String(),
		HeaderMatchers:           headerMatchers,
		VirtualHostId:            vHostId,
	}
}

func flushChanges(changes []memdb.Change) error {
	return nil
}

type idGeneratorMock struct {
	seq int32
}

func (generator *idGeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&generator.seq, 1))
	}
	return nil
}

func createRequestWithTimeout(microserviceUrl string, allowed bool, from, to, namespace string, timeout int64) *creator.RouteEntityRequest {
	routeRequest := creator.NewRouteRequest(microserviceUrl, allowed)
	routeEntry := creator.NewRouteEntry(from, to, namespace, timeout, creator.DefaultIdleTimeoutSpec, []*domain.HeaderMatcher{})
	return creator.NewRouteEntityRequest(&routeRequest, []*creator.RouteEntry{routeEntry})
}

func createV1Request(microserviceUrl string, allowed bool, from, to, namespace string) *creator.RouteEntityRequest {
	routeRequest := creator.NewRouteRequest(microserviceUrl, allowed)
	routeEntry := creator.NewRouteEntry(from, to, namespace, creator.DefaultTimeoutSpec, creator.DefaultIdleTimeoutSpec, []*domain.HeaderMatcher{})
	return creator.NewRouteEntityRequest(&routeRequest, []*creator.RouteEntry{routeEntry})
}

func getV1Service() (*Service, *dao.InMemDao) {
	entityService := entity.NewService("v1")
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, []func([]memdb.Change) error{flushChanges})
	routeModeService := routingmode.NewService(genericDao, "v1")
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routeModeService)
	v1Service := NewV1Service(entityService, genericDao, eventBus, routeModeService, registrationService)
	return v1Service, genericDao
}
