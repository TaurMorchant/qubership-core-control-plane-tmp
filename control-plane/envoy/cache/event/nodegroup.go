package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processNodeGroupChanges(actions action.ActionsMap, entityVersions map[string]string, _ string, changes []memdb.Change) {
	for _, change := range changes {
		if !change.Deleted() {
			nodeGroup, ok := change.After.(*domain.NodeGroup)
			if !ok {
				logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to cast %v to NodeGroup", change.After)
			}

			if err := parser.configBuilder.RegisterGateway(nodeGroup); err != nil {
				logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to register gateway %+v in builder using DAO:\n %v", nodeGroup, err)
			}

			if change.Before == nil {
				parser.updateRuntimeConfigs(actions, nodeGroup)
			}

			if change.Before == nil || change.Before.(*domain.NodeGroup).GatewayType != nodeGroup.GatewayType {
				// only gateway type change should trigger envoy route config and listener update
				parser.updateListenersAndRouteConfigs(actions, entityVersions, nodeGroup.Name)
			}
		}
	}
}

func (parser *changeEventParserImpl) updateRuntimeConfigs(actions action.ActionsMap, ng *domain.NodeGroup) {
	updateAction := parser.updateActionFactory.RuntimeUpdateAction(ng.Name)
	ac := &action.GranularEntityUpdate{Action: updateAction, IsDelete: false, EntityId: ng.GetId()}
	actions.Put(action.EnvoyRuntime, ac)
}

func (parser *changeEventParserImpl) updateListenersAndRouteConfigs(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string) {
	routeConfigs, err := parser.dao.FindRouteConfigsByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to load route configs for node group %s using DAO:\n %v", nodeGroup, err)
	}
	for _, routeConfig := range routeConfigs {
		granularUpdate := parser.updateActionFactory.RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig)
		actions.Put(action.EnvoyRouteConfig, &granularUpdate)
	}

	listeners, err := parser.dao.FindListenersByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to load listeners for node group %s using DAO:\n %v", nodeGroup, err)
	}
	for _, listener := range listeners {
		granularUpdate := parser.updateActionFactory.ListenerUpdate(nodeGroup, entityVersions[domain.ListenerTable], listener)
		actions.Put(action.EnvoyListener, &granularUpdate)
	}
}

func (builder *compositeUpdateBuilder) processNodeGroupChanges(changes []memdb.Change) {
	for _, change := range changes {
		if !change.Deleted() {
			nodeGroup, ok := change.After.(*domain.NodeGroup)
			if !ok {
				logger.Panicf("compositeUpdateBuilder#processNodeGroupChanges failed to cast %v to NodeGroup", change.After)
			}

			if err := builder.envoyConfigBuilder.RegisterGateway(nodeGroup); err != nil {
				logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to register gateway %+v in builder using DAO:\n %v", nodeGroup, err)
			}

			if change.Before == nil {
				builder.updateRuntimeConfigs(nodeGroup)
			}

			if change.Before == nil || change.Before.(*domain.NodeGroup).GatewayType != nodeGroup.GatewayType {
				// only gateway type change should trigger envoy route config and listener update
				builder.updateListenersAndRouteConfigs(nodeGroup.Name)
			}
		}
	}
}

func (builder *compositeUpdateBuilder) updateListenersAndRouteConfigs(nodeGroup string) {
	routeConfigs, err := builder.repo.FindRouteConfigsByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to load route configs for node group %s using DAO:\n %v", nodeGroup, err)
	}
	for _, routeConfig := range routeConfigs {
		builder.addUpdateAction(nodeGroup, action.EnvoyRouteConfig, routeConfig)
	}

	listeners, err := builder.repo.FindListenersByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Panicf("changeEventParserImpl#processNodeGroupChanges failed to load listeners for node group %s using DAO:\n %v", nodeGroup, err)
	}
	for _, listener := range listeners {
		builder.addUpdateAction(nodeGroup, action.EnvoyListener, listener)
	}
}
