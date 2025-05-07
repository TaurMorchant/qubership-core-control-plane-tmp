package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (srv *Service) PutWasmFilter(dao dao.Repository, filter *domain.WasmFilter) error {
	if filter.Id == 0 {
		existing, err := dao.FindWasmFilterByName(filter.Name)
		if err != nil {
			logger.Errorf("Error while trying to find existing wasm filter by name %s: %s", filter.Name, err.Error())
			return err
		}
		if existing != nil {
			filter.Id = existing.Id
		}
	}
	return dao.SaveWasmFilter(filter)
}

func (srv *Service) PutListenerWasmFilterIfAbsent(dao dao.Repository, relation *domain.ListenersWasmFilter) error {
	alreadyHasFilter, err := dao.HasWasmFilterWithId(relation.ListenerId, relation.WasmFilterId)
	if err != nil {
		logger.Errorf("Error while check relation by listenerId=%d and wasmFilterId=%d: %s", relation.ListenerId, relation.WasmFilterId, err.Error())
		return err
	}
	if alreadyHasFilter {
		logger.Infof("WASM filter with id=%d is already connected to listener with id=%d", relation.WasmFilterId, relation.ListenerId)
		return nil
	}
	err = dao.SaveListenerWasmFilter(relation)
	if err != nil {
		logger.Errorf("Error while saving listener wasm filter relation %v: %s", relation, err.Error())
		return err
	}
	return nil
}
