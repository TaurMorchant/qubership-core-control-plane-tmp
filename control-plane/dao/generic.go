package dao

import (
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (d *InMemRepo) SaveUnique(tableName string, unique domain.Unique) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	if err := d.idGenerator.Generate(unique); err != nil {
		return err
	}
	return d.storage.Save(txCtx.tx, tableName, unique)
}

func (d *InMemRepo) Delete(tableName string, entity any) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return txCtx.tx.Delete(tableName, entity)
}

func (d *InMemRepo) DeleteById(tableName string, id ...any) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	_, err := txCtx.tx.DeleteAll(tableName, "id", id...)
	return err
}

func FindById[T any](d *InMemRepo, tableName string, id ...any) (*T, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	if entity, err := d.storage.FindById(txCtx.tx, tableName, id...); err == nil {
		if entity != nil {
			return entity.(*T), nil
		}
		return nil, nil
	} else {
		return nil, errors.WrapPrefix(err, "error in mem storage FindById", 1)
	}
}

func FindByIndex[T any](d *InMemRepo, tableName, idxName string, values ...any) ([]*T, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	if entities, err := d.storage.FindByIndex(txCtx.tx, tableName, idxName, values...); err == nil {
		if entities != nil {
			return entities.([]*T), nil
		}
		return nil, nil
	} else {
		return nil, errors.WrapPrefix(err, "error in mem storage FindByIndex", 1)
	}
}

func FindFirstByIndex[T any](d *InMemRepo, tableName, idxName string, values ...any) (*T, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	if entity, err := d.storage.FindFirstByIndex(txCtx.tx, tableName, idxName, values...); err == nil {
		if entity == nil {
			return nil, nil
		}
		return entity.(*T), nil
	} else {
		return nil, errors.WrapPrefix(err, "error in mem storage FindFirstByIndex", 1)
	}
}

func FindAll[T any](d *InMemRepo, tableName string) ([]*T, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	if entities, err := d.storage.FindAll(txCtx.tx, tableName); err == nil {
		if entities != nil {
			return entities.([]*T), nil
		}
		return nil, nil
	} else {
		return nil, errors.WrapPrefix(err, "error in mem storage FindAll", 1)
	}
}
