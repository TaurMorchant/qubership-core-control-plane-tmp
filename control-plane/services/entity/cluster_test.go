package entity

import (
	"github.com/google/uuid"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PutCluster(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedCluster := domain.NewCluster("test-cluster", false)
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutCluster(dao, expectedCluster))
		return nil
	})
	assert.Nil(t, err)

	actualClusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.Contains(t, actualClusters, expectedCluster)

	expectedCluster = domain.NewCluster("test-cluster", false)
	expectedCluster.LbPolicy = domain.LbPolicyRingHash
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutCluster(dao, expectedCluster))
		return nil
	})
	assert.Nil(t, err)

	actualClusters, err = inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.Contains(t, actualClusters, expectedCluster)
}

func TestService_PutClustersNodeGroupIfAbsent(t *testing.T) {
	testNodeGroupName := "test-node-group"
	entityService, inMemDao := getService(t)
	expectedRelation := domain.NewClusterNodeGroups(1, testNodeGroupName)
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutClustersNodeGroupIfAbsent(dao, expectedRelation))
		return nil
	})
	assert.Nil(t, err)

	actualRelations, err := inMemDao.FindAllClusterWithNodeGroup()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRelations)
	assert.Equal(t, 1, len(actualRelations))
	assert.Contains(t, actualRelations, expectedRelation)

	expectedRelation = domain.NewClusterNodeGroups(1, testNodeGroupName)
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutClustersNodeGroupIfAbsent(dao, expectedRelation))
		return nil
	})
	assert.Nil(t, err)

	actualRelations, err = inMemDao.FindAllClusterWithNodeGroup()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRelations)
	assert.Equal(t, 1, len(actualRelations))
	assert.Contains(t, actualRelations, expectedRelation)

	expectedSecondRelation := domain.NewClusterNodeGroups(2, testNodeGroupName)
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutClustersNodeGroupIfAbsent(dao, expectedSecondRelation))
		return nil
	})
	assert.Nil(t, err)

	actualRelations, err = inMemDao.FindAllClusterWithNodeGroup()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRelations)
	assert.Equal(t, 2, len(actualRelations))
	assert.Contains(t, actualRelations, expectedRelation, expectedSecondRelation)
}

func TestService_GetClustersWithRelations(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedCluster := prepareClusterData(t, inMemDao)
	actualClusters, err := entityService.GetClustersWithRelations(inMemDao)
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 1, len(actualClusters))
	assert.NotEmpty(t, actualClusters[0].Endpoints)
	assert.Equal(t, 1, len(actualClusters[0].Endpoints))
	AssertDeepEqual(t, expectedCluster, actualClusters[0], domain.ClusterTable)
}

func TestService_GetClustersWithRelationsWithNoClusters(t *testing.T) {
	entityService, inMemDao := getService(t)
	actualClusters, err := entityService.GetClustersWithRelations(inMemDao)
	assert.Nil(t, err)
	assert.Empty(t, actualClusters)
}

func TestService_GetClusterWithRelations(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedCluster := prepareClusterData(t, inMemDao)
	actualClusters, err := entityService.GetClusterWithRelations(inMemDao, expectedCluster.Name)
	assert.Nil(t, err)
	assert.NotNil(t, actualClusters)
	assert.NotEmpty(t, actualClusters.Endpoints)
	assert.Equal(t, 1, len(actualClusters.Endpoints))
	AssertDeepEqual(t, expectedCluster, actualClusters, domain.ClusterTable)
}

func TestService_GetClusterWithRelationsWithEmptyCluster(t *testing.T) {
	entityService, inMemDao := getService(t)
	actualClusters, err := entityService.GetClusterWithRelations(inMemDao, "")
	assert.Nil(t, err)
	assert.Nil(t, actualClusters)
}

func TestService_DeleteClusterCascade(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedCluster := prepareClusterDataWithRoutes(t, inMemDao)

	actualClusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)

	actualEndpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)

	actualRoutes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)

	actualCircuitBreakers, err := inMemDao.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualCircuitBreakers)

	actualThresholds, err := inMemDao.FindAllThresholds()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualThresholds)

	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		err := entityService.DeleteClusterCascade(dao, expectedCluster)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	actualClusters, err = inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.Empty(t, actualClusters)

	actualEndpoints, err = inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.Empty(t, actualEndpoints)

	actualRoutes, err = inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.Empty(t, actualRoutes)

	actualCircuitBreakers, err = inMemDao.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Empty(t, actualCircuitBreakers)

	actualThresholds, err = inMemDao.FindAllThresholds()
	assert.Nil(t, err)
	assert.Empty(t, actualThresholds)
}

