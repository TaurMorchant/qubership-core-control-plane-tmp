package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (d *InMemRepo) DeleteHealthCheckById(id int32) error {
	return d.DeleteById(domain.HealthCheckTable, id)
}

func (d *InMemRepo) DeleteHealthChecksByClusterId(clusterId int32) (int, error) {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return txCtx.tx.DeleteAll(domain.HealthCheckTable, "clusterId", clusterId)
}

func (d *InMemRepo) FindHealthChecksByClusterId(clusterId int32) ([]*domain.HealthCheck, error) {
	return FindByIndex[domain.HealthCheck](d, domain.HealthCheckTable, "clusterId", clusterId)
}

func (d *InMemRepo) SaveHealthCheck(healthCheck *domain.HealthCheck) error {
	return d.SaveUnique(domain.HealthCheckTable, healthCheck)
}
