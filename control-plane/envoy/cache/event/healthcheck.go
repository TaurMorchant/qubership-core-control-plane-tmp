package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processHealthCheckChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			healthCheck := change.Before.(*domain.HealthCheck)
			parser.updateHealthCheck(actions, entityVersions, nodeGroup, healthCheck)
		} else {
			healthCheck := change.After.(*domain.HealthCheck)
			parser.updateHealthCheck(actions, entityVersions, nodeGroup, healthCheck)
		}
	}
}

func (parser *changeEventParserImpl) updateHealthCheck(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, healthCheck *domain.HealthCheck) {
	if healthCheck.ClusterId != 0 {
		cluster, err := parser.dao.FindClusterById(healthCheck.ClusterId)
		if err != nil {
			logger.Panicf("Failed to find HealthCheck's cluster by ClusterId using DAO: %v", err)
		}
		if cluster != nil {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}

func (builder *compositeUpdateBuilder) processHealthCheckChanges(changes []memdb.Change) {
	for _, change := range changes {
		var healthCheck *domain.HealthCheck = nil
		if change.Deleted() {
			healthCheck = change.Before.(*domain.HealthCheck)
		} else {
			healthCheck = change.After.(*domain.HealthCheck)
		}
		builder.updateCluster(healthCheck.ClusterId)
	}
}
