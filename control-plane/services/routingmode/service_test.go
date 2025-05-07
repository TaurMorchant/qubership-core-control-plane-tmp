package routingmode

import (
	"github.com/google/uuid"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/stretchr/testify/assert"
	"os"
	"sync/atomic"
	"testing"
)

func TestService_SetRoutingMode(t *testing.T) {
	routingModeService, _ := getService()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	routingModeService.SetRoutingMode(NAMESPACED)
	assert.Equal(t, NAMESPACED, routingModeService.GetRoutingMode())
}

func TestService_UpdateRouteModeDetailsWithSimpleMode(t *testing.T) {
	routingModeService, _ := getService()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	summary := routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	assert.Empty(t, summary.RouteKeys)
	assert.Equal(t, SIMPLE, summary.RoutingMode)
}

func TestService_UpdateRouteModeDetailsWithMixedMode(t *testing.T) {
	routingModeService, dao := getService()
	saveSeveralDeploymentVersionsWithRoutes(t, dao)
	saveSeveralRouteWithNamespaceHeader(t, dao)
	summary := routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, MIXED, routingModeService.GetRoutingMode())
	assert.NotEmpty(t, summary.RouteKeys)
	assert.Equal(t, 4, len(summary.RouteKeys))
	assert.Equal(t, MIXED, summary.RoutingMode)
}

func TestService_UpdateRouteModeDetailsWithVersionedMode(t *testing.T) {
	routingModeService, dao := getService()
	saveSeveralDeploymentVersionsWithRoutes(t, dao)
	summary := routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
	assert.NotEmpty(t, summary.RouteKeys)
	assert.Equal(t, 3, len(summary.RouteKeys))
	assert.Equal(t, VERSIONED, summary.RoutingMode)
}

func TestService_UpdateRouteModeDetailsWithNamespacedMode(t *testing.T) {
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	summary := routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, NAMESPACED, routingModeService.GetRoutingMode())
	assert.NotEmpty(t, summary.RouteKeys)
	assert.Equal(t, 1, len(summary.RouteKeys))
	assert.Equal(t, NAMESPACED, summary.RoutingMode)
}

func TestService_UpdateRouteModeDetailsForCurrentNamespace(t *testing.T) {
	os.Setenv(msaddr.CloudNamespace, "some-namespace")
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	summary := routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	assert.Empty(t, summary.RouteKeys)
	assert.Equal(t, SIMPLE, summary.RoutingMode)
	os.Unsetenv(msaddr.CloudNamespace)
}

func TestService_UpdateRoutingModeWithVersionedMode(t *testing.T) {
	routingModeService, _ := getService()
	assert.True(t, routingModeService.UpdateRoutingMode("v2", msaddr.NewNamespace("test-namespace")))
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
}

func TestService_UpdateRoutingModeWithNamespacedMode(t *testing.T) {
	routingModeService, _ := getService()
	assert.True(t, routingModeService.UpdateRoutingMode("v1", msaddr.NewNamespace("test-namespace")))
	assert.Equal(t, NAMESPACED, routingModeService.GetRoutingMode())
}

func TestService_UpdateRoutingModeWithoutUpdateFromVersioned(t *testing.T) {
	testVersion := "v2"
	testNamespace := "test-namespace"
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	assert.True(t, routingModeService.UpdateRoutingMode(testVersion, msaddr.NewNamespace(testNamespace)))
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
	os.Setenv(msaddr.CloudNamespace, testNamespace)
	assert.False(t, routingModeService.UpdateRoutingMode(testVersion, msaddr.NewNamespace(testNamespace)))
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
	os.Unsetenv(msaddr.CloudNamespace)
}

func TestService_UpdateRoutingModeWithoutUpdateForCurrentNamespaceAndSameVersion(t *testing.T) {
	testNamespace := "test-namespace"
	os.Setenv(msaddr.CloudNamespace, testNamespace)
	routingModeService, _ := getService()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	assert.False(t, routingModeService.UpdateRoutingMode("v1", msaddr.NewNamespace(testNamespace)))
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	os.Unsetenv(msaddr.CloudNamespace)
}

func TestService_IsForbiddenRoutingModeForSimple(t *testing.T) {
	routingModeService, _ := getService()
	assert.Equal(t, SIMPLE, routingModeService.GetRoutingMode())
	assert.False(t, routingModeService.IsForbiddenRoutingMode("v1", "test-namespace"))
}

func TestService_IsForbiddenRoutingModeForNamespacedWithDifferentVersion(t *testing.T) {
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, NAMESPACED, routingModeService.GetRoutingMode())
	assert.True(t, routingModeService.IsForbiddenRoutingMode("v2", "test-namespace"))
}

func TestService_IsForbiddenRoutingModeForNamespacedWithSameVersion(t *testing.T) {
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	routingModeService.UpdateRouteModeDetails()
	assert.Equal(t, NAMESPACED, routingModeService.GetRoutingMode())
	assert.False(t, routingModeService.IsForbiddenRoutingMode("v1", "test-namespace"))
}

func TestService_IsForbiddenRoutingModeForVersionedWithNotCurrentNamespace(t *testing.T) {
	testVersion := "v2"
	testNamespace := "test-namespace"
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	routingModeService.UpdateRoutingMode(testVersion, msaddr.NewNamespace(testNamespace))
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
	assert.True(t, routingModeService.IsForbiddenRoutingMode(testVersion, testNamespace))
}

func TestService_IsForbiddenRoutingModeForVersionedWithCurrentNamespace(t *testing.T) {
	testVersion := "v2"
	testNamespace := "test-namespace"
	os.Setenv(msaddr.CloudNamespace, testNamespace)
	routingModeService, dao := getService()
	saveSeveralRouteWithNamespaceHeader(t, dao)
	routingModeService.UpdateRoutingMode(testVersion, msaddr.NewNamespace(testNamespace))
	assert.Equal(t, VERSIONED, routingModeService.GetRoutingMode())
	assert.False(t, routingModeService.IsForbiddenRoutingMode(testVersion, testNamespace))
	os.Unsetenv(msaddr.CloudNamespace)
}

func getService() (*Service, *dao.InMemDao) {
	inMemStorage := ram.NewStorage()
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, nil)
	return NewService(genericDao, "v1"), genericDao
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

func saveSeveralDeploymentVersionsWithRoutes(t *testing.T, memDao *dao.InMemDao) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		for _, version := range []string{"v1", "v2", "v3"} {
			assert.Nil(t, dao.SaveDeploymentVersion(domain.NewDeploymentVersion(version, domain.CandidateStage)))
			assert.Nil(t, dao.SaveRoute(&domain.Route{DeploymentVersion: version, RouteKey: version, Uuid: uuid.New().String()}))
		}
		return nil
	})
	assert.Nil(t, err)
}

func saveSeveralRouteWithNamespaceHeader(t *testing.T, memDao *dao.InMemDao) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveDeploymentVersion(domain.NewDeploymentVersion("v4", domain.ActiveStage)))
		assert.Nil(t, dao.SaveRoute(&domain.Route{Id: 10, RouteKey: "test-key", Uuid: uuid.New().String(), DeploymentVersion: "v4"}))
		assert.Nil(t, dao.SaveHeaderMatcher(&domain.HeaderMatcher{RouteId: 10, Name: "namespace", ExactMatch: "some-namespace"}))
		return nil
	})
	assert.Nil(t, err)
}
