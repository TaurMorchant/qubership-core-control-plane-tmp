package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (d *InMemRepo) FindAllWasmFilters() ([]*domain.WasmFilter, error) {
	return FindAll[domain.WasmFilter](d, domain.WasmFilterTable)
}

func (d *InMemRepo) FindWasmFilterByName(filterName string) (*domain.WasmFilter, error) {
	return FindFirstByIndex[domain.WasmFilter](d, domain.WasmFilterTable, "name", filterName)
}

func (d *InMemRepo) FindWasmFilterByListenerId(listenerId int32) ([]*domain.WasmFilter, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	if found, err := d.storage.FindByIndex(txCtx.tx, domain.ListenersWasmFilterTable, "listenerId", listenerId); err == nil {
		listenerToWasmFilters := found.([]*domain.ListenersWasmFilter)
		wasmFilters := make([]*domain.WasmFilter, len(listenerToWasmFilters))
		for i, listenerToWasmFilter := range listenerToWasmFilters {
			if wf, err := d.storage.FindById(txCtx.tx, domain.WasmFilterTable, listenerToWasmFilter.WasmFilterId); err == nil {
				wasmFilters[i] = wf.(*domain.WasmFilter)
			} else {
				return nil, err
			}
		}
		return wasmFilters, nil
	} else {
		return nil, err
	}
}

func (d *InMemRepo) SaveWasmFilter(wasmFilter *domain.WasmFilter) error {
	return d.SaveUnique(domain.WasmFilterTable, wasmFilter)
}

func (d *InMemRepo) DeleteWasmFilterByName(filterName string) (int32, error) {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	filterToDelete, err := d.FindWasmFilterByName(filterName)
	if err != nil {
		return 0, err
	}
	_, err = txCtx.tx.DeleteAll(domain.WasmFilterTable, "id", filterToDelete.Id)
	if err != nil {
		return 0, err
	}
	return filterToDelete.Id, nil
}

func (d *InMemRepo) DeleteWasmFilterById(id int32) error {
	return d.DeleteById(domain.WasmFilterTable, id)
}
