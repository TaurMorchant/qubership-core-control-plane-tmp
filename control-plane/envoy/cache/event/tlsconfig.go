package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/provider"
)

func (parser *changeEventParserImpl) processTlsConfigChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			tlsConfig := change.Before.(*domain.TlsConfig)
			parser.updateTlsConfig(actions, entityVersions, nodeGroup, tlsConfig)
		} else {
			tlsConfig := change.After.(*domain.TlsConfig)
			parser.updateTlsConfig(actions, entityVersions, nodeGroup, tlsConfig)
		}
	}
}

func (parser *changeEventParserImpl) updateTlsConfig(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, tlsConfig *domain.TlsConfig) {
	clusters, err := parser.dao.FindAllClusters()
	if err != nil {
		logger.Panicf("failed to find clusters using DAO: %v", err)
	}

	for _, cluster := range clusters {
		hasGlobalTLSConfig, err := parser.hasGlobalTLSConfig(cluster, nodeGroup)
		if err != nil {
			logger.Panicf("failed to find clusters global TLS configs: %v", err)
		}
		if cluster.TLSId == tlsConfig.Id || hasGlobalTLSConfig || (cluster.TLS != nil && cluster.TLS.Enabled != tlsConfig.Enabled) {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}

func (parser *changeEventParserImpl) hasGlobalTLSConfig(cluster *domain.Cluster, nodeGroup string) (bool, error) {
	enabledTlsConfigs := make([]*domain.TlsConfig, 0)
	tlsConfigs, err := provider.GetTlsService().GetGlobalTlsConfigs(cluster, nodeGroup)
	if err != nil {
		return false, err
	}
	for _, tlsConfig := range tlsConfigs {
		if tlsConfig.Enabled {
			enabledTlsConfigs = append(enabledTlsConfigs, tlsConfig)
		}
	}
	return len(enabledTlsConfigs) != 0, nil
}
