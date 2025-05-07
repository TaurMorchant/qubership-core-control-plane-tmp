package testutils

import (
	"github.com/hashicorp/go-memdb"
	cpCfg "github.com/netcracker/qubership-core-control-plane/control-plane/v2/configuration"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	v1 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/v1"
	v2 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/v2"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/bluegreen"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/loadbalance"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	v12 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/v1"
	v22 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/v2"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
)

type TestEnvironment struct {
	Dao                   *dao.InMemDao
	EntityService         *entity.Service
	EventBus              *bus.EventBusAggregator
	RouteServiceV2        *v22.Service
	RouteControllerV1     *v1.Controller
	RouteControllerV2     *v2.RoutesController
	BlueGreenControllerV2 *v2.BlueGreenController
}

func InitCPTestEnvironment() TestEnvironment {
	defaultVersion := "v1"
	entityService := entity.NewService(defaultVersion)
	inMemStorage := ram.NewStorage()
	genericDao := dao.NewInMemDao(inMemStorage, &generatorMock{}, []func([]memdb.Change) error{})
	commonConfiguration := cpCfg.NewCommonConfiguration(genericDao, entityService, false)
	_, err := genericDao.WithWTx(func(dao dao.Repository) error {
		err := dao.SaveNodeGroup(&domain.NodeGroup{Name: "private-gateway-service"})
		if err != nil {
			return err
		}
		err = dao.SaveNodeGroup(&domain.NodeGroup{Name: "public-gateway-service"})
		if err != nil {
			return err
		}
		err = dao.SaveNodeGroup(&domain.NodeGroup{Name: "internal-gateway-service"})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	commonConfiguration.CreateCommonConfiguration()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(genericDao, internalBus, internalBus, nil, nil)
	routingModeService := routingmode.NewService(genericDao, defaultVersion)
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService)
	v2RouteService := v22.NewV2Service(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService, registrationService)
	v2RouteController := v2.NewRoutesController(v2RouteService, dto.RoutingV2RequestValidator{})
	v1RouteService := v12.NewV1Service(entityService, genericDao, eventBus, routingModeService, registrationService)
	v1RouteController := v1.NewController(v1RouteService, dto.RoutingV1RequestValidator{})
	loadBalanceService := loadbalance.NewLoadBalanceService(genericDao, entityService, eventBus)
	bgRegistry := bluegreen.NewVersionsRegistry(genericDao, entityService, eventBus)
	blueGreenService := bluegreen.NewService(entityService, loadBalanceService, genericDao, eventBus, bgRegistry)
	v2BlueGreenController := v2.NewBlueGreenController(blueGreenService, genericDao)

	return TestEnvironment{
		Dao:                   genericDao,
		EntityService:         entityService,
		EventBus:              eventBus,
		RouteServiceV2:        v2RouteService,
		RouteControllerV1:     v1RouteController,
		RouteControllerV2:     v2RouteController,
		BlueGreenControllerV2: v2BlueGreenController,
	}
}

type generatorMock struct {
	inc int32
}

func (g *generatorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		g.inc++
		uniqEntity.SetId(g.inc)
	}
	return nil
}
