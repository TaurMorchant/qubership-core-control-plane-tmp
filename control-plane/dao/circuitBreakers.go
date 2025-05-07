package dao

import "github.com/netcracker/qubership-core-control-plane/domain"

func (d *InMemRepo) SaveCircuitBreaker(circuitBreaker *domain.CircuitBreaker) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	if err := d.idGenerator.Generate(circuitBreaker); err != nil {
		return err
	}
	err := d.storage.Save(txCtx.tx, domain.CircuitBreakerTable, circuitBreaker)
	if err != nil {
		return err
	}
	return nil
}

func (d *InMemRepo) FindAllCircuitBreakers() ([]*domain.CircuitBreaker, error) {
	return FindAll[domain.CircuitBreaker](d, domain.CircuitBreakerTable)
}

func (d *InMemRepo) FindCircuitBreakerById(id int32) (*domain.CircuitBreaker, error) {
	return FindById[domain.CircuitBreaker](d, domain.CircuitBreakerTable, id)
}

func (d *InMemRepo) DeleteCircuitBreakerById(id int32) error {
	return d.DeleteById(domain.CircuitBreakerTable, id)
}
