package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_FindDeploymentVersions(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	dVersions := []*domain.DeploymentVersion{
		{
			Version: "v1",
			Stage:   "LEGACY",
		},
		{
			Version: "v2",
			Stage:   "ACTIVE",
		},
		{
			Version: "v3",
			Stage:   "CANDIDATE",
		},
		{
			Version: "v4",
			Stage:   "CANDIDATE",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, deploymentVersion := range dVersions {
			assert.Nil(t, dao.SaveDeploymentVersion(deploymentVersion))
		}
		return nil
	})
	assert.Nil(t, err)

	foundDVersions, err := testable.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(foundDVersions))

	foundDVersions, err = testable.FindDeploymentVersionsByStage("CANDIDATE")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundDVersions))

	foundDVersion, err := testable.FindDeploymentVersion("v2")
	assert.Nil(t, err)
	assert.Equal(t, dVersions[1], foundDVersion)
}
