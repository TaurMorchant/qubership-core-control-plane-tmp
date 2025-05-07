package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_FindHashPolicyByClusterAndVersions(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	hashPolicies := []*domain.HashPolicy{
		{
			EndpointId: 1,
		},
		{
			EndpointId: 2,
		},
		{
			EndpointId: 1,
		},
		{
			EndpointId: 3,
		},
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
			ClusterId:         1,
			DeploymentVersion: "v3",
		},
	}
	clusters := []*domain.Cluster{
		{
			Id:   1,
			Name: "cluster1",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, hashPolicy := range hashPolicies {
			assert.Nil(t, dao.SaveHashPolicy(hashPolicy))
		}
		for _, endpoint := range endpoints {
			assert.Nil(t, dao.SaveEndpoint(endpoint))
		}
		for _, cluster := range clusters {
			assert.Nil(t, dao.SaveCluster(cluster))
		}
		return nil
	})
	assert.Nil(t, err)

	foundHashPolicies, err := testable.FindHashPolicyByClusterAndVersions("cluster1", "v1", "v2")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(foundHashPolicies))

}
