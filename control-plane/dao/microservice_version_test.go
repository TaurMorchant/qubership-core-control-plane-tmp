package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_MicroserviceVersion(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	msVersions := []*domain.MicroserviceVersion{
		{
			Name:                     "microservice-1",
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		},
		{
			Name:                     "microservice-2",
			Namespace:                "test-ns",
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		},
		{
			Name:                     "microservice-3",
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v1",
		},
		{
			Name:                     "microservice-1",
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v2",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, msVersion := range msVersions {
			assert.Nil(t, dao.SaveMicroserviceVersion(msVersion))
		}
		return nil
	})
	assert.Nil(t, err)

	allMsVersions, err := testable.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(allMsVersions))

	actualVersion, err := testable.FindMicroserviceVersionByNameAndInitialVersion("microservice-3", msaddr.CurrentNamespace(), "v1")
	assert.Nil(t, err)
	assert.NotNil(t, actualVersion)

	actualVersion, err = testable.FindMicroserviceVersionByNameAndInitialVersion("microservice-2", msaddr.Namespace{}, "v1")
	assert.Nil(t, err)
	assert.Nil(t, actualVersion)

	actualVersion, err = testable.FindMicroserviceVersionByNameAndInitialVersion("microservice-2", msaddr.Namespace{Namespace: "test-ns"}, "v1")
	assert.Nil(t, err)
	assert.NotNil(t, actualVersion)

	actualVersions, err := testable.FindMicroserviceVersionsByVersion(nil)
	assert.NotNil(t, err)

	actualVersions, err = testable.FindMicroserviceVersionsByVersion(&domain.DeploymentVersion{Version: "v1"})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualVersions))

	actualVersions, err = testable.FindMicroserviceVersionsByVersion(&domain.DeploymentVersion{Version: "v2"})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualVersions))

	actualVersions, err = testable.FindMicroserviceVersionsByNameAndNamespace("microservice-1", msaddr.CurrentNamespace())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualVersions))

	actualVersions, err = testable.FindMicroserviceVersionsByNameAndNamespace("microservice-2", msaddr.Namespace{Namespace: "test-ns"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualVersions))

	_, err = testable.WithWTx(func(dao Repository) error {
		for _, msVersion := range allMsVersions {
			assert.Nil(t, dao.DeleteMicroserviceVersion(msVersion.Name, msaddr.Namespace{Namespace: msVersion.Namespace}, msVersion.InitialDeploymentVersion))
		}
		return nil
	})
	assert.Nil(t, err)

	actualVersions, err = testable.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)
}
