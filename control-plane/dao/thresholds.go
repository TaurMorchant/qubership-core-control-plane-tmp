package dao

import "github.com/netcracker/qubership-core-control-plane/domain"

func (d *InMemRepo) SaveThreshold(threshold *domain.Threshold) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	if err := d.idGenerator.Generate(threshold); err != nil {
		return err
	}
	err := d.storage.Save(txCtx.tx, domain.ThresholdTable, threshold)
	if err != nil {
		return err
	}
	return nil
}

func (d *InMemRepo) FindThresholdById(id int32) (*domain.Threshold, error) {
	return FindById[domain.Threshold](d, domain.ThresholdTable, id)
}

func (d *InMemRepo) FindAllThresholds() ([]*domain.Threshold, error) {
	return FindAll[domain.Threshold](d, domain.ThresholdTable)
}

func (d *InMemRepo) DeleteThresholdById(id int32) error {
	return d.DeleteById(domain.ThresholdTable, id)
}
