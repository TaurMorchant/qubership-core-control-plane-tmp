package loadbalance

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestLoadBalanceService_ApplyLoadBalanceConfigNoSuchCluster(t *testing.T) {
	lbService, _ := getLBService()
	ctx := context.Background()
	err := lbService.ApplyLoadBalanceConfig(ctx, "test-cluster", "v1", nil)
	assert.NotNil(t, err)
}

func TestLoadBalanceService_ApplyLoadBalanceConfigUpdateWithEmptyPolicies(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	clusterName := "test-cluster"
	prepareSingleCluster(t, clusterName, inMemDao)
	err := lbService.ApplyLoadBalanceConfig(ctx, clusterName, "v1", make([]*domain.HashPolicy, 0))
	assert.Nil(t, err)
	actualClusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.Equal(t, domain.LbPolicyLeastRequest, actualClusters[0].LbPolicy)
}

func TestLoadBalanceService_ApplyLoadBalanceConfigUpdateWithEmptyPoliciesButKeepLb(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	clusterName := "test-cluster"
	prepareSingleClusterWithHashPolicy(t, clusterName, inMemDao)
	inMemDao.WithWTx(func(dao dao.Repository) error {
		endpoints, err := dao.FindEndpointsByDeploymentVersion("v2")
		assert.Nil(t, err)
		assert.NotEmpty(t, endpoints)
		hashPolicy := &domain.HashPolicy{HeaderName: "namespace", EndpointId: endpoints[0].Id}
		assert.Nil(t, dao.SaveHashPolicy(hashPolicy))
		return nil
	})
	err := lbService.ApplyLoadBalanceConfig(ctx, clusterName, "v1", make([]*domain.HashPolicy, 0))
	assert.Nil(t, err)
	actualClusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.Equal(t, domain.LbPolicyRingHash, actualClusters[0].LbPolicy)
}

func TestLoadBalanceService_ApplyLoadBalanceConfigUpdateHashPolicy(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	clusterName := "test-cluster"
	deployVersionV1 := "v1"
	entityService := entity.NewService(deployVersionV1)
	hashPolicies := []*domain.HashPolicy{
		{
			HeaderName: "BID",
		}, {
			HeaderName: "X-BID",
		},
	}
	prepareSingleCluster(t, clusterName, inMemDao)
	err := lbService.ApplyLoadBalanceConfig(ctx, clusterName, deployVersionV1, hashPolicies)
	assert.Nil(t, err)

	actualClusters, err := entityService.GetClustersWithRelations(inMemDao)
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.Equal(t, domain.LbPolicyRingHash, actualClusters[0].LbPolicy)
	for _, endpoint := range actualClusters[0].Endpoints {
		if endpoint.DeploymentVersion == deployVersionV1 {
			assert.NotEmpty(t, endpoint.HashPolicies)
			assert.Equal(t, 2, len(endpoint.HashPolicies))
		} else {
			assert.Empty(t, endpoint.HashPolicies)
		}
	}
}

func TestLoadBalanceService_ApplyLoadBalanceConfigConfigureLoadBalanceForAllClusters(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	clusterName1 := "test-cluster-1"
	clusterName2 := "test-cluster-2"
	prepareTwoClustersOneWithHashPolicy(t, clusterName1, clusterName2, inMemDao)
	err := lbService.ConfigureLoadBalanceForAllClusters(ctx)
	assert.Nil(t, err)

	actualCluster1, err := inMemDao.FindClusterByName(clusterName1)
	assert.Nil(t, err)
	assert.NotNil(t, actualCluster1)
	assert.Equal(t, domain.LbPolicyLeastRequest, actualCluster1.LbPolicy)
	actualCluster2, err := inMemDao.FindClusterByName(clusterName2)
	assert.Nil(t, err)
	assert.NotNil(t, actualCluster2)
	assert.Equal(t, domain.LbPolicyRingHash, actualCluster2.LbPolicy)
}

func TestLoadBalanceService_ApplyLoadBalanceConfigConfigureLoadBalanceForCluster(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	cluster := domain.NewCluster("test-cluster", false)
	nodeGroup := &domain.NodeGroup{Name: "test-nodegroup"}
	cluster.NodeGroups = []*domain.NodeGroup{nodeGroup}
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	cluster.LbPolicy = domain.LbPolicyLeastRequest
	endpoint := domain.NewEndpoint("endpoint-1", 8080, "v1", "v1", cluster.Id)
	endpoint.DeploymentVersionVal = v1
	hashPolicy := &domain.HashPolicy{
		HeaderName: "header",
		EndpointId: endpoint.Id,
	}
	endpoint.HashPolicies = []*domain.HashPolicy{hashPolicy}
	cluster.Endpoints = []*domain.Endpoint{endpoint}
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, lbService.configureLoadBalanceForCluster(ctx, dao, cluster))
		return nil
	})
	assert.Nil(t, err)

	actualClusterRingHash, err := inMemDao.FindClusterByName("test-cluster")
	assert.Nil(t, err)
	assert.NotNil(t, actualClusterRingHash)
	assert.Equal(t, domain.LbPolicyRingHash, actualClusterRingHash.LbPolicy)

	endpoint.HashPolicies = nil
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, lbService.configureLoadBalanceForCluster(ctx, dao, cluster))
		return nil
	})
	assert.Nil(t, err)
	actualClusterLeastRequest, err := inMemDao.FindClusterByName("test-cluster")
	assert.Nil(t, err)
	assert.NotNil(t, actualClusterLeastRequest)
	assert.Equal(t, domain.LbPolicyLeastRequest, actualClusterLeastRequest.LbPolicy)
}

