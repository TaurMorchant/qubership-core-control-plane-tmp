package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSaveThreshold_shouldSaveThresholds(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	thresholds := generateThresholds(t, testable)

	allThresholds, err := testable.FindAllThresholds()
	assert.Nil(t, err)
	assert.Equal(t, len(thresholds), len(allThresholds))
}

func TestInMemRepo_FindThreshold(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	thresholds := []*domain.Threshold{
		{
			Id:             1,
			MaxConnections: 1,
		},
		{
			Id:             2,
			MaxConnections: 2,
		},
		{
			Id:             3,
			MaxConnections: 3,
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, threshold := range thresholds {
			assert.Nil(t, dao.SaveThreshold(threshold))
		}
		return nil
	})
	assert.Nil(t, err)

	allThresholds, err := testable.FindAllThresholds()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(allThresholds))

	thresholdById, err := testable.FindThresholdById(1)
	assert.Nil(t, err)
	assert.Equal(t, thresholds[0], thresholdById)
}

func TestDeleteThresholdById_shouldDeleteThreshold(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	thresholds := generateThresholds(t, testable)
	threshold := thresholds[0]

	allThresholds, err := testable.FindAllThresholds()
	assert.Nil(t, err)
	assert.Equal(t, len(thresholds), len(allThresholds))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteThresholdById(threshold.Id))
		return nil
	})
	assert.Nil(t, err)

	allThresholds, err = testable.FindAllThresholds()
	assert.Nil(t, err)
	assert.Equal(t, len(thresholds)-1, len(allThresholds))

	thresholdShouldBeDeleted(t, allThresholds, threshold)
}

func thresholdShouldBeDeleted(t *testing.T, foundThresholds []*domain.Threshold, deletedThreshold *domain.Threshold) {
	found := false
	for _, threshold := range foundThresholds {
		if deletedThreshold == threshold {
			found = true
			break
		}

	}
	assert.False(t, found)
}

func generateThresholds(t *testing.T, testable *InMemRepo) []*domain.Threshold {
	thresholds := []*domain.Threshold{
		{
			Id:             1,
			MaxConnections: 1,
		},
		{
			Id:             2,
			MaxConnections: 2,
		},
		{
			Id:             3,
			MaxConnections: 3,
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, threshold := range thresholds {
			assert.Nil(t, dao.SaveThreshold(threshold))
		}
		return nil
	})
	assert.Nil(t, err)

	return thresholds
}
