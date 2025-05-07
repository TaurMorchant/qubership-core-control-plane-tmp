package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_RateLimit(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	rateLimits, err := testable.FindAllRateLimits()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rateLimits))

	_, err = testable.WithWTx(func(repo Repository) error {
		err := repo.SaveRateLimit(&domain.RateLimit{
			Name:                   "rl1",
			LimitRequestsPerSecond: 10,
			Priority:               domain.Product,
		})
		assert.Nil(t, err)
		err = repo.SaveRateLimit(&domain.RateLimit{
			Name:                   "rl1",
			LimitRequestsPerSecond: 100,
			Priority:               domain.Project,
		})
		assert.Nil(t, err)
		err = repo.SaveRateLimit(&domain.RateLimit{
			Name:                   "rl2",
			LimitRequestsPerSecond: 200,
			Priority:               domain.Product,
		})
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	rateLimit, err := testable.FindRateLimitByNameWithHighestPriority("rl1")
	assert.Nil(t, err)
	assert.NotNil(t, rateLimit)
	assert.Equal(t, uint32(100), rateLimit.LimitRequestsPerSecond)

	rateLimit, err = testable.FindRateLimitByNameWithHighestPriority("rl2")
	assert.Nil(t, err)
	assert.NotNil(t, rateLimit)
	assert.Equal(t, uint32(200), rateLimit.LimitRequestsPerSecond)

	rateLimits, err = testable.FindAllRateLimits()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(rateLimits))
}
