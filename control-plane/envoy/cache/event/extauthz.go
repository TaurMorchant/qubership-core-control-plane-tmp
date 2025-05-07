package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processExtAuthzFilterChanges(actions action.ActionsMap, entityVersions map[string]string, changes []memdb.Change) {
	processChanges(changes, func(entity *domain.ExtAuthzFilter) {
		parser.updateListenersAndRouteConfigs(actions, entityVersions, entity.NodeGroup)
	})
}

func (builder *compositeUpdateBuilder) processExtAuthzFilterChanges(changes []memdb.Change) {
	processChanges(changes, func(entity *domain.ExtAuthzFilter) {
		builder.updateListenersAndRouteConfigs(entity.NodeGroup)
	})
}

func processChanges[T any](changes []memdb.Change, updateFunc func(entity T)) {
	for _, change := range changes {
		var entity T
		if change.Deleted() {
			entity = change.Before.(T)
		} else {
			entity = change.After.(T)
		}
		updateFunc(entity)
	}
}
