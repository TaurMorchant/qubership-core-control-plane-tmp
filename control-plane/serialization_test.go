package main

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/route"
	"github.com/netcracker/qubership-core-control-plane/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/services/route/v2"
	"github.com/netcracker/qubership-core-control-plane/services/routingmode"
	"math/rand"
	"testing"
	"time"
)

var (
	//log              = logging.GetLogger("ram")
	defaultNodeGroup = "private-gateway-service"
	letterRunes      = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestSerialization(t *testing.T) {
	/*
		gob.Register([]*domain.HeaderMatcher{})
		gob.Register([]*domain.NodeGroup{})
		gob.Register([]*domain.VirtualHost{})
		gob.Register([]*domain.VirtualHostDomain{})
		gob.Register([]*domain.DeploymentVersion{})
		gob.Register([]*domain.HashPolicy{})
		gob.Register([]*domain.Route{})
		gob.Register([]*domain.ClustersNodeGroup{})
		gob.Register([]*domain.RouteConfiguration{})
		gob.Register([]*domain.Listener{})
		gob.Register([]*domain.Cluster{})
		gob.Register([]*domain.EnvoyConfigVersion{})
		gob.Register([]*domain.Endpoint{})

		storage := ram.NewStorage()
		routeRegService := createTestRouteRegistrationService(storage)
		err := registerRandomRoutes(routeRegService, 50, 10)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		snapshot, _ := storage.Backup()
		b := bytes.Buffer{}
		e := gob.NewEncoder(&b)
		err = e.Encode(snapshot)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		log.Info("Size of snapshot is %d bytes", b.Len())

		decoder := gob.NewDecoder(&b)
		restoredSnapshot := &data.Snapshot{}
		err = decoder.Decode(restoredSnapshot)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		assert.True(t, len(snapshot.Data) == len(restoredSnapshot.Data))*/
}

type DummyIDGenerator struct {
	counter int32
}

func (g *DummyIDGenerator) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		g.counter = g.counter + 1
		uniqEntity.SetId(g.counter)
	}
	return nil
}

type MockEventBus struct{}

func (MockEventBus) Publish(topic string, data interface{}) error {
	return nil
}

func (MockEventBus) Shutdown() {
}

func createTestRouteRegistrationService(storage *ram.Storage) *v2.Service {
	entityService := entity.NewService("v1")
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	genericDao := dao.NewInMemDao(storage, &DummyIDGenerator{}, []func([]memdb.Change) error{})
	routingModeService := routingmode.NewService(genericDao, "v1")
	eventBus := &MockEventBus{}
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService)
	return v2.NewV2Service(routeComponentsFactory, entityService, genericDao, eventBus, routingModeService, registrationService)
}

func registerRandomRoutes(routeRegService *v2.Service, services, routesPerService int64) error {
	allowed := true
	registrationRequests := make([]dto.RouteRegistrationRequest, services)
	for i := int64(0); i < services; i++ {
		service := RandomAbstractService()
		routeItems := make([]dto.RouteItem, routesPerService)
		for j := int64(0); j < routesPerService; j++ {
			routeItems[j] = RandomRouteItem()
		}
		registrationRequests[i] = dto.RouteRegistrationRequest{
			Cluster:  service.Cluster,
			Routes:   routeItems,
			Endpoint: service.Endpoint,
			Allowed:  &allowed,
			Version:  "v1",
		}
	}
	for _, routeRegReq := range registrationRequests {
		err := routeRegService.RegisterRoutes(nil, []dto.RouteRegistrationRequest{routeRegReq}, defaultNodeGroup)
		if err != nil {
			return err
		}

	}
	return nil
	//return routeRegService.RegisterRoutes(registrationRequests, defaultNodeGroup)
}

type AbstractService struct {
	Cluster  string
	Endpoint string
}

func RandomAbstractService() *AbstractService {
	msName := RandStringRunes(25)
	return &AbstractService{
		Cluster:  msName,
		Endpoint: "http://" + msName + ":8080",
	}
}

func RandomRouteItem() dto.RouteItem {
	matchBase := RandStringRunes(rand.Intn(100))
	return dto.RouteItem{
		Prefix:        "/api/v1/" + matchBase,
		PrefixRewrite: "/api/v2/" + matchBase,
	}
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
