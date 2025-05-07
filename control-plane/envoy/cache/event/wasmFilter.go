package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processWasmFilterChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string) {
	listeners, err := parser.dao.FindListenersByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Panicf("failed to find listeners using DAO: %v", err)
	}

	for _, listener := range listeners {
		granularUpdate := parser.updateActionFactory.ListenerUpdate(nodeGroup, entityVersions[domain.ListenerTable], listener)
		actions.Put(action.EnvoyListener, &granularUpdate)
	}
}

func (builder *compositeUpdateBuilder) processWasmFilterChanges(changes []memdb.Change) {
	for _, change := range changes {
		var wasmFilter *domain.WasmFilter = nil
		if change.Deleted() {
			wasmFilter = change.Before.(*domain.WasmFilter)
		} else {
			wasmFilter = change.After.(*domain.WasmFilter)
		}
		listenerIds, err := builder.repo.FindListenerIdsByWasmFilterId(wasmFilter.Id)
		if err != nil {
			logger.Panicf("failed to find listeners using DAO: %v", err)
		}
		for _, listenerId := range listenerIds {
			listener, err := builder.repo.FindListenerById(listenerId)
			if err != nil {
				logger.Panicf("failed to find listener by id using DAO: %v", err)
			}
			builder.addUpdateAction(listener.NodeGroupId, action.EnvoyListener, listener)
		}
	}
}

func (builder *compositeUpdateBuilder) processListenersWasmFilterChanges(changes []memdb.Change) {
	for _, change := range changes {
		var listenersWasmFilter *domain.ListenersWasmFilter = nil
		if change.Deleted() {
			listenersWasmFilter = change.Before.(*domain.ListenersWasmFilter)
		} else {
			listenersWasmFilter = change.After.(*domain.ListenersWasmFilter)
		}
		listener, err := builder.repo.FindListenerById(listenersWasmFilter.ListenerId)
		if err != nil {
			logger.Panicf("failed to find listener by id using DAO: %v", err)
		}
		builder.addUpdateAction(listener.NodeGroupId, action.EnvoyListener, listener)
	}
}
