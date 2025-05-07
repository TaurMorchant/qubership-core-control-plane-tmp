package v2

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

// registration

func TestV2Service_RegisterRoutesConfig_BadNodeGroup(t *testing.T) {
	routingRequestResource, inMemDao := getRoutingRequestResource()
	configresources.RegisterResource(routingRequestResource)

	spec := []dto.RouteRegistrationRequest{createV2RouteRegistrationRequest(false, "namespace1")}
	config := configresources.ConfigResource{
		NodeGroup: "gateway1",
		Spec:      spec,
	}
	_, errCp := configresources.HandleConfigResource(nil, config)
	assert.Nil(t, errCp)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))

	config.NodeGroup = ""
	_, errCp = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, errCp)

	config.NodeGroup = "   "
	_, errCp = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, errCp)
}

func TestService_SingleOverriddenWithTrueValue(t *testing.T) {

	srv, _ := getV2Service()
	specs := []dto.RouteRegistrationRequest{createV2RouteRegistrationRequest(true, "namespace2")}
	isOverridden := srv.GetRegisterRoutesResource().GetDefinition().IsOverriddenByCR(nil, nil, &specs)
	assert.True(t, isOverridden)
}
func TestService_DifferentOverriddenValues(t *testing.T) {

	srv, _ := getV2Service()
	specs := []dto.RouteRegistrationRequest{createV2RouteRegistrationRequest(true, "namespace2"), createV2RouteRegistrationRequest(false, "namespace2"),
		createV2RouteRegistrationRequest(true, "namespace3")}
	isOverridden := srv.GetRegisterRoutesResource().GetDefinition().IsOverriddenByCR(nil, nil, &specs)
	assert.False(t, isOverridden)
}

func createV2RouteRegistrationRequest(overridden bool, namespace string) dto.RouteRegistrationRequest {
	return dto.RouteRegistrationRequest{
		Namespace:  namespace,
		Overridden: overridden,
		Cluster:    "test-cluster",
		Endpoint:   "http://tenant-manager-v1:8080",
		Routes: []dto.RouteItem{
			{
				Prefix:        "/api/v3/tenant-manager/tenant/{tenantId}/routes",
				PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
			},
		},
	}
}

func getRoutingRequestResource() (routingRequestResource, *dao.InMemDao) {
	v2Service, inMemDao := getV2Service()
	return routingRequestResource{v2Service, dto.RoutingV2RequestValidator{}}, inMemDao
}

func getV2Service() (*Service, *dao.InMemDao) {
	entityService := entity.NewService("v1")
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, []func([]memdb.Change) error{flushChanges})
	routeModeService := routingmode.NewService(genericDao, "v1")
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routeModeService)
	v2Service := NewV2Service(routeComponentsFactory, entityService, genericDao, eventBus, routeModeService, registrationService)
	_, _ = genericDao.WithWTx(func(dao dao.Repository) error {
		_ = dao.SaveDeploymentVersion(&domain.DeploymentVersion{
			Version: "v1",
			Stage:   domain.ActiveStage,
		})
		return nil
	})
	return v2Service, genericDao
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
