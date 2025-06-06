package cleanup

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
	"time"
)

func TestCleanupRoutes_DeletesLocalDevRoutes(t *testing.T) {
	localDevNamespaceFirst := "192.168.0.1" + localDevNamespacePostfix
	localDevNamespaceSecond := "192.168.0.2" + localDevNamespacePostfix

	routeFirst := constructRoute(1, localDevNamespaceFirst, namespaceHeader)
	routeSecond := constructRoute(2, localDevNamespaceSecond, namespaceHeader)

	entityService := entity.NewService("v1")
	callback := func([]memdb.Change) error {
		return nil
	}
	inMemoryDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, []func([]memdb.Change) error{callback})
	_, _ = inMemoryDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}))
		assert.Nil(t, entityService.PutRoutes(dao, []*domain.Route{routeFirst, routeSecond}))
		return nil
	})

	routesFromDB, err := inMemoryDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.Len(t, routesFromDB, 2)

	var cleanupService RoutesCleanupService = localDevCleanupService{
		dao:           inMemoryDao,
		entityService: entityService,
	}

	err = cleanupService.CleanupRoutes()
	assert.Nil(t, err)

	routesFromDB, err = inMemoryDao.FindAllRoutes()
	assert.Len(t, routesFromDB, 0)
}

func TestCleanupRoutes_DeletesOnlyLocalDevRoutes(t *testing.T) {
	localDevNamespace := "192.168.0.1" + localDevNamespacePostfix

	localDevRoute := constructRoute(1, localDevNamespace, namespaceHeader)
	routeWithoutLocalDevNamespace := constructRoute(2, "my-namespace", namespaceHeader)
	routeWithLocalDevNamespace := constructRoute(3, localDevNamespace, "my-header")

	routes := []*domain.Route{
		localDevRoute, routeWithoutLocalDevNamespace, routeWithLocalDevNamespace,
	}

	entityService := entity.NewService("v1")
	callback := func([]memdb.Change) error {
		return nil
	}
	inMemoryDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, []func([]memdb.Change) error{callback})
	_, _ = inMemoryDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveDeploymentVersion(&domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}))
		assert.Nil(t, entityService.PutRoutes(dao, routes))
		return nil
	})

	routesFromDB, err := inMemoryDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.Len(t, routesFromDB, 3)

	var cleanupService RoutesCleanupService = localDevCleanupService{dao: inMemoryDao}

	err = cleanupService.CleanupRoutes()
	assert.Nil(t, err)

	routesFromDB, err = inMemoryDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.Len(t, routesFromDB, 2)
}

func TestCleanupRoutes_DeletesLocalDevClusters(t *testing.T) {
	cluster := domain.Cluster{
		Name: "test-microservice||test-microservice.nip.io||test-microservice:8080",
	}
	callback := func([]memdb.Change) error {
		return nil
	}
	inMemoryDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, []func([]memdb.Change) error{callback})
	_, err := inMemoryDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveCluster(&cluster))
		return nil
	})
	assert.Nil(t, err)

	clustersFromDB, err := inMemoryDao.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clustersFromDB))
	assert.Equal(t, cluster, *clustersFromDB[0])

	var cleanupService RoutesCleanupService = localDevCleanupService{dao: inMemoryDao}

	err = cleanupService.CleanupRoutes()
	assert.Nil(t, err)

	clustersFromDB, err = inMemoryDao.FindAllClusters()
	assert.Nil(t, err)
	assert.Empty(t, clustersFromDB)
}

func Test_RoutesCleanupWorkerImpl_computeCleanupStartupTime(t *testing.T) {
	nowTime := time.Now()
	cleanupWorker := routesCleanupWorkerImpl{
		stop:           nil,
		cleanupService: nil,
		nowFunc:        nil,
	}

	for hour := 0; hour < 24; hour++ {
		nowFunc := func() time.Time {
			return time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), hour, nowTime.Minute(),
				nowTime.Second(), nowTime.Nanosecond(), time.Local)
		}
		cleanupWorker.nowFunc = nowFunc
		var expectedTime time.Time
		if hour < 2 {
			expectedTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day(), 2, 0, 0, 0, time.Local)
		} else {
			expectedTime = time.Date(nowTime.Year(), nowTime.Month(), nowTime.Day()+1, 2, 0, 0, 0, time.Local)
		}
		cleanupTime := cleanupWorker.computeCleanupStartupTime()
		assert.Equal(t, expectedTime, cleanupTime)
	}
}

func constructRoute(routeId int32, headerMatchExact, headerName string) *domain.Route {
	return &domain.Route{
		Id:   routeId,
		Uuid: uuid.New().String(),
		HeaderMatchers: []*domain.HeaderMatcher{
			{ExactMatch: headerMatchExact, Name: headerName, RouteId: routeId},
		},
		Autogenerated:            true,
		VirtualHostId:            1,
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
		RouteKey:                 "1",
	}
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
