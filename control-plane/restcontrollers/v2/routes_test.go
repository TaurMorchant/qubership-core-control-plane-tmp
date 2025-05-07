package v2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dr"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	v2ctrl "github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/v2"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/creator"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/routekey"
	v2 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/v2"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
	mock_v2 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/restcontrollers/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
)

const (
	routeKeyFormat             = "%s||%s||%s"
	testVirtualHostName        = "test-nodegroup"
	testRouteConfigurationName = "test-nodegroup-routes"
	testNodeGroupName          = "test-nodegroup"
	testRoutesEndpoint         = "/api/v2/control-plane/routes/:nodeGroup"
	testRoutesUrlPath          = "/api/v2/control-plane/routes/" + testNodeGroupName
)

func TestRoutesController_HandleDeleteRoutes(t *testing.T) {
	t.SkipNow()
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)
	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"routes\":[{\"prefix\": \"/api/v1/test\"}]}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route1)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.Contains(t, actualRoutes, route2, route3, route4)
}

func TestRoutesController_HandleDeleteRoutesOnlyVersionWithBadRequest(t *testing.T) {
	v2Controller, _ := getV2Controller()
	responseRecorder := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("{\"version\": \"v3\"}"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusBadRequest, responseRecorder.StatusCode)
}

func TestRoutesController_HandleDeleteRoutesWithVersion(t *testing.T) {
	t.SkipNow()
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)

	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"routes\":[{\"prefix\":\"/api/v1/test\"}],\"version\":\"v2\"}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route4)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.Contains(t, actualRoutes, route1, route2, route3)
}

func TestRoutesController_HandleDeleteRoutesOnlyNamespacePresent(t *testing.T) {
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)

	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"namespace\": \"test-namespace\"}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
}

func TestRoutesController_HandleDeleteRoutesOnlyNamespacePresentWithPrefix(t *testing.T) {
	t.SkipNow()
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)

	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"namespace\":\"test-namespace\",\"routes\":[{\"prefix\":\"/api/v1/test\"}]}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route3)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.Contains(t, actualRoutes, route1, route2, route4)
}

func TestRoutesController_HandleDeleteRoutesOnlyNamespacePresentWithPrefixWithVersion(t *testing.T) {
	t.SkipNow()
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "test-namespace", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)

	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"namespace\":\"test-namespace\",\"routes\":[{\"prefix\":\"/api/v1/test\"}], \"version\": \"v2\"}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route4)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.Contains(t, actualRoutes, route1, route2, route3)
}

func TestRoutesController_HandleDeleteRoutesForbiddenRoute(t *testing.T) {
	t.SkipNow()
	v2Controller, inMemDao := getV2Controller()
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v2ctrl.SaveDeploymentVersions(t, inMemDao, v1, v2)
	route1 := createRoute("/api/v1/test", "", v1)
	route2 := createRoute("/api/v2/new-test", "", v1)
	route3 := createRoute("/api/v1/test", "test-namespace", v1)
	route4 := createRoute("/api/v1/test", "", v2)
	createAndSaveDefault(t, inMemDao, route1, route2, route3, route4)

	response := v2ctrl.SendHttpRequestWithBody(t, http.MethodDelete, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[{\"routes\":[{\"prefix\":\"/api/v1/test\"}],\"version\":\"v2\"}]"), v2Controller.DeleteRouteUnsecure)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	bytesBody, _ := io.ReadAll(response.Body)
	var deletedRoutes []*domain.Route
	err := json.Unmarshal(bytesBody, &deletedRoutes)
	assert.Nil(t, err)
	assert.NotEmpty(t, deletedRoutes)
	assert.Equal(t, 1, len(deletedRoutes))
	assert.Contains(t, deletedRoutes, route4)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	assert.Contains(t, actualRoutes, route1, route2, route3)
}

func TestRoutesController_HandlePostRoutesWithNodeGroup_ReturnsStatusServiceUnavailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	v2RouteService := mock_v2.NewMockRouteService(ctrl)

	eee1 := errorcodes.NewCpError(errorcodes.MasterNodeError, "Master node was switched at the moment. Please, Try again later", nil)
	v2RouteService.EXPECT().
		RegisterRoutes(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(eee1)

	v2Controller := v2ctrl.NewRoutesController(v2RouteService, dto.RoutingV2RequestValidator{})

	response := v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[]"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, errorcodes.StatusMasterNodeUnavailable, response.StatusCode)

	eee2 := entity.LegacyRouteDisallowed
	v2RouteService.EXPECT().
		RegisterRoutes(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(eee2)
	response = v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[]"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	eee3 := fmt.Errorf("unexpected error")
	v2RouteService.EXPECT().
		RegisterRoutes(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(eee3)
	response = v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[]"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	v2RouteService.EXPECT().
		RegisterRoutes(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	response = v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[]"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	response = v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[}"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	body, _ := io.ReadAll(response.Body)
	assert.Contains(t, string(body), "Could not unmarshal request body JSON")

	os.Setenv("EXECUTION_MODE", "standby")
	dr.ReloadMode()
	defer func() {
		os.Remove("EXECUTION_MODE")
		dr.ReloadMode()
	}()
	response = v2ctrl.SendHttpRequestWithBody(t,
		http.MethodPost, testRoutesEndpoint, testRoutesUrlPath,
		bytes.NewBufferString("[]"),
		v2Controller.HandlePostRoutesWithNodeGroup)
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

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

func createRoute(prefix, namespace string, version *domain.DeploymentVersion) *domain.Route {
	entry := creator.NewRouteEntry(prefix, "", namespace, 120, creator.GetInt64Timeout(nil), []*domain.HeaderMatcher{})
	route := entry.CreateRoute(0, prefix, "", "", 120, entry.GetIdleTimeout(), version.Version, version.Version, nil, nil, nil)
	route.RouteKey = routekey.GenerateKey(*route)
	return route
	/*return &domain.Route{
		RouteKey:                 fmt.Sprintf(routeKeyFormat, namespace, prefix, version.Version),
		Prefix:                   prefix,
		DeploymentVersion:        version.Version,
		InitialDeploymentVersion: version.Version,
		Uuid:                     uuid.New().String(),
		HostAutoRewrite:          domain.NewNullBool(false),
		Timeout:                  domain.NewNullInt(120),
	}*/
}

func getV2Controller() (*v2ctrl.RoutesController, *dao.InMemDao) {
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &v2ctrl.IdGeneratorMock{}, []func([]memdb.Change) error{v2ctrl.FlushChanges})
	entityService := entity.NewService("v1")
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	routingModeService := routingmode.NewService(genericDao, "v1")
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService)
	v2RouteService := v2.NewV2Service(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService, registrationService)
	v2RouteController := v2ctrl.NewRoutesController(v2RouteService, dto.RoutingV2RequestValidator{})
	return v2RouteController, genericDao
}
