package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processRouteConfigurationChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			granularUpdate := parser.updateActionFactory.RouteConfigDelete(nodeGroup, entityVersions[domain.RouteConfigurationTable], change.Before)
			actions.Put(action.EnvoyRouteConfig, &granularUpdate)
		} else {
			granularUpdate := parser.updateActionFactory.RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], change.After)
			actions.Put(action.EnvoyRouteConfig, &granularUpdate)
		}
	}
}

func (parser *changeEventParserImpl) updateRouteConfig(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, routeConfig *domain.RouteConfiguration) {
	granularUpdate := parser.updateActionFactory.RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig)
	actions.Put(action.EnvoyRouteConfig, &granularUpdate)
}

func (builder *compositeUpdateBuilder) processRouteConfigurationChanges(changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			routeConfig := change.Before.(*domain.RouteConfiguration)
			builder.addDeleteAction(routeConfig.NodeGroupId, action.EnvoyRouteConfig, routeConfig)
		} else {
			routeConfig := change.After.(*domain.RouteConfiguration)
			builder.addUpdateAction(routeConfig.NodeGroupId, action.EnvoyRouteConfig, routeConfig)
		}
	}
}

func (builder *compositeUpdateBuilder) updateRouteConfig(routeConfigId int32) {
	logger.Debugf("Updating route config with id %d", routeConfigId)
	routeConfig, err := builder.repo.FindRouteConfigById(routeConfigId)
	if err != nil {
		logger.Panicf("Failed to find route config by id using DAO: %v", err)
	}
	if routeConfig != nil { // if nil, this route config is being deleted by domain.RouteConfiguration change event
		builder.addUpdateAction(routeConfig.NodeGroupId, action.EnvoyRouteConfig, routeConfig)
		logger.Debugf("Added update route config %+v action", *routeConfig)
	}
}
