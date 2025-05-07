package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processThresholdChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			threshold := change.Before.(*domain.Threshold)
			parser.updateThreshold(nodeGroup, actions, entityVersions, threshold, changes)
		} else {
			threshold := change.After.(*domain.Threshold)
			parser.updateThreshold(nodeGroup, actions, entityVersions, threshold, changes)
		}
	}
}

func (parser *changeEventParserImpl) updateThreshold(nodeGroup string, actions action.ActionsMap, entityVersions map[string]string, threshold *domain.Threshold, changes []memdb.Change) {
	circuitBreakers, err := parser.dao.FindAllCircuitBreakers()
	if err != nil {
		logger.Panicf("failed to find circuit breakers using DAO: %v", err)
	}
	clusters, err := parser.dao.FindAllClusters()
	if err != nil {
		logger.Panicf("failed to find clusters using DAO: %v", err)
	}

	for _, circuitBreaker := range circuitBreakers {
		if circuitBreaker.ThresholdId == threshold.Id {
			for _, cluster := range clusters {
				if cluster.CircuitBreakerId == circuitBreaker.Id {
					granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
					actions.Put(action.EnvoyCluster, &granularUpdate)
				}
			}
		}
	}
}
