package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processTcpKeepaliveChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			entity := change.Before.(*domain.TcpKeepalive)
			parser.updateTcpKeepalive(nodeGroup, actions, entityVersions, entity, changes)
		} else {
			entity := change.After.(*domain.TcpKeepalive)
			parser.updateTcpKeepalive(nodeGroup, actions, entityVersions, entity, changes)
		}
	}
}

func (parser *changeEventParserImpl) updateTcpKeepalive(nodeGroup string, actions action.ActionsMap, entityVersions map[string]string, tcpKeepalive *domain.TcpKeepalive, changes []memdb.Change) {
	clusters, err := parser.dao.FindAllClusters()
	if err != nil {
		logger.Panicf("Could not to find clusters using DAO: %v", err)
	}

	for _, cluster := range clusters {
		if cluster.TcpKeepaliveId == tcpKeepalive.Id {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}

func (builder *compositeUpdateBuilder) processTcpKeepaliveChanges(changes []memdb.Change) {
	for _, change := range changes {
		var tcpKeepalive *domain.TcpKeepalive = nil
		if change.Deleted() {
			tcpKeepalive = change.Before.(*domain.TcpKeepalive)
		} else {
			tcpKeepalive = change.After.(*domain.TcpKeepalive)
		}
		clusters, err := builder.repo.FindAllClusters()
		if err != nil {
			logger.Panicf("Could not to find clusters using DAO: %v", err)
		}
		for _, cluster := range clusters {
			if cluster.TcpKeepaliveId == tcpKeepalive.Id {
				builder.updateClusterInternal(cluster)
			}
		}

	}
}
