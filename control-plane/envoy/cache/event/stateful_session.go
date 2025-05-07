package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
)

func (parser *changeEventParserImpl) processStatefulSessionChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		var statefulSes *domain.StatefulSession
		if change.Deleted() {
			statefulSes = change.Before.(*domain.StatefulSession)
		} else {
			statefulSes = change.After.(*domain.StatefulSession)
		}

		route, err := parser.dao.FindRouteByStatefulSession(statefulSes.Id)
		if err != nil {
			logger.Panicf("Failed to find route by id using DAO:\n %v", err)
		}
		if route != nil {
			parser.updateVirtualHostIfConnectedToNodeGroup(actions, entityVersions, nodeGroup, route.VirtualHostId)
			continue
		}

		endpoint, err := parser.dao.FindEndpointByStatefulSession(statefulSes.Id)
		if err != nil {
			logger.Panicf("Failed to find route by id using DAO:\n %v", err)
		}
		if endpoint != nil {
			// physically we cannot set stateful session for endpoint in envoy,
			// so need to update stateful session for each route of the endpoint instead
			routes := findRoutesByEndpointIfExist(parser.dao, endpoint)
			for _, route := range routes {
				parser.updateVirtualHostIfConnectedToNodeGroup(actions, entityVersions, nodeGroup, route.VirtualHostId)
			}
			continue
		}

		// this is per-cluster configuration, physically we cannot set stateful session for cluster in envoy,
		// so need to update stateful session for each route of the cluster instead
		routes := findRoutesByClusterIfExist(parser.dao, statefulSes.ClusterName, msaddr.Namespace{Namespace: statefulSes.Namespace})
		for _, route := range routes {
			parser.updateVirtualHostIfConnectedToNodeGroup(actions, entityVersions, nodeGroup, route.VirtualHostId)
		}
	}
}

func (parser *changeEventParserImpl) updateVirtualHostIfConnectedToNodeGroup(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, virtualHostId int32) {
	virtualHost, err := parser.dao.FindVirtualHostById(virtualHostId)
	if err != nil {
		logger.Panicf("Failed to find virtual host by id using DAO:\n %v", err)
	}
	routeConfig, err := parser.dao.FindRouteConfigById(virtualHost.RouteConfigurationId)
	if err != nil {
		logger.Panicf("Failed to find route configuration by id using DAO: %v", err)
	}
	if routeConfig != nil && routeConfig.NodeGroupId == nodeGroup {
		parser.updateRouteConfig(actions, entityVersions, nodeGroup, routeConfig)
	}
}

func (builder *compositeUpdateBuilder) processStatefulSessionCookieChanges(changes []memdb.Change) {
	logger.Debug("Processing stateful session multiple change event")
	for _, change := range changes {
		var statefulSes *domain.StatefulSession = nil
		if change.Deleted() {
			statefulSes = change.Before.(*domain.StatefulSession)
		} else {
			statefulSes = change.After.(*domain.StatefulSession)
		}

		route, err := builder.repo.FindRouteByStatefulSession(statefulSes.Id)
		if err != nil {
			logger.Panicf("Failed to find route by stateful session using DAO:\n %v", err)
		}
		if route != nil {
			builder.updateVirtualHost(route.VirtualHostId)
			continue
		}

		endpoint, err := builder.repo.FindEndpointByStatefulSession(statefulSes.Id)
		if err != nil {
			logger.Panicf("Failed to find endpoint by stateful session using DAO:\n %v", err)
		}
		if endpoint != nil {
			builder.updateEndpointStatefulSession(endpoint)
			continue
		}

		builder.updateClusterStatefulSession(statefulSes.ClusterName, msaddr.Namespace{Namespace: statefulSes.Namespace})
	}
}

func (builder *compositeUpdateBuilder) updateClusterStatefulSession(clusterName string, namespace msaddr.Namespace) {
	// physically we cannot set stateful session for cluster in envoy,
	// so need to set stateful session for each route of this cluster instead
	routes := findRoutesByClusterIfExist(builder.repo, clusterName, namespace)
	for _, route := range routes {
		builder.updateVirtualHost(route.VirtualHostId)
	}
}

func findRoutesByClusterIfExist(repo dao.Repository, clusterName string, namespace msaddr.Namespace) []*domain.Route {
	clusterKeyPrefix := clusterkey.DefaultClusterKeyGenerator.BuildKeyPrefix(clusterName, namespace)
	routes, err := repo.FindRoutesByClusterNamePrefix(clusterKeyPrefix)
	if err != nil {
		logger.Panicf("Failed to find cluster routes using DAO:\n %v", err)
	}
	logger.Debugf("Found %d routes by stateful session %s[%s] change event", len(routes), clusterName, namespace.Namespace)
	return routes
}

func (builder *compositeUpdateBuilder) updateEndpointStatefulSession(endpoint *domain.Endpoint) {
	// physically we cannot set stateful session for endpoint in envoy,
	// so need to set stateful session for each route of this endpoint instead
	routes := findRoutesByEndpointIfExist(builder.repo, endpoint)
	for _, route := range routes {
		builder.updateVirtualHost(route.VirtualHostId)
	}
}

func findRoutesByEndpointIfExist(repo dao.Repository, endpoint *domain.Endpoint) []*domain.Route {
	if endpoint == nil {
		return nil // endpoint is already deleted: nothing to do
	}
	cluster, err := repo.FindClusterById(endpoint.ClusterId)
	if err != nil {
		logger.Panicf("Failed to find cluster by id using DAO:\n %v", err)
	}
	if cluster == nil {
		return nil // cluster is already deleted: nothing to do
	}
	routes, err := repo.FindRoutesByClusterNameAndDeploymentVersion(cluster.Name, endpoint.DeploymentVersion)
	if err != nil {
		logger.Panicf("Failed to find routes by cluster name and deployment version using DAO:\n %v", err)
	}
	return routes
}
