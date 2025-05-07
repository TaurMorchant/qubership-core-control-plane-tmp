package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"testing"
)

func TestCompositeUpdateBuilder_processVirtualHostDomainChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupId1 := "1"
	nodeGroupId2 := "2"
	entityVersions := nodeGroupEntityVersions{
		nodeGroupId1: {
			action.EnvoyRouteConfig: "test1",
		},
		nodeGroupId2: {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, entityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.VirtualHostDomain{
				VirtualHostId: int32(1),
			},
		},
		{
			Before: &domain.VirtualHostDomain{
				VirtualHostId: int32(2),
			},
		},
	}
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, changes[0].After.(*domain.VirtualHostDomain).VirtualHostId, changes[1].Before.(*domain.VirtualHostDomain).VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId1, entityVersions[routeConfigurations[0].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId2, entityVersions[routeConfigurations[1].NodeGroupId][action.EnvoyRouteConfig], routeConfigurations[1]).Return(granularUpdate)

	compositeUpdateBuilder.processVirtualHostDomainChanges(changes)
}

func TestCompositeUpdateBuilder_processVirtualHostChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupId1 := "1"
	nodeGroupId2 := "2"
	entityVersions := nodeGroupEntityVersions{
		nodeGroupId1: {
			action.EnvoyRouteConfig: "test1",
		},
		nodeGroupId2: {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, entityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.VirtualHost{
				RouteConfigurationId: int32(1),
			},
		},
		{
			Before: &domain.VirtualHost{
				RouteConfigurationId: int32(2),
			},
		},
	}
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, changes[0].After.(*domain.VirtualHost).RouteConfigurationId, changes[1].Before.(*domain.VirtualHost).RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId1, entityVersions[nodeGroupId1][action.EnvoyRouteConfig], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId2, entityVersions[nodeGroupId2][action.EnvoyRouteConfig], routeConfigurations[1]).Return(granularUpdate)

	compositeUpdateBuilder.processVirtualHostChanges(changes)
}

func TestChangeEventParser_processVirtualHostDomainChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{
		domain.RouteConfigurationTable: "test",
	}
	nodeGroup := "nodeGroup1"
	changes := []memdb.Change{
		{
			After: &domain.VirtualHostDomain{
				VirtualHostId: int32(1),
			},
		},
		{
			Before: &domain.VirtualHostDomain{
				VirtualHostId: int32(2),
			},
		},
	}
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, changes[0].After.(*domain.VirtualHostDomain).VirtualHostId, changes[1].Before.(*domain.VirtualHostDomain).VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(2)

	changeEventParser.processVirtualHostDomainChanges(actions, entityVersions, nodeGroup, changes)
}

func TestChangeEventParser_processVirtualHostChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{
		domain.RouteConfigurationTable: "test",
	}
	nodeGroup := "nodeGroup1"
	changes := []memdb.Change{
		{
			After: &domain.VirtualHost{
				Id:                   int32(1),
				Name:                 "after",
				RouteConfigurationId: int32(1),
			},
		},
		{
			Before: &domain.VirtualHost{
				Id:                   int32(2),
				Name:                 "before",
				RouteConfigurationId: int32(2),
			},
		},
	}
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, changes[0].After.(*domain.VirtualHost).RouteConfigurationId, changes[1].Before.(*domain.VirtualHost).RouteConfigurationId)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(2)

	changeEventParser.processVirtualHostChanges(actions, entityVersions, nodeGroup, changes)
}
