package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processListenerChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			granularUpdate := parser.updateActionFactory.ListenerDelete(nodeGroup, entityVersions[domain.ListenerTable], change.Before)
			actions.Put(action.EnvoyListener, &granularUpdate)
		} else {
			granularUpdate := parser.updateActionFactory.ListenerUpdate(nodeGroup, entityVersions[domain.ListenerTable], change.After)
			actions.Put(action.EnvoyListener, &granularUpdate)
		}
	}
}

func (builder *compositeUpdateBuilder) processListenerChanges(changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			listener := change.Before.(*domain.Listener)
			builder.addDeleteAction(listener.NodeGroupId, action.EnvoyListener, listener)
		} else {
			listener := change.After.(*domain.Listener)
			builder.addUpdateAction(listener.NodeGroupId, action.EnvoyListener, listener)
		}
	}
}