func TestService_DeleteClusterCascadeWhenClusterNotFound(t *testing.T) {
	entityService, inMemDao := getService(t)
	cluster := domain.NewCluster("test-cluster", false)

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		err := entityService.DeleteClusterCascade(dao, cluster)
		assert.NotNil(t, err)
		return nil
	})
	assert.Nil(t, err)
}

func prepareClusterDataWithRoutes(t *testing.T, memDao *dao.InMemDao) *domain.Cluster {
	route1 := &domain.Route{Uuid: uuid.New().String(), RouteKey: "1", VirtualHostId: 1, ClusterName: "test-cluster", DeploymentVersion: "v1"}
	route2 := &domain.Route{Uuid: uuid.New().String(), RouteKey: "1", VirtualHostId: 1, ClusterName: "test-cluster", DeploymentVersion: "v1"}
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveRoute(route1))
		assert.Nil(t, dao.SaveRoute(route2))
		return nil
	})
	assert.Nil(t, err)
	return prepareClusterData(t, memDao)
}

func prepareClusterData(t *testing.T, memDao *dao.InMemDao) *domain.Cluster {
	var httpVersion int32 = 1
	cluster := domain.NewCluster2("test-cluster", &httpVersion)
	endpoint := domain.NewEndpoint("some-address", 8080, "v1", "v1", 0)
	threshold := &domain.Threshold{MaxConnections: 2}
	circuitBreaker := &domain.CircuitBreaker{Threshold: threshold}
	saveCircuitBreakerData(t, memDao, circuitBreaker)
	cluster.CircuitBreakerId = circuitBreaker.Id
	cluster.CircuitBreaker = circuitBreaker
	cluster.Id, endpoint.Id = saveClusterData(t, memDao, *cluster, *endpoint)
	endpoint.ClusterId = cluster.Id
	cluster.Endpoints = []*domain.Endpoint{endpoint}
	cluster.NodeGroups = []*domain.NodeGroup{}
	return cluster
}

func saveClusterData(t *testing.T, memDao *dao.InMemDao, cluster domain.Cluster, endpoint domain.Endpoint) (int32, int32) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveCluster(&cluster))
		endpoint.ClusterId = cluster.Id
		assert.Nil(t, dao.SaveEndpoint(&endpoint))
		return nil
	})
	assert.Nil(t, err)
	return cluster.Id, endpoint.Id
}

func saveCircuitBreakerData(t *testing.T, memDao *dao.InMemDao, circuitBreaker *domain.CircuitBreaker) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveThreshold(circuitBreaker.Threshold))
		circuitBreaker.ThresholdId = circuitBreaker.Threshold.Id
		assert.Nil(t, dao.SaveCircuitBreaker(circuitBreaker))
		return nil
	})
	assert.Nil(t, err)
}

func ClustersEqual(expected, actual *domain.Cluster) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if expected.Id != actual.Id ||
		expected.Name != actual.Name ||
		expected.HttpVersion != actual.HttpVersion ||
		expected.DiscoveryType != actual.DiscoveryType ||
		expected.LbPolicy != actual.LbPolicy {

		return false
	}
	if len(expected.Endpoints) != len(actual.Endpoints) {
		return false
	}
	for _, expectedEndpoint := range expected.Endpoints {
		presentsInBothLists := false
		for _, actualEndpoint := range actual.Endpoints {
			if EndpointsEqual(expectedEndpoint, actualEndpoint) {
				presentsInBothLists = true
				break
			}
		}
		if !presentsInBothLists {
			return false
		}
	}
	for _, actualEndpoint := range actual.Endpoints {
		presentsInBothLists := false
		for _, expectedEndpoint := range expected.Endpoints {
			if EndpointsEqual(expectedEndpoint, actualEndpoint) {
				presentsInBothLists = true
				break
			}
		}
		if !presentsInBothLists {
			return false
		}
	}
	return true
}
