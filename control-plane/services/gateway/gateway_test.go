package gateway

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	mock_bus "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestService_ApplyAndDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inMemDao := newInMemDao()
	entitySrv := entity.NewService("v1")
	mockBus := mock_bus.NewMockBusPublisher(ctrl)

	srv := NewService(inMemDao, entitySrv, mockBus)

	inMemDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
		assert.Nil(t, err)
		return nil
	})

	expected := domain.NodeGroup{
		Name:               "integration-gateway",
		GatewayType:        domain.Ingress,
		ForbidVirtualHosts: true,
	}

	// create gateway declaration
	mockBus.EXPECT().Publish(bus.TopicChanges, gomock.Any())

	_, err := srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:              "integration-gateway",
		GatewayType:       domain.Ingress,
		AllowVirtualHosts: util.WrapValue(false),
	})
	assert.Nil(t, err)

	nodeGroup, err := inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.Equal(t, expected, *nodeGroup)

	// delete gateway declaration
	mockBus.EXPECT().Publish(bus.TopicChanges, gomock.Any())

	_, err = srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:   "integration-gateway",
		Exists: util.WrapValue(false),
	})
	assert.Nil(t, err)

	nodeGroup, err = inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.Nil(t, nodeGroup)
}

func TestService_DeleteFailure_HasCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inMemDao := newInMemDao()
	entitySrv := entity.NewService("v1")
	mockBus := mock_bus.NewMockBusPublisher(ctrl)

	srv := NewService(inMemDao, entitySrv, mockBus)

	inMemDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
		assert.Nil(t, err)
		return nil
	})

	expected := domain.NodeGroup{
		Name:               "integration-gateway",
		GatewayType:        domain.Ingress,
		ForbidVirtualHosts: true,
	}

	// create gateway declaration
	mockBus.EXPECT().Publish(bus.TopicChanges, gomock.Any())

	_, err := srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:              "integration-gateway",
		GatewayType:       domain.Ingress,
		AllowVirtualHosts: util.WrapValue(false),
	})
	assert.Nil(t, err)

	// create bounded cluster
	inMemDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveCluster(&domain.Cluster{
			Id:      1,
			Name:    "public-gateway-service||public-gateway-service||8080",
			Version: 1,
		})
		assert.Nil(t, err)
		err = repo.SaveClustersNodeGroup(&domain.ClustersNodeGroup{ClustersId: 1, NodegroupsName: "integration-gateway"})
		assert.Nil(t, err)
		return nil
	})

	nodeGroup, err := inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.Equal(t, expected, *nodeGroup)

	// delete gateway declaration and expect failure

	_, err = srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:   "integration-gateway",
		Exists: util.WrapValue(false),
	})
	assert.NotNil(t, err)
	assert.Equal(t, ErrHasCluster, err)

	nodeGroup, err = inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.NotNil(t, nodeGroup)
}

func TestService_DeleteFailure_HasTlsConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inMemDao := newInMemDao()
	entitySrv := entity.NewService("v1")
	mockBus := mock_bus.NewMockBusPublisher(ctrl)

	srv := NewService(inMemDao, entitySrv, mockBus)

	inMemDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
		assert.Nil(t, err)
		return nil
	})

	expected := domain.NodeGroup{
		Name:               "integration-gateway",
		GatewayType:        domain.Ingress,
		ForbidVirtualHosts: true,
	}

	// create gateway declaration
	mockBus.EXPECT().Publish(bus.TopicChanges, gomock.Any())

	_, err := srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:              "integration-gateway",
		GatewayType:       domain.Ingress,
		AllowVirtualHosts: util.WrapValue(false),
	})
	assert.Nil(t, err)

	// create bounded cluster
	inMemDao.WithWTx(func(repo dao.Repository) error {
		err = repo.SaveTlsConfig(&domain.TlsConfig{
			Id:        1,
			Name:      "integration-gw-tls",
			Enabled:   true,
			TrustedCA: "CA",
		})
		assert.Nil(t, err)
		err := repo.SaveCluster(&domain.Cluster{
			Id:      1,
			Name:    "public-gateway-service||public-gateway-service||8080",
			TLSId:   1,
			Version: 1,
		})
		assert.Nil(t, err)
		err = repo.SaveClustersNodeGroup(&domain.ClustersNodeGroup{ClustersId: 1, NodegroupsName: "integration-gateway"})
		assert.Nil(t, err)
		return nil
	})

	nodeGroup, err := inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.Equal(t, expected, *nodeGroup)

	// delete gateway declaration and expect failure

	_, err = srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:   "integration-gateway",
		Exists: util.WrapValue(false),
	})
	assert.NotNil(t, err)
	assert.Equal(t, ErrHasCluster, err)

	nodeGroup, err = inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.NotNil(t, nodeGroup)
}

func TestService_DeleteFailure_HasVirtualHost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inMemDao := newInMemDao()
	entitySrv := entity.NewService("v1")
	mockBus := mock_bus.NewMockBusPublisher(ctrl)

	srv := NewService(inMemDao, entitySrv, mockBus)

	inMemDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
		assert.Nil(t, err)
		return nil
	})

	expected := domain.NodeGroup{
		Name:               "integration-gateway",
		GatewayType:        domain.Ingress,
		ForbidVirtualHosts: true,
	}

	// create gateway declaration
	mockBus.EXPECT().Publish(bus.TopicChanges, gomock.Any())

	_, err := srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:              "integration-gateway",
		GatewayType:       domain.Ingress,
		AllowVirtualHosts: util.WrapValue(false),
	})
	assert.Nil(t, err)

	// create bounded routes cluster
	inMemDao.WithWTx(func(repo dao.Repository) error {
		assert.Nil(t, err)
		err = repo.SaveRouteConfig(&domain.RouteConfiguration{
			Id:          1,
			Name:        "integration-gateway-routes",
			Version:     1,
			NodeGroupId: "integration-gateway",
		})
		assert.Nil(t, err)
		err = repo.SaveVirtualHost(&domain.VirtualHost{
			Id:                   1,
			Name:                 "integration-gateway-routes",
			Version:              1,
			RouteConfigurationId: 1,
		})
		assert.Nil(t, err)
		return nil
	})

	nodeGroup, err := inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.Equal(t, expected, *nodeGroup)

	// delete gateway declaration and expect failure

	_, err = srv.Apply(context.Background(), dto.GatewayDeclaration{
		Name:   "integration-gateway",
		Exists: util.WrapValue(false),
	})
	assert.NotNil(t, err)
	assert.Equal(t, ErrHasVirtualHost, err)

	nodeGroup, err = inMemDao.FindNodeGroupByName("integration-gateway")
	assert.Nil(t, err)
	assert.NotNil(t, nodeGroup)
}

func newInMemDao() *dao.InMemDao {
	return dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, []func([]memdb.Change) error{})
}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}
