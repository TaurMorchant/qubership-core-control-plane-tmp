package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processDeploymentVersionChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			logger.Warnf("Deployment version deleted")
		} else {
			// update everything related to this version
			dv := change.After.(*domain.DeploymentVersion)
			parser.processDeploymentVersionChangesForClusters(actions, entityVersions, nodeGroup, dv)
			parser.processDeploymentVersionChangesForRouteConfigs(actions, entityVersions, nodeGroup, dv)
		}
	}
}

func (parser *changeEventParserImpl) processDeploymentVersionChangesForClusters(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, dv *domain.DeploymentVersion) {
	endpoints, err := parser.dao.FindEndpointsByDeploymentVersion(dv.Version)
	if err != nil {
		logger.Panicf("Failed to find endpoints by deployment version %v using DAO: %v", dv.Version, err)
	}
	if endpoints != nil && len(endpoints) > 0 {
		clusters, err := parser.dao.FindClusterByEndpointIn(endpoints)
		if err != nil {
			logger.Panicf("Failed to find clusters by endpoints list using DAO: %v", err)
		}
		for _, cluster := range clusters {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}

func (parser *changeEventParserImpl) processDeploymentVersionChangesForRouteConfigs(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, dv *domain.DeploymentVersion) {
	routeConfigs, err := parser.dao.FindRouteConfigsByRouteDeploymentVersion(dv.Version)
	if err != nil {
		logger.Panicf("Failed to find route configs by route version %v using DAO: %v", dv.Version, err)
	}
	if routeConfigs != nil {
		for _, routeConfig := range routeConfigs {
			granularUpdate := parser.updateActionFactory.RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig)
			actions.Put(action.EnvoyRouteConfig, &granularUpdate)
		}
	}
}

func (builder *compositeUpdateBuilder) processDeploymentVersionChanges(changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			logger.Debugf("Deletion of deployment itself will not delete any entities: change event should also contain changes on explicit deletion of those entities")
			return
		} else {
			// update everything related to this version
			dv := change.After.(*domain.DeploymentVersion)
			endpoints, err := builder.repo.FindEndpointsByDeploymentVersion(dv.Version)
			if err != nil {
				logger.Panicf("Failed to find endpoints by deployment version %v using DAO: %v", dv.Version, err)
			}
			if endpoints != nil && len(endpoints) > 0 {
				clusters, err := builder.repo.FindClusterByEndpointIn(endpoints)
				if err != nil {
					logger.Panicf("Failed to find clusters by endpoints list using DAO: %v", err)
				}
				for _, cluster := range clusters {
					nodeGroups, err := builder.repo.FindNodeGroupsByCluster(cluster)
					if err != nil {
						logger.Panicf("Failed to find node groups by cluster using DAO: %v", err)
					}
					for _, nodeGroup := range nodeGroups {
						builder.addUpdateAction(nodeGroup.Name, action.EnvoyCluster, cluster)
					}
				}
			}
			routeConfigs, err := builder.repo.FindRouteConfigsByRouteDeploymentVersion(dv.Version)
			if err != nil {
				logger.Panicf("Failed to find route configs by route version %v using DAO: %v", dv.Version, err)
			}
			if routeConfigs != nil {
				for _, routeConfig := range routeConfigs {
					builder.addUpdateAction(routeConfig.NodeGroupId, action.EnvoyRouteConfig, routeConfig)
				}
			}
		}
	}
}
