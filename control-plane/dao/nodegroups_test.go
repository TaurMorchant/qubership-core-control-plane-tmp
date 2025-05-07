package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteNodeGroupByName_shouldDeleteNodeGroup(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	nodeGroups, _, _ := createClusters(t, testable)
	assert.NotNil(t, nodeGroups)
	groupName := "nodeGroup1"
	found := false
	for _, nodeGroup := range nodeGroups {
		if nodeGroup.Name == groupName {
			found = true
		}
	}
	assert.True(t, found)

	_, err := testable.WithWTx(func(dao Repository) error {
		err := dao.DeleteNodeGroupByName(groupName)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	foundNodeGroups, err := testable.FindAllNodeGroups()
	assert.Nil(t, err)
	found = false
	for _, foundNodeGroup := range foundNodeGroups {
		if foundNodeGroup.Name == groupName {
			found = true
		}
	}
	assert.False(t, found)
}

func TestDeleteClustersNodeGroupByClusterId_shouldDeleteRelation(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	_, _, relations := createClusters(t, testable)
	assert.NotNil(t, relations)
	clusterId := int32(3)
	found := false
	for _, relation := range relations {
		if relation.ClustersId == clusterId {
			found = true
		}
	}
	assert.True(t, found)

	_, err := testable.WithWTx(func(dao Repository) error {
		deletedQuantity, err := dao.DeleteClustersNodeGroupByClusterId(clusterId)
		assert.Nil(t, err)
		assert.Equal(t, 1, deletedQuantity)
		return nil
	})
	assert.Nil(t, err)

	foundRelations, err := testable.FindAllClusterWithNodeGroup()
	assert.Nil(t, err)
	found = false
	for _, foundRelation := range foundRelations {
		if foundRelation.ClustersId == clusterId {
			found = true
		}
	}
	assert.False(t, found)
}

func TestFindAllClusterWithNodeGroup_shouldFoundAllRelations(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	_, _, relations := createClusters(t, testable)

	foundRelations, err := testable.FindAllClusterWithNodeGroup()
	assert.Nil(t, err)
	assert.NotNil(t, foundRelations)
	assert.Equal(t, len(relations), len(foundRelations))
	assert.Equal(t, relations, foundRelations)
}

func TestFindClustersNodeGroup_shouldFoundRelation_whenRelationExist(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	_, _, relations := createClusters(t, testable)

	foundRelation, err := testable.FindClustersNodeGroup(relations[0])
	assert.Nil(t, err)
	found := false
	for _, relation := range relations {
		if relation == foundRelation {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestFindClustersNodeGroup_shouldNotFoundRelation_whenRelationNotExist(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	createClusters(t, testable)

	relation := &domain.ClustersNodeGroup{
		ClustersId:     -1,
		NodegroupsName: "testNodeGroupNotExist",
	}

	foundRelation, err := testable.FindClustersNodeGroup(relation)
	assert.Nil(t, err)
	assert.Nil(t, foundRelation)
}

func TestInMemDao_FindNodeGroups(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	nodeGroups, clusters, relations := createClusters(t, testable)

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

	foundNodeGroups, err := testable.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundNodeGroups))

	foundNodeGroups, err = testable.FindNodeGroupsByCluster(clusters[0])
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundNodeGroups))

	foundNodeGroup, err := testable.FindNodeGroupByName("nodeGroup1")
	assert.Nil(t, err)
	assert.Equal(t, nodeGroups[0], foundNodeGroup)
}

func createClusters(t *testing.T, testable *InMemRepo) ([]*domain.NodeGroup, []*domain.Cluster, []*domain.ClustersNodeGroup) {
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
			ClustersId:     1,
			NodegroupsName: "nodeGroup2",
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

	return nodeGroups, clusters, relations
}
