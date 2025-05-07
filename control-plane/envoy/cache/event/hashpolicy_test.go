package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_processHashPolicyChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1", domain.ClusterTable: "test2"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.HashPolicy{
				Id:         int32(1),
				RouteId:    int32(1),
				EndpointId: int32(1),
			},
		},
		{
			Before: &domain.HashPolicy{
				Id:         int32(2),
				RouteId:    int32(2),
				EndpointId: int32(2),
			},
		},
	}

	routes := getAndExpectRoutesWithVirtualHostId(mockDao, changes[0].After.(*domain.HashPolicy).RouteId, changes[1].Before.(*domain.HashPolicy).RouteId)
	endpoints := getAndExpectEndpointsWithClusterId(mockDao, changes[0].After.(*domain.HashPolicy).EndpointId, changes[1].Before.(*domain.HashPolicy).EndpointId)
	virtualHosts := getAndExpectVirtualHostWithRouteConfigurationId(mockDao, routes[0].VirtualHostId, routes[1].VirtualHostId)
	routeConfigurations := getAndExpectRouteConfigurationWithNodeGroupId(mockDao, virtualHosts[0].RouteConfigurationId, virtualHosts[1].RouteConfigurationId)

	mockDao.EXPECT().FindRouteConfigsByEndpoint(endpoints[0]).Return([]*domain.RouteConfiguration{routeConfigurations[0]}, nil)
	mockDao.EXPECT().FindRouteConfigsByEndpoint(endpoints[1]).Return([]*domain.RouteConfiguration{routeConfigurations[1]}, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[0]).Times(2).Return(granularUpdate)
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigurations[1]).Times(2).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(4)

	clusters := getAndExpectClusters(mockDao, int32(1), int32(2))

	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate).Times(2)

	changeEventParser.processHashPolicyChanges(actions, entityVersions, nodeGroup, changes)
}
