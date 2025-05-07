package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processClusterNodeGroupChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			parser.processClusterNodeGroupDelete(actions, entityVersions, nodeGroup, change)
		} else {
			parser.processClusterNodeGroupUpdate(actions, entityVersions, change)
		}
	}
}

func (parser *changeEventParserImpl) processClusterNodeGroupDelete(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, change memdb.Change) {
	clusterNodeGroup := change.Before.(*domain.ClustersNodeGroup)
	if clusterNodeGroup.NodegroupsName == nodeGroup {
		cluster, err := parser.dao.FindClusterById(clusterNodeGroup.ClustersId)
		if err != nil {
			logger.Panicf("Failed to find cluster by id using DAO: %v", err)
		}
		if cluster != nil {
			granularUpdate := parser.updateActionFactory.ClusterDelete(nodeGroup, entityVersions[domain.ClusterTable], cluster)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	} // else it is not the node group that processed by this change event
}

func (parser *changeEventParserImpl) processClusterNodeGroupUpdate(actions action.ActionsMap, entityVersions map[string]string, change memdb.Change) {
	clusterNodeGroup := change.After.(*domain.ClustersNodeGroup)
	cluster, err := parser.dao.FindClusterById(clusterNodeGroup.ClustersId)
	if err != nil {
		logger.Panicf("Failed to find cluster by id using DAO: %v", err)
	}
	granularUpdate := parser.updateActionFactory.ClusterUpdate(clusterNodeGroup.NodegroupsName, entityVersions[domain.ClusterTable], cluster)
	actions.Put(action.EnvoyCluster, &granularUpdate)
}

func (parser *changeEventParserImpl) processClusterChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			granularUpdate := parser.updateActionFactory.ClusterDelete(nodeGroup, entityVersions[domain.ClusterTable], change.Before)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		} else {
			granularUpdate := parser.updateActionFactory.ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], change.After)
			actions.Put(action.EnvoyCluster, &granularUpdate)
		}
	}
}

func (builder *compositeUpdateBuilder) processClusterNodeGroupChanges(changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			clusterNodeGroup := change.Before.(*domain.ClustersNodeGroup)
			cluster := domain.Cluster{Id: clusterNodeGroup.ClustersId}
			builder.addDeleteAction(clusterNodeGroup.NodegroupsName, action.EnvoyCluster, cluster)
		} else {
			clusterNodeGroup := change.After.(*domain.ClustersNodeGroup)
			cluster, err := builder.repo.FindClusterById(clusterNodeGroup.ClustersId)
			if err != nil {
				logger.Panicf("Failed to find cluster by id using DAO: %v", err)
			}
			builder.addUpdateAction(clusterNodeGroup.NodegroupsName, action.EnvoyCluster, cluster)
		}
	}
}

func (builder *compositeUpdateBuilder) processClusterChanges(changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			// nothing to do since here wo don't know to which nodeGroup this cluster was bound:
			// actual cluster deletion from each nodeGroup will be performed by domain.ClustersNodeGroup changes
		} else {
			cluster := change.After.(*domain.Cluster)
			clusterNodeGroups, err := builder.repo.FindNodeGroupsByCluster(cluster)
			if err != nil {
				logger.Panicf("Failed to find node groups cluster using DAO: %v", err)
			}
			for _, nodeGroup := range clusterNodeGroups {
				builder.addUpdateAction(nodeGroup.Name, action.EnvoyCluster, cluster)
			}
		}
	}
}

func (builder *compositeUpdateBuilder) updateCluster(clusterId int32) {
	cluster, err := builder.repo.FindClusterById(clusterId)
	if err != nil {
		logger.Panicf("Failed to find cluster by id using DAO: %v", err)
	}
	builder.updateClusterInternal(cluster)
}

func (builder *compositeUpdateBuilder) updateClusterInternal(cluster *domain.Cluster) {
	if cluster == nil { // cluster is deleted and should be updated by domain.ClustersNodeGroup change event
		return
	} else {
		nodeGroups, err := builder.repo.FindNodeGroupsByCluster(cluster)
		if err != nil {
			logger.Panicf("Failed to find node groups by cluster using DAO: %v", err)
		}
		for _, nodeGroup := range nodeGroups {
			builder.addUpdateAction(nodeGroup.Name, action.EnvoyCluster, cluster)
		}
	}
}
