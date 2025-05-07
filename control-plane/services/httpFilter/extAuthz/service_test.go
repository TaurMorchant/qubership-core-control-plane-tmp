package extAuthz

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/constancy"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/registration"
	mock_dao "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/dao"
	mock_bus "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/event/bus"
	mock_route "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestService_Apply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := newInMemDao()
	busMock := mock_bus.NewMockBusPublisher(ctrl)
	clusterRegSrv := mock_route.NewMockClusterRegistrationService(ctrl)

	v3RequestProcessor := registration.NewV3RequestProcessor(daoMock)
	srv := NewService(daoMock, busMock, entity.NewService("v1"), clusterRegSrv, v3RequestProcessor)

	daoMock.WithWTx(func(dao dao.Repository) error {
		err := dao.SaveDeploymentVersion(&domain.DeploymentVersion{
			Version:     "v1",
			Stage:       domain.ActiveStage,
			CreatedWhen: time.Now(),
			UpdatedWhen: time.Now(),
		})
		assert.Nil(t, err)
		return nil
	})

	extAuthz := dto.ExtAuthz{
		Name:              "name",
		Destination:       dto.RouteDestination{Cluster: "cluster", Endpoint: "cluster:8080"},
		ContextExtensions: map[string]string{"key1": "val1"},
		Timeout:           util.WrapValue(int64(10000)),
	}

	clusterRegSrv.EXPECT().SaveCluster(gomock.Any(), gomock.Any(), gomock.Any(), "", "gw").Return(nil)
	busMock.EXPECT().Publish(bus.TopicChanges, gomock.Any()).Return(nil)

	err := srv.Apply(context.Background(), extAuthz, "gw")
	assert.Nil(t, err)

	nodeGroup, err := daoMock.FindNodeGroupByName("gw")
	assert.Nil(t, err)
	assert.Equal(t, "gw", nodeGroup.Name)

	actualFilter, err := daoMock.FindExtAuthzFilterByName("name")
	assert.Nil(t, err)
	assert.Equal(t, domain.ExtAuthzFilter{
		Name:              "name",
		ClusterName:       "cluster||cluster||8080",
		Timeout:           int64(10000),
		ContextExtensions: map[string]string{"key1": "val1"},
		NodeGroup:         "gw",
	}, *actualFilter)

	err = srv.Apply(context.Background(), extAuthz, "gw2")
	assert.Equal(t, ErrNameTaken, err)

	clusterRegSrv.EXPECT().SaveCluster(gomock.Any(), gomock.Any(), gomock.Any(), "", "gw").Return(nil)
	busMock.EXPECT().Publish(bus.TopicChanges, gomock.Any()).Return(errors.New("expected test err"))

	err = srv.Apply(context.Background(), extAuthz, "gw")
	assert.NotNil(t, err)

	clusterRegSrv.EXPECT().SaveCluster(gomock.Any(), gomock.Any(), gomock.Any(), "", "gw").Return(errors.New("expected test err"))

	err = srv.Apply(context.Background(), extAuthz, "gw")
	assert.NotNil(t, err)
}

func TestService_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := newInMemDao()
	busMock := mock_bus.NewMockBusPublisher(ctrl)
	clusterRegSrv := mock_route.NewMockClusterRegistrationService(ctrl)

	v3RequestProcessor := registration.NewV3RequestProcessor(daoMock)
	srv := NewService(daoMock, busMock, entity.NewService("v1"), clusterRegSrv, v3RequestProcessor)

	daoMock.WithWTx(func(dao dao.Repository) error {
		err := dao.SaveDeploymentVersion(&domain.DeploymentVersion{
			Version:     "v1",
			Stage:       domain.ActiveStage,
			CreatedWhen: time.Now(),
			UpdatedWhen: time.Now(),
		})
		assert.Nil(t, err)
		return nil
	})

	extAuthz := dto.ExtAuthz{
		Name:              "name",
		Destination:       dto.RouteDestination{Cluster: "cluster", Endpoint: "cluster:8080"},
		ContextExtensions: map[string]string{"key1": "val1"},
		Timeout:           util.WrapValue(int64(10000)),
	}

	clusterRegSrv.EXPECT().SaveCluster(gomock.Any(), gomock.Any(), gomock.Any(), "", "gw").Return(nil)
	busMock.EXPECT().Publish(bus.TopicChanges, gomock.Any()).Return(nil)

	err := srv.Apply(context.Background(), extAuthz, "gw")
	assert.Nil(t, err)

	nodeGroup, err := daoMock.FindNodeGroupByName("gw")
	assert.Nil(t, err)
	assert.Equal(t, "gw", nodeGroup.Name)

	actualFilter, err := daoMock.FindExtAuthzFilterByName("name")
	assert.Nil(t, err)
	assert.Equal(t, domain.ExtAuthzFilter{
		Name:              "name",
		ClusterName:       "cluster||cluster||8080",
		Timeout:           int64(10000),
		ContextExtensions: map[string]string{"key1": "val1"},
		NodeGroup:         "gw",
	}, *actualFilter)

	err = srv.Delete(context.Background(), extAuthz, "gw2")
	assert.Equal(t, ErrNameTaken, err)

	busMock.EXPECT().Publish(bus.TopicChanges, gomock.Any()).Return(nil)

	err = srv.Delete(context.Background(), extAuthz, "gw")
	assert.Nil(t, err)

	filter, err := daoMock.FindExtAuthzFilterByName("name")
	assert.Nil(t, err)
	assert.Nil(t, filter)
}

