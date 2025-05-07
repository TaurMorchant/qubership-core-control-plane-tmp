package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestCompositeUpdateBuilder_updateRouteConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroupId := "NodeGroupId1"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroupId: {
			action.EnvoyRouteConfig: "test1",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	routeConfigId := int32(1)
	routeConfig := &domain.RouteConfiguration{
		Id:          routeConfigId,
		NodeGroupId: nodeGroupId,
	}
	mockDao.EXPECT().FindRouteConfigById(routeConfigId).Return(routeConfig, nil)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId, "test1", routeConfig)

	compositeUpdateBuilder.updateRouteConfig(routeConfigId)
}

func TestCompositeUpdateBuilder_processRouteConfigurationChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroupId1 := "NodeGroupId1"
	nodeGroupId2 := "NodeGroupId2"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroupId1: {
			action.EnvoyRouteConfig: "test1",
		},
		nodeGroupId2: {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.RouteConfiguration{
				Id:          int32(2),
				Name:        "after",
				NodeGroupId: "NodeGroupId1",
			},
		},
		{
			Before: &domain.RouteConfiguration{
				Id:          int32(1),
				Name:        "before",
				NodeGroupId: "NodeGroupId2",
			},
		},
	}

	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroupId1, "test1", changes[0].After)
	mockUpdateAction.EXPECT().RouteConfigDelete(nodeGroupId2, "test2", changes[1].Before)

	compositeUpdateBuilder.processRouteConfigurationChanges(changes)
}

func TestChangeEventParser_UpdateRouteConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParserImpl := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	routeConfig := &domain.RouteConfiguration{}
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate)

	changeEventParserImpl.updateRouteConfig(actions, entityVersions, nodeGroup, routeConfig)
}

func TestChangeEventParser_ProcessRouteConfigurationChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParserImpl := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	changes := []memdb.Change{
		{
			After: &domain.RouteConfiguration{
				Id:   int32(2),
				Name: "after",
			},
		},
		{
			Before: &domain.RouteConfiguration{
				Id:   int32(1),
				Name: "before",
			},
		},
	}

	granularUpdate := action.GranularEntityUpdate{}
	granularDelete := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate("", entityVersions[domain.RouteConfigurationTable], changes[0].After).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigDelete("", entityVersions[domain.RouteConfigurationTable], changes[1].Before).Return(granularDelete)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularDelete)

	changeEventParserImpl.processRouteConfigurationChanges(actions, entityVersions, "", changes)
}
