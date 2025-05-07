package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_FindEndpoints(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	endpoints := []*domain.Endpoint{
		{
			Id:                1,
			ClusterId:         1,
			DeploymentVersion: "v1",
		},
		{
			Id:                2,
			ClusterId:         1,
			DeploymentVersion: "v2",
		},
		{
			Id:                3,
			ClusterId:         2,
			DeploymentVersion: "v1",
		},
		{
			Id:                4,
			ClusterId:         2,
			DeploymentVersion: "v2",
		},
		{
			Id:                5,
			ClusterId:         2,
			DeploymentVersion: "v3",
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
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, endpoint := range endpoints {
			assert.Nil(t, dao.SaveEndpoint(endpoint))
		}
		for _, cluster := range clusters {
			assert.Nil(t, dao.SaveCluster(cluster))
		}
		return nil
	})
	assert.Nil(t, err)

	foundEndpoint, err := testable.FindEndpointById(3)
	assert.Nil(t, err)
	assert.Equal(t, endpoints[2], foundEndpoint)

	foundEndpoints, err := testable.FindEndpointsByClusterId(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundEndpoints))

	foundEndpoints, err = testable.FindEndpointsByClusterName("cluster2")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(foundEndpoints))

	foundEndpoints, err = testable.FindEndpointsByDeploymentVersion("v2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundEndpoints))
}