func TestService_ValidateApply(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	busMock := mock_bus.NewMockBusPublisher(ctrl)
	clusterRegSrv := mock_route.NewMockClusterRegistrationService(ctrl)

	v3RequestProcessor := registration.NewV3RequestProcessor(daoMock)
	srv := NewService(daoMock, busMock, entity.NewService("v1"), clusterRegSrv, v3RequestProcessor)

	extAuthz := dto.ExtAuthz{
		Name:        "name",
		Destination: dto.RouteDestination{Cluster: "cluster", Endpoint: "endpoint", TlsEndpoint: "tls-endpoint"},
	}

	ok, _ := srv.ValidateApply(context.Background(), extAuthz)
	assert.False(t, ok)

	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw1", "gw2")
	assert.False(t, ok)

	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "")
	assert.False(t, ok)

	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Name = ""
	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.False(t, ok)

	extAuthz.Name = "name"
	extAuthz.Destination.Cluster = ""
	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.False(t, ok)

	extAuthz.Destination.Cluster = "cluster"
	extAuthz.Destination.Endpoint = ""
	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Destination.TlsEndpoint = ""
	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.False(t, ok)

	extAuthz.Destination.Endpoint = "endpoint"
	ok, _ = srv.ValidateApply(context.Background(), extAuthz, "gw")
	assert.True(t, ok)
}

func TestService_ValidateDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	busMock := mock_bus.NewMockBusPublisher(ctrl)
	clusterRegSrv := mock_route.NewMockClusterRegistrationService(ctrl)

	v3RequestProcessor := registration.NewV3RequestProcessor(daoMock)
	srv := NewService(daoMock, busMock, entity.NewService("v1"), clusterRegSrv, v3RequestProcessor)

	extAuthz := dto.ExtAuthz{
		Name:        "name",
		Destination: dto.RouteDestination{Cluster: "cluster", Endpoint: "endpoint", TlsEndpoint: "tls-endpoint"},
	}

	ok, _ := srv.ValidateDelete(context.Background(), extAuthz)
	assert.False(t, ok)

	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw1", "gw2")
	assert.False(t, ok)

	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "")
	assert.False(t, ok)

	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Name = ""
	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.False(t, ok)

	extAuthz.Name = "name"
	extAuthz.Destination.Cluster = ""
	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Destination.Cluster = "cluster"
	extAuthz.Destination.Endpoint = ""
	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Destination.TlsEndpoint = ""
	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.True(t, ok)

	extAuthz.Destination.Endpoint = "endpoint"
	ok, _ = srv.ValidateDelete(context.Background(), extAuthz, "gw")
	assert.True(t, ok)
}

func newInMemDao() *dao.InMemDao {
	return dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, []func([]memdb.Change) error{flushChanges})
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

func flushChanges(changes []memdb.Change) error {
	flusher := &constancy.Flusher{BatchTm: &batchTransactionManagerMock{}}
	return flusher.Flush(changes)
}

type batchTransactionManagerMock struct{}

func (tm *batchTransactionManagerMock) WithTxBatch(_ func(tx constancy.BatchStorage) error) error {
	return nil
}

func (tm *batchTransactionManagerMock) IsCurrentPodDefinedAsMaster() (bool, error) {
	return true, nil
}
