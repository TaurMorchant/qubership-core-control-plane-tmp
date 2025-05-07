package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (d *InMemRepo) FindRetryPolicyByRouteId(routeId int32) (*domain.RetryPolicy, error) {
	return FindFirstByIndex[domain.RetryPolicy](d, domain.RetryPolicyTable, "routeId", routeId)
}

func (d *InMemRepo) SaveRetryPolicy(retryPolicy *domain.RetryPolicy) error {
	return d.SaveUnique(domain.RetryPolicyTable, retryPolicy)
}

func (d *InMemRepo) DeleteRetryPolicyByRouteId(routeId int32) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	_, err := txCtx.tx.DeleteAll(domain.RetryPolicyTable, "routeId", routeId)
	return err
}

func (d *InMemRepo) DeleteRetryPolicyById(routeId int32) error {
	return d.DeleteById(domain.RetryPolicyTable, routeId)
}
