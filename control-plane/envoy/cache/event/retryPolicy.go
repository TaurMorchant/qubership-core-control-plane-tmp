package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processRetryPolicyChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			retryPolicy := change.Before.(*domain.RetryPolicy)
			parser.updateRetryPolicy(actions, entityVersions, nodeGroup, retryPolicy)
		} else {
			retryPolicy := change.After.(*domain.RetryPolicy)
			parser.updateRetryPolicy(actions, entityVersions, nodeGroup, retryPolicy)
		}
	}
}

func (parser *changeEventParserImpl) updateRetryPolicy(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, retryPolicy *domain.RetryPolicy) {
	if retryPolicy.RouteId != 0 {
		route, err := parser.dao.FindRouteById(retryPolicy.RouteId)
		if err != nil {
			logger.Panicf("Failed to find route by id using DAO: %v", err)
		}
		if route != nil {
			parser.updateRoute(actions, entityVersions, nodeGroup, route)
		}
	}
}
func (builder *compositeUpdateBuilder) processRetryPolicyChanges(changes []memdb.Change) {
	for _, change := range changes {
		var retryPolicy *domain.RetryPolicy = nil
		if change.Deleted() {
			retryPolicy = change.Before.(*domain.RetryPolicy)
		} else {
			retryPolicy = change.After.(*domain.RetryPolicy)
		}
		builder.updateRoute(retryPolicy.RouteId)
	}
}
