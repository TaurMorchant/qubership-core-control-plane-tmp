package dao

import (
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/data"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
)

type InMemDao struct {
	InMemRepo
}

type InMemRepo struct {
	storage               ram.RamStorage
	idGenerator           IDGenerator
	txCtx                 *TxContext
	beforeCommitCallbacks []func([]memdb.Change) error
}

type TxContext struct {
	tx    ram.Txn
	local bool
	write bool
}

func NewInMemDao(storage ram.RamStorage, generator IDGenerator, beforeCommitCallbacks []func([]memdb.Change) error) *InMemDao {
	return &InMemDao{InMemRepo{
		storage:               storage,
		idGenerator:           generator,
		beforeCommitCallbacks: beforeCommitCallbacks,
	}}
}

func (d *InMemRepo) getTxCtx(write bool) *TxContext {
	if d.txCtx == nil {
		return &TxContext{tx: d.storage.Tx(write), write: write, local: true}
	}
	if !d.txCtx.write && write {
		panic("Requested write transaction, but current transaction is read-only")
	}
	return d.txCtx
}

func (ctx *TxContext) closeIfLocal() {
	if ctx.local {
		ctx.tx.Abort()
	}
}

func (d *InMemRepo) executeBeforeCommitCallbacks(changes []memdb.Change) error {
	for _, callback := range d.beforeCommitCallbacks {
		if err := callback(changes); err != nil {
			return errors.WrapPrefix(err, "Before commit callback caused error", 0)
		}
	}
	return nil
}

func (d *InMemRepo) WithWTx(payload func(dao Repository) error) ([]memdb.Change, error) {
	_, changes, err := d.WithWTxVal(func(dao Repository) (interface{}, error) {
		return nil, payload(dao)
	})

	return changes, err
}

func (d *InMemRepo) WithWTxVal(payload func(dao Repository) (interface{}, error)) (interface{}, []memdb.Change, error) {
	tx := d.storage.WriteTx()
	txCtx := &TxContext{local: false, tx: tx, write: true}
	tx.TrackChanges()
	defer tx.Abort()
	val, err := payload(&InMemRepo{storage: d.storage, txCtx: txCtx, idGenerator: d.idGenerator, beforeCommitCallbacks: d.beforeCommitCallbacks})
	if err != nil {
		return nil, nil, errors.WrapPrefix(err, "Executing actions in write transaction caused error", 0)
	}
	changes := tx.Changes()
	if err := d.executeBeforeCommitCallbacks(changes); err != nil {
		return nil, nil, errors.WrapPrefix(err, "Flushing changes to persistence storage caused error", 0)
	}
	tx.Commit()
	return val, changes, nil
}

func (d *InMemRepo) WithRTxVal(payload func(dao Repository) (interface{}, error)) (interface{}, error) {
	tx := d.storage.ReadTx()
	defer tx.Abort()
	txCtx := &TxContext{local: false, tx: tx, write: false}
	val, err := payload(&InMemRepo{storage: d.storage, txCtx: txCtx, idGenerator: d.idGenerator, beforeCommitCallbacks: d.beforeCommitCallbacks})
	if err != nil {
		return nil, errors.WrapPrefix(err, "Executing actions in read transaction caused error", 0)
	}
	return val, nil
}

func (d *InMemRepo) WithRTx(payload func(dao Repository) error) error {
	_, err := d.WithRTxVal(func(dao Repository) (interface{}, error) {
		return nil, payload(dao)
	})

	return err
}

func (d *InMemRepo) SaveEnvoyConfigVersion(version *domain.EnvoyConfigVersion) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return d.storage.Save(txCtx.tx, domain.EnvoyConfigVersionTable, version)
}

func (d *InMemRepo) SaveEntity(table string, entity interface{}) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return d.storage.Save(txCtx.tx, table, entity)
}

func (d *InMemRepo) FindEnvoyConfigVersion(nodeGroup, entityType string) (*domain.EnvoyConfigVersion, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	version, err := d.storage.FindById(txCtx.tx, domain.EnvoyConfigVersionTable, nodeGroup, entityType)
	if err != nil {
		return nil, err
	}
	if version != nil {
		return version.(*domain.EnvoyConfigVersion), nil
	}
	return nil, nil
}

func (d InMemDao) Backup() (*data.Snapshot, error) {
	return d.storage.Backup()
}

func (d InMemDao) Restore(snapshot data.Snapshot) error {
	return d.storage.Restore(snapshot)
}
