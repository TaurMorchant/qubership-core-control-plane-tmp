package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PutEndpointWithUpdate(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := domain.NewEndpoint("test-endpoint", 8080, "v1", "v1", 1)
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutEndpoint(dao, expectedEndpoint))
		return nil
	})
	assert.Nil(t, err)

	actualEndpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 1, len(actualEndpoints))
	assert.EqualValues(t, expectedEndpoint, actualEndpoints[0])

	expectedEndpoint = domain.NewEndpoint("test-endpoint-1", 8080, "v1", "v1", 1)
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutEndpoint(dao, expectedEndpoint))
		return nil
	})
	assert.Nil(t, err)

	actualEndpoints, err = inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 1, len(actualEndpoints))
	assert.EqualValues(t, expectedEndpoint, actualEndpoints[0])
}

func TestService_PutEndpointWithoutUpdate(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := domain.NewEndpoint("test-endpoint", 8080, "v1", "v1", 1)
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutEndpoint(dao, expectedEndpoint))
		return nil
	})
	assert.Nil(t, err)

	actualEndpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 1, len(actualEndpoints))
	assert.Contains(t, actualEndpoints, expectedEndpoint)

	expectedSecondEndpoint := domain.NewEndpoint("test-endpoint-1", 8080, "v2", "v1", 1)
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutEndpoint(dao, expectedSecondEndpoint))
		return nil
	})
	assert.Nil(t, err)

	actualEndpoints, err = inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 2, len(actualEndpoints))
	assert.Contains(t, actualEndpoints, expectedSecondEndpoint)
	assert.Contains(t, actualEndpoints, expectedEndpoint)
}

func TestService_FindEndpointsByClusterId(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := prepareEndpointWithRelation(t, inMemDao)

	actualEndpoints, err := entityService.FindEndpointsByClusterId(inMemDao, expectedEndpoint.ClusterId)
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 1, len(actualEndpoints))
	assert.Contains(t, actualEndpoints, expectedEndpoint)
}

func TestService_FindEndpointsByClusterIdWhenNotFound(t *testing.T) {
	entityService, inMemDao := getService(t)

	actualEndpoints, err := entityService.FindEndpointsByClusterId(inMemDao, 1)
	assert.Nil(t, err)
	assert.Empty(t, actualEndpoints)
}

func TestService_LoadEndpointRelationsWithHash(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := prepareEndpointWithRelation(t, inMemDao)

	actualEndpoint, err := entityService.LoadEndpointRelations(inMemDao, expectedEndpoint)
	assert.Nil(t, err)
	assert.NotNil(t, actualEndpoint)
	assert.NotEmpty(t, actualEndpoint.HashPolicies)
	assert.NotEmpty(t, 1, len(actualEndpoint.HashPolicies))
	assert.EqualValues(t, expectedEndpoint, actualEndpoint)
}

func TestService_DeleteEndpointCascade(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := prepareEndpointWithRelation(t, inMemDao)

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		err := entityService.DeleteEndpointCascade(dao, expectedEndpoint)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	actualEndpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.Empty(t, actualEndpoints)

	actualHashPolicies, err := inMemDao.FindHashPolicyByEndpointId(expectedEndpoint.Id)
	assert.Nil(t, err)
	assert.Empty(t, actualHashPolicies)
}

func TestService_DeleteEndpointCascadeWhenEndpointNotFound(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedEndpoint := domain.NewEndpoint("test-endpoint", 8080, "v1", "v1", 1)

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		err := entityService.DeleteEndpointCascade(dao, expectedEndpoint)
		assert.NotNil(t, err)
		return nil
	})
	assert.Nil(t, err)
}

func prepareEndpointWithRelation(t *testing.T, memDao *dao.InMemDao) *domain.Endpoint {
	endpoint := domain.NewEndpoint("test-endpoint", 8080, "v1", "v1", 1)
	hashPolicy := &domain.HashPolicy{HeaderName: "namespace"}
	endpoint.Id, hashPolicy.Id = saveEndpointWithRelation(t, *memDao, *endpoint, *hashPolicy)
	hashPolicy.EndpointId = endpoint.Id
	endpoint.HashPolicies = []*domain.HashPolicy{hashPolicy}
	version, err := memDao.FindDeploymentVersion("v1")
	assert.Nil(t, err)
	endpoint.DeploymentVersionVal = version
	return endpoint
}

func saveEndpointWithRelation(t *testing.T, memDao dao.InMemDao, endpoint domain.Endpoint, hashPolicy domain.HashPolicy) (int32, int32) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveEndpoint(&endpoint))
		hashPolicy.EndpointId = endpoint.Id
		assert.Nil(t, dao.SaveHashPolicy(&hashPolicy))
		return nil
	})
	assert.Nil(t, err)
	return endpoint.Id, hashPolicy.Id
}

func EndpointsEqual(expected, actual *domain.Endpoint) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if expected.Id != actual.Id ||
		expected.DeploymentVersion != actual.DeploymentVersion ||
		expected.ClusterId != actual.ClusterId ||
		expected.Address != actual.Address ||
		expected.Port != actual.Port {
		return false
	}
	if len(expected.HashPolicies) != len(actual.HashPolicies) {
		return false
	}
	for _, expectedPolicy := range expected.HashPolicies {
		presentsInBothEndpoints := false
		for _, actualPolicy := range actual.HashPolicies {
			if expectedPolicy.Equals(actualPolicy) {
				presentsInBothEndpoints = true
				break
			}
		}
		if !presentsInBothEndpoints {
			return false
		}
	}
	for _, actualPolicy := range actual.HashPolicies {
		presentsInBothEndpoints := false
		for _, expectedPolicy := range expected.HashPolicies {
			if expectedPolicy.Equals(actualPolicy) {
				presentsInBothEndpoints = true
				break
			}
		}
		if !presentsInBothEndpoints {
			return false
		}
	}
	return true
}
