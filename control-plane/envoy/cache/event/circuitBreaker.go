package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processCircuitBreakerChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			circuitBreaker := change.Before.(*domain.CircuitBreaker)
			parser.updateCircuitBreaker(nodeGroup, actions, entityVersions, circuitBreaker)
		} else {
			circuitBreaker := change.After.(*domain.CircuitBreaker)
			parser.updateCircuitBreaker(nodeGroup, actions, entityVersions, circuitBreaker)
		}
	}
}

func (parser *changeEventParserImpl) updateCircuitBreaker(nodeGroup string, actions action.ActionsMap, entityVersions map[string]string, circuitBreaker *domain.CircuitBreaker) {
	clusters, err := parser.dao.FindAllClusters()
	if err != nil {
		logger.Panicf("failed to find clusters using DAO: %v", err)
	}

	for _, cluster := range clusters {
		if cluster.CircuitBreakerId == circuitBreaker.Id {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}
