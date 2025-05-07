package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_GetOrCreateNodeGroupWhichDoesNotExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedNodeGroup := "testNodeGroup"
	nodeGroup, _, err := inMemDao.WithWTxVal(func(dao dao.Repository) (interface{}, error) {
		return entityService.CreateOrUpdateNodeGroup(dao, domain.NodeGroup{Name: expectedNodeGroup})
	})
	assert.Nil(t, err)
	assert.NotNil(t, nodeGroup)
	actualNodeGroup := nodeGroup.(domain.NodeGroup)
	assert.Equal(t, expectedNodeGroup, actualNodeGroup.Name)
}

func TestService_GetOrCreateNodeGroupWhichExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedNodeGroup := &domain.NodeGroup{Name: "testNodeGroup"}

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return dao.SaveNodeGroup(expectedNodeGroup)
	})
	assert.Nil(t, err)

	nodeGroup, _, err := inMemDao.WithWTxVal(func(dao dao.Repository) (interface{}, error) {
		return entityService.CreateOrUpdateNodeGroup(dao, *expectedNodeGroup)
	})

	assert.Nil(t, err)
	assert.NotNil(t, nodeGroup)
	actualNodeGroup := nodeGroup.(domain.NodeGroup)
	assert.Equal(t, expectedNodeGroup.Name, actualNodeGroup.Name)
}
