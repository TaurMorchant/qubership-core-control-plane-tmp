package action

import (
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/builder"
)

//go:generate mockgen -source=factory.go -destination=../../../test/mock/envoy/cache/action/stub_factory.go -package=mock_action
type UpdateActionFactory interface {
	ClusterUpdateAction(nodeGroup, version string, entity interface{}) SnapshotUpdateAction
	ClusterDeleteAction(version string, entity interface{}) SnapshotUpdateAction
	ListenerUpdateAction(version string, entity interface{}) SnapshotUpdateAction
	ListenerDeleteAction(version string, entity interface{}) SnapshotUpdateAction
	RouteConfigUpdateAction(version string, entity interface{}) SnapshotUpdateAction
	RouteConfigDeleteAction(version string, entity interface{}) SnapshotUpdateAction
	RuntimeUpdateAction(nodeGroup string) SnapshotUpdateAction
	DeleteAllClustersAction(version string) SnapshotUpdateAction
	DeleteAllListenersAction(version string) SnapshotUpdateAction
	DeleteAllRouteConfigsAction(version string) SnapshotUpdateAction
	ClusterUpdate(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	ClusterDelete(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	ListenerUpdate(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	ListenerDelete(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	RouteConfigUpdate(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	RouteConfigDelete(nodeGroup, version string, entity interface{}) GranularEntityUpdate
	RuntimeUpdate(nodeGroup, version string, entity interface{}) GranularEntityUpdate
}

type updateActionFactoryImpl struct {
	envoyConfigBuilder builder.EnvoyConfigBuilder
}

func NewUpdateActionFactory(envoyConfigBuilder builder.EnvoyConfigBuilder) *updateActionFactoryImpl {
	return &updateActionFactoryImpl{envoyConfigBuilder: envoyConfigBuilder}
}

func (factory *updateActionFactoryImpl) ClusterUpdateAction(nodeGroup, version string, entity interface{}) SnapshotUpdateAction {
	return NewClusterUpdateAction(factory.envoyConfigBuilder, nodeGroup, version, entity.(*domain.Cluster))
}

func (factory *updateActionFactoryImpl) ClusterDeleteAction(version string, entity interface{}) SnapshotUpdateAction {
	return NewClusterDeleteAction(version, entity.(*domain.Cluster))
}

func (factory *updateActionFactoryImpl) ListenerUpdateAction(version string, entity interface{}) SnapshotUpdateAction {
	return NewListenerUpdateAction(factory.envoyConfigBuilder, version, entity.(*domain.Listener), "")
}

func (factory *updateActionFactoryImpl) ListenerDeleteAction(version string, entity interface{}) SnapshotUpdateAction {
	return NewListenerDeleteAction(version, entity.(*domain.Listener))
}

func (factory *updateActionFactoryImpl) RouteConfigUpdateAction(version string, entity interface{}) SnapshotUpdateAction {
	return NewRouteConfigUpdateAction(factory.envoyConfigBuilder, version, entity.(*domain.RouteConfiguration))
}

func (factory *updateActionFactoryImpl) RouteConfigDeleteAction(version string, entity interface{}) SnapshotUpdateAction {
	return NewRouteConfigDeleteAction(version, entity.(*domain.RouteConfiguration))
}

func (factory *updateActionFactoryImpl) RuntimeUpdateAction(nodeGroup string) SnapshotUpdateAction {
	return NewRuntimeUpdateAction(factory.envoyConfigBuilder, nodeGroup)
}

func (factory *updateActionFactoryImpl) DeleteAllClustersAction(version string) SnapshotUpdateAction {
	return NewDeleteAllByTypeAction(version, resource.ClusterType)
}

func (factory *updateActionFactoryImpl) DeleteAllListenersAction(version string) SnapshotUpdateAction {
	return NewDeleteAllByTypeAction(version, resource.ListenerType)
}

func (factory *updateActionFactoryImpl) DeleteAllRouteConfigsAction(version string) SnapshotUpdateAction {
	return NewDeleteAllByTypeAction(version, resource.RouteType)
}

func (factory *updateActionFactoryImpl) ClusterUpdate(nodeGroup, version string, entity interface{}) GranularEntityUpdate {
	cluster := entity.(*domain.Cluster)
	action := NewClusterUpdateAction(factory.envoyConfigBuilder, nodeGroup, version, cluster)
	return GranularEntityUpdate{Action: action, IsDelete: false, EntityId: cluster.Id}
}

func (factory *updateActionFactoryImpl) ClusterDelete(_, version string, entity interface{}) GranularEntityUpdate {
	cluster := entity.(*domain.Cluster)
	action := NewClusterDeleteAction(version, cluster)
	return GranularEntityUpdate{Action: action, IsDelete: true, EntityId: cluster.Id}
}

func (factory *updateActionFactoryImpl) ListenerUpdate(_, version string, entity interface{}) GranularEntityUpdate {
	listener := entity.(*domain.Listener)
	action := NewListenerUpdateAction(factory.envoyConfigBuilder, version, listener, "")
	return GranularEntityUpdate{Action: action, IsDelete: false, EntityId: listener.Id}
}

func (factory *updateActionFactoryImpl) RuntimeUpdate(nodeGroupName, version string, nodeGroup interface{}) GranularEntityUpdate {
	action := NewRuntimeUpdateAction(factory.envoyConfigBuilder, nodeGroupName)
	return GranularEntityUpdate{Action: action, IsDelete: false, EntityId: nodeGroup.(*domain.NodeGroup).GetId()}
}

func (factory *updateActionFactoryImpl) ListenerDelete(_, version string, entity interface{}) GranularEntityUpdate {
	listener := entity.(*domain.Listener)
	action := NewListenerDeleteAction(version, listener)
	return GranularEntityUpdate{Action: action, IsDelete: true, EntityId: listener.Id}
}

func (factory *updateActionFactoryImpl) RouteConfigUpdate(_, version string, entity interface{}) GranularEntityUpdate {
	routeConfig := entity.(*domain.RouteConfiguration)
	action := NewRouteConfigUpdateAction(factory.envoyConfigBuilder, version, routeConfig)
	return GranularEntityUpdate{Action: action, IsDelete: false, EntityId: routeConfig.Id}
}

func (factory *updateActionFactoryImpl) RouteConfigDelete(_, version string, entity interface{}) GranularEntityUpdate {
	routeConfig := entity.(*domain.RouteConfiguration)
	action := NewRouteConfigDeleteAction(version, routeConfig)
	return GranularEntityUpdate{Action: action, IsDelete: true, EntityId: routeConfig.Id}
}

// GranularEntityUpdate structure is used for storing SnapshotUpdateActions in updateActionsMap,
// binding operation type (save or delete) and entity id (e.g. Cluster.Id) to them.
type GranularEntityUpdate struct {
	Action   SnapshotUpdateAction
	IsDelete bool
	// EntityId identifies entity in the envoy configuration snapshot cache.
	EntityId int32
}

type EnvoyEntity int

const (
	EnvoyCluster EnvoyEntity = iota
	EnvoyListener
	EnvoyRouteConfig
	EnvoyRuntime
)

func (e EnvoyEntity) ToTypeURL() resource.Type {
	switch e {
	case EnvoyCluster:
		return resource.ClusterType
	case EnvoyListener:
		return resource.ListenerType
	case EnvoyRouteConfig:
		return resource.RouteType
	default:
		logger.Panicf("Cannot resolve envoy entity typeURL by EnvoyEntity %v", e)
		return ""
	}
}

func EnvoyEntityByTable(tableName string) EnvoyEntity {
	switch tableName {
	case domain.ClusterTable:
		return EnvoyCluster
	case domain.RouteConfigurationTable:
		return EnvoyRouteConfig
	case domain.ListenerTable:
		return EnvoyListener
	case domain.NodeGroupTable:
		return EnvoyRuntime
	}
	logger.Panicf("Cannot resolve envoy entity type by table name %s", tableName)
	return 0
}

type ActionsMap interface {
	Put(entityType EnvoyEntity, granularUpdate *GranularEntityUpdate)
	CompositeAction() *CompositeUpdateAction
}

type updateActionsMap struct {
	actions map[EnvoyEntity]map[int32]*GranularEntityUpdate
}

func NewUpdateActionsMap() *updateActionsMap {
	return &updateActionsMap{actions: make(map[EnvoyEntity]map[int32]*GranularEntityUpdate)}
}

func (actionsMap *updateActionsMap) Put(entityType EnvoyEntity, granularUpdate *GranularEntityUpdate) {
	if existingUpdates, found := actionsMap.actions[entityType]; found {
		if existingUpdate, exists := existingUpdates[granularUpdate.EntityId]; exists {
			if !existingUpdate.IsDelete && granularUpdate.IsDelete {
				actionsMap.actions[entityType][granularUpdate.EntityId] = granularUpdate
			}
		} else {
			actionsMap.actions[entityType][granularUpdate.EntityId] = granularUpdate
		}
	} else {
		newUpdatesMap := make(map[int32]*GranularEntityUpdate)
		newUpdatesMap[granularUpdate.EntityId] = granularUpdate
		actionsMap.actions[entityType] = newUpdatesMap
	}
}

func (actionsMap *updateActionsMap) CompositeAction() *CompositeUpdateAction {
	actionsToPerform := make([]SnapshotUpdateAction, 0)
	for _, actionsByType := range actionsMap.actions {
		for _, action := range actionsByType {
			actionsToPerform = append(actionsToPerform, action.Action)
		}
	}
	return NewCompositeUpdateAction(actionsToPerform)
}
