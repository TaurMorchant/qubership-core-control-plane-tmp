package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveCircuitBreakers_shouldSaveCircuitBreakers(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	circuitBreakers := generateCircuitBreakers(t, testable)

	findAllCircuitBreakers, err := testable.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Equal(t, len(circuitBreakers), len(findAllCircuitBreakers))
}

func TestInMemRepo_FindCircuitBreaker(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	circuitBreakers := []*domain.CircuitBreaker{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, circuitBreaker := range circuitBreakers {
			assert.Nil(t, dao.SaveCircuitBreaker(circuitBreaker))
		}
		return nil
	})
	assert.Nil(t, err)

	allCircuitBreakers, err := testable.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allCircuitBreakers))

	circuitBreakerById, err := testable.FindCircuitBreakerById(1)
	assert.Nil(t, err)
	assert.Equal(t, circuitBreakers[0], circuitBreakerById)
}

func TestDeleteCircuitBreakerById_shouldDeleteCircuitBreaker(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	circuitBreakers := generateCircuitBreakers(t, testable)
	circuitBreaker := circuitBreakers[0]

	findAllCircuitBreakers, err := testable.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Equal(t, len(circuitBreakers), len(findAllCircuitBreakers))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteCircuitBreakerById(circuitBreaker.Id))
		return nil
	})
	assert.Nil(t, err)

	findAllCircuitBreakers, err = testable.FindAllCircuitBreakers()
	assert.Nil(t, err)
	assert.Equal(t, len(circuitBreakers)-1, len(findAllCircuitBreakers))

	circuitBreakerShouldBeDeleted(t, findAllCircuitBreakers, circuitBreaker)
}

func circuitBreakerShouldBeDeleted(t *testing.T, foundCircuitBreakers []*domain.CircuitBreaker, deletedCircuitBreaker *domain.CircuitBreaker) {
	found := false
	for _, circuitBreaker := range foundCircuitBreakers {
		if deletedCircuitBreaker == circuitBreaker {
			found = true
			break
		}

	}
	assert.False(t, found)
}

func generateCircuitBreakers(t *testing.T, testable *InMemRepo) []*domain.CircuitBreaker {
	circuitBreakers := []*domain.CircuitBreaker{
		{
			Id: 1,
		},
		{
			Id: 2,
		},
		{
			Id: 3,
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, circuitBreaker := range circuitBreakers {
			assert.Nil(t, dao.SaveCircuitBreaker(circuitBreaker))
		}
		return nil
	})
	assert.Nil(t, err)

	return circuitBreakers
}
