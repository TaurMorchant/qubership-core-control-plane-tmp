package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestCompositeUpdateBuilder_processHeaderMatcherChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupId1 := "1"
	nodeGroupId2 := "2"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroupId1: {
			action.EnvoyRouteConfig: "test1",
		},
		nodeGroupId2: {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.HeaderMatcher{
				Id:      int32(1),
				RouteId: int32(1),
			},
		},
		{
			Before: &domain.HeaderMatcher{
				Id:      int32(2),
				RouteId: int32(2),
			},
		},
	}
	routes := getAndExpectRoutesWithVirtualHostId(mockDao, changes[0].After.(*domain.HeaderMatcher).RouteId, changes[1].Before.(*domain.HeaderMatcher).RouteId)
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, routes[0].VirtualHostId, routes[1].VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId1, nodeGroupEntityVersions[routeConfigurations[0].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId2, nodeGroupEntityVersions[routeConfigurations[1].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[1]).Return(granularUpdate)

	compositeUpdateBuilder.processHeaderMatcherChanges(changes)
}

func TestCompositeUpdateBuilder_processRouteChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupId1 := "1"
	nodeGroupId2 := "2"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroupId1: {
			action.EnvoyRouteConfig: "test1",
		},
		nodeGroupId2: {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.Route{
				Id:            int32(1),
				VirtualHostId: int32(1),
			},
		},
		{
			Before: &domain.Route{
				Id:            int32(2),
				VirtualHostId: int32(2),
			},
		},
	}
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, changes[0].After.(*domain.Route).VirtualHostId, changes[1].Before.(*domain.Route).VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId1, nodeGroupEntityVersions[routeConfigurations[0].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId2, nodeGroupEntityVersions[routeConfigurations[1].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[1]).Return(granularUpdate)

	compositeUpdateBuilder.processRouteChanges(changes)
}

func TestChangeEventParserImpl_processHeaderMatcherChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.HeaderMatcher{
				Id:      int32(1),
				RouteId: int32(1),
			},
		},
		{
			Before: &domain.HeaderMatcher{
				Id:      int32(2),
				RouteId: int32(2),
			},
		},
	}

	routes := getAndExpectRoutesWithVirtualHostId(mockDao, changes[0].After.(*domain.HeaderMatcher).RouteId, changes[1].Before.(*domain.HeaderMatcher).RouteId)
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, routes[0].VirtualHostId, routes[1].VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(2)

	changeEventParser.processHeaderMatcherChanges(actions, entityVersions, nodeGroup, changes)
}

func TestChangeEventParserImpl_processRouteChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.Route{
				Id:            int32(1),
				VirtualHostId: int32(1),
			},
		},
		{
			Before: &domain.Route{
				Id:            int32(2),
				VirtualHostId: int32(2),
			},
		},
	}

	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, changes[0].After.(*domain.Route).VirtualHostId, changes[1].Before.(*domain.Route).VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(2)

	changeEventParser.processRouteChanges(actions, entityVersions, nodeGroup, changes)
}