func TestLoadBalanceService_ApplyLoadBalanceConfigConfigureLoadBalanceForClusterWithFlag(t *testing.T) {
	lbService, inMemDao := getLBService()
	ctx := context.Background()
	nodeGroup := &domain.NodeGroup{Name: "test-nodegroup"}
	cluster := domain.NewCluster("test-cluster", false)
	cluster.NodeGroups = []*domain.NodeGroup{nodeGroup}
	cluster.LbPolicy = domain.LbPolicyLeastRequest
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, lbService.applyLoadBalanceForCluster(ctx, true, dao, cluster))
		return nil
	})
	assert.Nil(t, err)

	actualClusterRingHash, err := inMemDao.FindClusterByName("test-cluster")
	assert.Nil(t, err)
	assert.NotNil(t, actualClusterRingHash)
	assert.Equal(t, domain.LbPolicyRingHash, actualClusterRingHash.LbPolicy)

	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, lbService.applyLoadBalanceForCluster(ctx, false, dao, cluster))
		return nil
	})
	assert.Nil(t, err)
	actualClusterLeastRequest, err := inMemDao.FindClusterByName("test-cluster")
	assert.Nil(t, err)
	assert.NotNil(t, actualClusterLeastRequest)
	assert.Equal(t, domain.LbPolicyLeastRequest, actualClusterLeastRequest.LbPolicy)
}

func getLBService() (*LoadBalanceService, *dao.InMemDao) {
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, nil)
	entityService := entity.NewService("v1")
	return NewLoadBalanceService(genericDao, entityService, eventBus), genericDao
}

func prepareSingleCluster(t *testing.T, clusterName string, memDao *dao.InMemDao) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		prepareClusterWithEndpoints(t, clusterName, dao)
		return nil
	})
	assert.Nil(t, err)
}

func prepareSingleClusterWithHashPolicy(t *testing.T, clusterName string, memDao *dao.InMemDao) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		endpointId := prepareClusterWithEndpoints(t, clusterName, dao)
		addHashPolicy(t, endpointId, dao)
		return nil
	})
	assert.Nil(t, err)
}

func prepareTwoClustersOneWithHashPolicy(t *testing.T, clusterName1, clusterName2 string, memDao *dao.InMemDao) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		prepareClusterWithEndpoints(t, clusterName1, dao)
		endpointId := prepareClusterWithEndpoints(t, clusterName2, dao)
		addHashPolicy(t, endpointId, dao)
		return nil
	})
	assert.Nil(t, err)
}

func addHashPolicy(t *testing.T, id int32, dao dao.Repository) {
	hashPolicy1 := &domain.HashPolicy{
		HeaderName: "header",
		EndpointId: id,
	}
	assert.Nil(t, dao.SaveHashPolicy(hashPolicy1))
}

func prepareClusterWithEndpoints(t *testing.T, clusterName string, dao dao.Repository) int32 {
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	assert.Nil(t, dao.SaveDeploymentVersion(v1))
	assert.Nil(t, dao.SaveDeploymentVersion(v2))
	cluster := domain.NewCluster(clusterName, false)
	assert.Nil(t, dao.SaveCluster(cluster))
	nodeGroup := &domain.NodeGroup{Name: "test-nodegroup"}
	assert.Nil(t, dao.SaveNodeGroup(nodeGroup))
	clusterNodeGroup := domain.NewClusterNodeGroups(cluster.Id, nodeGroup.Name)
	assert.Nil(t, dao.SaveClustersNodeGroup(clusterNodeGroup))
	endpoint1 := domain.NewEndpoint("endpoint-1", 8080, "v1", "v1", cluster.Id)
	endpoint2 := domain.NewEndpoint("endpoint-2", 8080, "v1", "v1", cluster.Id)
	endpoint3 := domain.NewEndpoint("endpoint-3", 8080, "v2", "v2", cluster.Id)
	assert.Nil(t, dao.SaveEndpoint(endpoint1))
	assert.Nil(t, dao.SaveEndpoint(endpoint2))
	assert.Nil(t, dao.SaveEndpoint(endpoint3))
	return endpoint1.Id
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
