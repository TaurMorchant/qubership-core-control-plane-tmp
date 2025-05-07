package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestDefaultAndActiveDeploymentVersion(t *testing.T) {
	entityService, mockDao := getService(t)
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		expectedDeploymentVersion := domain.NewDeploymentVersion("v1", domain.ActiveStage)
		selectedVersion, err := entityService.GetActiveDeploymentVersion(dao)
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)

		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(dao, "v1")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)
		return nil
	})
	assert.Nil(t, err)
}

func TestNewCandidateDeploymentVersion(t *testing.T) {
	runTestInWTx(t, func(entityService *Service, mockDao dao.Repository) {
		expectedDeploymentVersion := domain.NewDeploymentVersion("v2", domain.CandidateStage)
		selectedVersion, err := entityService.GetOrCreateDeploymentVersion(mockDao, "v2")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)

		expectedDeploymentVersion = domain.NewDeploymentVersion("v1", domain.ActiveStage)
		selectedVersion, err = entityService.GetActiveDeploymentVersion(mockDao)
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)
	})
}

func TestService_GetOrCreateDeploymentVersion_WithEmptyVersion(t *testing.T) {
	runTestInWTx(t, func(entityService *Service, mockDao dao.Repository) {
		expectedDeploymentVersion := domain.NewDeploymentVersion("v1", domain.ActiveStage)
		selectedVersion, err := entityService.GetOrCreateDeploymentVersion(mockDao, expectedDeploymentVersion.Version)
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)

		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(mockDao, "")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, expectedDeploymentVersion.Version, selectedVersion.Version)
		assert.Equal(t, expectedDeploymentVersion.Stage, selectedVersion.Stage)
	})
}

func TestService_DifferentActiveAndDefaultVersions(t *testing.T) {
	runTestInWTx(t, func(entityService *Service, mockDao dao.Repository) {
		v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
		selectedVersion, err := entityService.GetOrCreateDeploymentVersion(mockDao, v1.Version)
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, v1.Version, selectedVersion.Version)
		assert.Equal(t, v1.Stage, selectedVersion.Stage)

		v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(mockDao, v2.Version)
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, v2.Version, selectedVersion.Version)
		assert.Equal(t, v2.Stage, selectedVersion.Stage)

		// verify that default = v1 and active = v1
		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(mockDao, "")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, v1.Version, entityService.GetDefaultVersion())
		assert.Equal(t, v1.Version, selectedVersion.Version)
		assert.Equal(t, v1.Stage, selectedVersion.Stage)

		// verify that there are no duplicates of ACTIVE version
		activeVersions, err := mockDao.FindDeploymentVersionsByStage(domain.ActiveStage)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(activeVersions))

		// "promote" v2
		// create new objects to avoid affecting in-memory indexes
		v1 = domain.NewDeploymentVersion("v1", domain.LegacyStage)
		v2 = domain.NewDeploymentVersion("v2", domain.ActiveStage)
		err = entityService.SaveDeploymentVersion(mockDao, v1)
		assert.Nil(t, err)
		err = entityService.SaveDeploymentVersion(mockDao, v2)
		assert.Nil(t, err)

		// verify that default = v1 and active = v2
		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(mockDao, "")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, v1.Version, entityService.GetDefaultVersion())
		assert.Equal(t, v2.Version, selectedVersion.Version)
		assert.Equal(t, v2.Stage, selectedVersion.Stage)
		// verify that there are no duplicates of ACTIVE version
		activeVersions, err = mockDao.FindDeploymentVersionsByStage(domain.ActiveStage)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(activeVersions))

		// delete v1
		err = mockDao.DeleteDeploymentVersion(v1)
		assert.Nil(t, err)

		// verify that default = v1 and active = v2
		selectedVersion, err = entityService.GetOrCreateDeploymentVersion(mockDao, "")
		assert.Nil(t, err)
		assert.NotNil(t, selectedVersion)
		assert.Equal(t, v1.Version, entityService.GetDefaultVersion())
		assert.Equal(t, v2.Version, selectedVersion.Version)
		assert.Equal(t, v2.Stage, selectedVersion.Stage)
		// verify that there are no duplicates of ACTIVE version
		activeVersions, err = mockDao.FindDeploymentVersionsByStage(domain.ActiveStage)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(activeVersions))
	})
}

func TestService_SaveDeploymentVersion(t *testing.T) {
	entityService, mockDao := getService(t)
	expectedDeploymentVersion := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		return entityService.SaveDeploymentVersion(dao, expectedDeploymentVersion)
	})
	assert.Nil(t, err)

	dVs, err := mockDao.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, dVs)
	assert.Equal(t, 1, len(dVs))
	assert.Contains(t, dVs, expectedDeploymentVersion)
}

func TestService_SaveDeploymentVersionWithNewStage(t *testing.T) {
	entityService, mockDao := getService(t)
	expectedDeploymentVersion := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		return entityService.SaveDeploymentVersion(dao, expectedDeploymentVersion)
	})
	assert.Nil(t, err)

	expectedDeploymentVersion.Stage = domain.LegacyStage
	_, err = mockDao.WithWTx(func(dao dao.Repository) error {
		return entityService.SaveDeploymentVersion(dao, expectedDeploymentVersion)
	})
	assert.Nil(t, err)

	dVs, err := mockDao.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, dVs)
	assert.Equal(t, 1, len(dVs))
	assert.Contains(t, dVs, expectedDeploymentVersion)

	dV, err := mockDao.FindDeploymentVersion(expectedDeploymentVersion.Version)
	assert.Nil(t, err)
	assert.NotNil(t, dV)
	assert.Equal(t, expectedDeploymentVersion.Version, dV.Version)
	assert.Equal(t, expectedDeploymentVersion.Stage, dV.Stage)

	dVs, err = mockDao.FindDeploymentVersionsByStage(expectedDeploymentVersion.Stage)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVs)
	assert.Equal(t, 1, len(dVs))
	assert.Equal(t, expectedDeploymentVersion.Version, dV.Version)
	assert.Equal(t, expectedDeploymentVersion.Stage, dV.Stage)
}

func runTestInWTx(t *testing.T, testFunc func(entityService *Service, mockDao dao.Repository)) {
	entityService, testDao := getService(t)

	_, err := testDao.WithWTx(func(dao dao.Repository) error {
		return entityService.SaveDeploymentVersion(dao, domain.NewDeploymentVersion("v1", domain.ActiveStage))
	})
	assert.Nil(t, err)

	_, _ = testDao.WithWTx(func(mockDao dao.Repository) error {
		testFunc(entityService, mockDao)
		return nil
	})
}

func getService(t *testing.T) (*Service, *dao.InMemDao) {
	mockDao := dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
	v1 := &domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		return dao.SaveDeploymentVersion(v1)
	})
	assert.Nil(t, err)
	entityService := NewService("v1")
	return entityService, mockDao
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
