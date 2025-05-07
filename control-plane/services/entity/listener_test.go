package entity

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PutListenerIfDoesNotExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedListener := domain.NewListener("test-listener", "test-host", "8080", "test-nodegroup", "test-routeconfig")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutListener(dao, expectedListener))
		return nil
	})
	assert.Nil(t, err)

	actualListeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualListeners)
	assert.Equal(t, 1, len(actualListeners))
	assert.Contains(t, actualListeners, expectedListener)
}

func TestService_PutListenerIfExists(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedListener := domain.NewListener("test-listener", "test-host", "8080", "test-nodegroup", "test-routeconfig")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutListener(dao, expectedListener))
		return nil
	})
	assert.Nil(t, err)

	expectedListener = domain.NewListener("test-listener", "test-host2", "8080", "test-nodegroup", "test-routeconfig")
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutListener(dao, expectedListener))
		return nil
	})
	assert.Nil(t, err)

	actualListeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualListeners)
	assert.Equal(t, 1, len(actualListeners))
	assert.Contains(t, actualListeners, expectedListener)
}

func TestService_PutListenersWhichDoNotExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedFirstListener := domain.NewListener("test-listener1", "test-host1", "8080", "test-nodegroup", "test-routeconfig")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutListener(dao, expectedFirstListener))
		return nil
	})
	assert.Nil(t, err)

	expectedSecondListener := domain.NewListener("test-listener2", "test-host2", "8080", "test-nodegroup", "test-routeconfig")
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutListener(dao, expectedSecondListener))
		return nil
	})
	assert.Nil(t, err)

	actualListeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualListeners)
	assert.Equal(t, 2, len(actualListeners))
	assert.Contains(t, actualListeners, expectedFirstListener)
	assert.Contains(t, actualListeners, expectedSecondListener)
}
