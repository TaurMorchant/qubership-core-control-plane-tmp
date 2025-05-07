package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveClusters_shouldSaveClusters(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	clusters := generateClusters(t, testable)

	foundClusters, err := testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters), len(foundClusters))
}

func TestDeleteClusterByName_shouldDeleteCluster(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	clusters := generateClusters(t, testable)
	cluster := clusters[0]

	foundClusters, err := testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters), len(foundClusters))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteClusterByName(cluster.Name))
		return nil
	})
	assert.Nil(t, err)

	foundClusters, err = testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters)-1, len(foundClusters))

	clusterShouldBeDeleted(t, foundClusters, cluster)
}

func TestDeleteCluster_shouldDeleteCluster(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	clusters := generateClusters(t, testable)
	cluster := clusters[0]

	foundClusters, err := testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters), len(foundClusters))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteCluster(cluster))
		return nil
	})
	assert.Nil(t, err)

	foundClusters, err = testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters)-1, len(foundClusters))

	clusterShouldBeDeleted(t, foundClusters, cluster)
}

func TestDeleteClusterById_shouldDeleteCluster(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	clusters := generateClusters(t, testable)
	cluster := clusters[0]

	foundClusters, err := testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters), len(foundClusters))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteCluster(cluster))
		return nil
	})
	assert.Nil(t, err)

	foundClusters, err = testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, len(clusters)-1, len(foundClusters))

	clusterShouldBeDeleted(t, foundClusters, cluster)
}

func clusterShouldBeDeleted(t *testing.T, foundClusters []*domain.Cluster, deletedCluster *domain.Cluster) {
	found := false
	for _, cluster := range foundClusters {
		if deletedCluster == cluster {
			found = true
			break
		}

	}
	assert.False(t, found)
}

func generateClusters(t *testing.T, testable *InMemRepo) []*domain.Cluster {
	clusters := []*domain.Cluster{
		{
			Id:   1,
			Name: "cluster1",
		},
		{
			Id:   2,
			Name: "cluster2",
		},
		{
			Id:   3,
			Name: "cluster3",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, cluster := range clusters {
			assert.Nil(t, dao.SaveCluster(cluster))
		}
		return nil
	})
	assert.Nil(t, err)

	return clusters
}

func TestInMemRepo_FindClusterByUniqFields(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	clusters := []*domain.Cluster{
		{
			Id:   1,
			Name: "cluster1",
		},
		{
			Id:   2,
			Name: "cluster2",
		},
		{
			Id:   3,
			Name: "cluster3",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, cluster := range clusters {
			assert.Nil(t, dao.SaveCluster(cluster))
		}
		return nil
	})
	assert.Nil(t, err)

	foundClusters, err := testable.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(foundClusters))

	foundCluster, err := testable.FindClusterById(1)
	assert.Nil(t, err)
	assert.Equal(t, clusters[0], foundCluster)

	foundCluster, err = testable.FindClusterByName("cluster2")
	assert.Nil(t, err)
	assert.Equal(t, clusters[1], foundCluster)
}

func TestInMemRepo_FindClusterByNodeGroup(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	nodeGroups := []*domain.NodeGroup{
		{
			Name: "nodeGroup1",
		},
		{
			Name: "nodeGroup2",
		},
	}
	clusters := []*domain.Cluster{
		{
			Id:   1,
			Name: "cluster1",
		},
		{
			Id:   2,
			Name: "cluster2",
		},
		{
			Id:   3,
			Name: "cluster3",
		},
	}
	relations := []*domain.ClustersNodeGroup{
		{
			ClustersId:     1,
			NodegroupsName: "nodeGroup1",
		},
		{
			ClustersId:     2,
			NodegroupsName: "nodeGroup1",
		},
		{
			ClustersId:     3,
			NodegroupsName: "nodeGroup2",
		},
	}

	_, err := testable.WithWTx(func(dao Repository) error {
		for _, cluster := range clusters {
			assert.Nil(t, dao.SaveCluster(cluster))
		}
		for _, nodeGroup := range nodeGroups {
			assert.Nil(t, dao.SaveNodeGroup(nodeGroup))
		}
		for _, relation := range relations {
			assert.Nil(t, dao.SaveClustersNodeGroup(relation))
		}
		return nil
	})
	assert.Nil(t, err)

	foundClusters, err := testable.FindClusterByNodeGroup(nodeGroups[0])
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundClusters))
}

func TestInMemRepo_FindClusterByEndpointIn(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	endpoints := []*domain.Endpoint{
		{
			Id:                1,
			ClusterId:         1,
			DeploymentVersion: "dv0",
		},
		{
			Id:                2,
			ClusterId:         1,
			DeploymentVersion: "dv0",
		},
		{
			Id:                3,
			ClusterId:         2,
			DeploymentVersion: "dv0",
		},
	}

	clusters := []*domain.Cluster{
		{
			Id:   1,
			Name: "cluster1",
		},
		{
			Id:   2,
			Name: "cluster2",
		},
		{
			Id:   3,
			Name: "cluster3",
		},
	}

	_, err := testable.WithWTx(func(repo Repository) error {
		for _, cluster := range clusters {
			assert.Nil(t, repo.SaveCluster(cluster))
		}
		for _, endpoint := range endpoints {
			assert.Nil(t, repo.SaveEndpoint(endpoint))
		}
		return nil
	})
	assert.Nil(t, err)

	firstResult, err := testable.FindClusterByEndpointIn([]*domain.Endpoint{endpoints[0], endpoints[1]})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(firstResult))

	secondResult, err := testable.FindClusterByEndpointIn([]*domain.Endpoint{endpoints[2]})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(secondResult))

	thirdResult, err := testable.FindClusterByEndpointIn([]*domain.Endpoint{{ClusterId: 4}})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(thirdResult))
}
