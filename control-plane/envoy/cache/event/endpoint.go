package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processEndpointChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			endpoint := change.Before.(*domain.Endpoint)
			parser.updateEndpoint(actions, entityVersions, nodeGroup, endpoint)
		} else {
			endpoint := change.After.(*domain.Endpoint)
			parser.updateEndpoint(actions, entityVersions, nodeGroup, endpoint)
		}
	}
}

func (parser *changeEventParserImpl) updateEndpoint(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, endpoint *domain.Endpoint) {
	cluster, err := parser.dao.FindClusterById(endpoint.ClusterId)
	if err != nil {
		logger.Panicf("Failed to find cluster by id using DAO: %v", err)
	}
	if cluster != nil {
		granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
		actions.Put(action.EnvoyCluster, &granularUpdate)
	}
}

func (builder *compositeUpdateBuilder) processEndpointChanges(changes []memdb.Change) {
	for _, change := range changes {
		var endpoint *domain.Endpoint = nil
		if change.Deleted() {
			endpoint = change.Before.(*domain.Endpoint)
		} else {
			endpoint = change.After.(*domain.Endpoint)
		}
		builder.updateCluster(endpoint.ClusterId)
	}
}

func (builder *compositeUpdateBuilder) updateEndpoint(endpointId int32) {
	endpoint, err := builder.repo.FindEndpointById(endpointId)
	if err != nil {
		logger.Panicf("Failed to find endpoint by id using DAO: %v", err)
	}
	if endpoint != nil { // if nil, endpoint is deleted and its deletion is processed by domain.Endpoint change event
		builder.updateCluster(endpoint.ClusterId)
	}
}
