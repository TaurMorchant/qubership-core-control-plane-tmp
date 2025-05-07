package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_DeleteCircuitBreakerCascadeById(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedCircuitBreaker := prepareCircuitBreaker(t, inMemDao)

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.DeleteCircuitBreakerCascadeById(dao, expectedCircuitBreaker.Id)
	})
	assert.Nil(t, err)

	actualCircuitBreakers, err := inMemDao.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Empty(t, actualCircuitBreakers)

	actualThreshold, err := inMemDao.FindThresholdById(expectedCircuitBreaker.ThresholdId)
	assert.Nil(t, err)
	assert.Empty(t, actualThreshold)
}

func prepareCircuitBreaker(t *testing.T, memDao *dao.InMemDao) *domain.CircuitBreaker {
	circuitBreaker := domain.CircuitBreaker{Threshold: &domain.Threshold{MaxConnections: 2}}
	saveCircuitBreaker(t, memDao, &circuitBreaker)
	return &circuitBreaker
}

func saveCircuitBreaker(t *testing.T, memDao *dao.InMemDao, circuitBreaker *domain.CircuitBreaker) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveThreshold(circuitBreaker.Threshold))
		circuitBreaker.ThresholdId = circuitBreaker.Threshold.Id
		assert.Nil(t, dao.SaveCircuitBreaker(circuitBreaker))
		return nil
	})
	assert.Nil(t, err)
}
