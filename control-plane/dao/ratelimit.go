package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (d *InMemRepo) FindRateLimitByNameWithHighestPriority(name string) (*domain.RateLimit, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	rateLimits, err := d.storage.FindByIndex(txCtx.tx, domain.RateLimitTable, "name", name)
	if err != nil {
		return nil, err
	}
	if rateLimits != nil {
		maxPriority := domain.Product - 1
		var result *domain.RateLimit
		for _, rateLimit := range rateLimits.([]*domain.RateLimit) {
			if rateLimit.Priority > maxPriority {
				maxPriority = rateLimit.Priority
				result = rateLimit
			}
		}
		return result, nil
	}
	return nil, nil
}

func (d *InMemRepo) SaveRateLimit(rateLimit *domain.RateLimit) error {
	return d.SaveEntity(domain.RateLimitTable, rateLimit)
}

func (d *InMemRepo) DeleteRateLimitByNameAndPriority(name string, priority domain.ConfigPriority) error {
	return d.DeleteById(domain.RateLimitTable, name, priority)
}

func (d *InMemRepo) FindAllRateLimits() ([]*domain.RateLimit, error) {
	return FindAll[domain.RateLimit](d, domain.RateLimitTable)
}
