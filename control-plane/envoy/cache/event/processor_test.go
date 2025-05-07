package event

import (
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompositeUpdateBuilder_withReloadForVersions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		"1": {
			action.EnvoyCluster: "test1",
		},
		"2": {
			action.EnvoyListener: "test2",
		},
		"3": {
			action.EnvoyRouteConfig: "test3",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	granularUpdate := action.GranularEntityUpdate{}
	clusters := []*domain.Cluster{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindClusterByNodeGroup(&domain.NodeGroup{Name: "1"}).Return(clusters, nil)
	mockUpdateAction.EXPECT().ClusterUpdate("1", "test1", clusters[0]).Return(granularUpdate)

	listeners := []*domain.Listener{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindListenersByNodeGroupId("2").Return(listeners, nil)
	mockUpdateAction.EXPECT().ListenerUpdate("2", "test2", listeners[0]).Return(granularUpdate)

	routeConfigs := []*domain.RouteConfiguration{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindRouteConfigsByNodeGroupId("3").Return(routeConfigs, nil)
	mockUpdateAction.EXPECT().RouteConfigUpdate("3", "test3", routeConfigs[0]).Return(granularUpdate)

	compositeUpdateBuilder.withReloadForVersions()
}

func TestNodeGroupEntityVersions_Put(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	envoyConfigVersion := &domain.EnvoyConfigVersion{
		NodeGroup:  "nodeGroup",
		EntityType: domain.ClusterTable,
		Version:    int64(1),
	}
	versionsMap := make(nodeGroupEntityVersions, 2)
	versionsMap.put(envoyConfigVersion)

	assert.Equal(t, "1", versionsMap[envoyConfigVersion.NodeGroup][action.EnvoyCluster])
}

func TestAddBeforeAction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		"1": {
			action.EnvoyRouteConfig: "test1",
		},
		"2": {
			action.EnvoyRouteConfig: "test2",
		},
	}
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	nodeGroup := "nodeGroup"
	mockFirstBeforeAction := getMockSnapshotUpdateAction(ctrl)
	compositeUpdateBuilder.addBeforeAction(nodeGroup, mockFirstBeforeAction)

	resultActions := compositeUpdateBuilder.nodeGroupBeforeActions[nodeGroup]
	assert.Equal(t, 1, len(resultActions))
	assert.Equal(t, mockFirstBeforeAction, resultActions[0])

	mockSecondBeforeAction := &action.CompositeUpdateAction{}
	compositeUpdateBuilder.addBeforeAction(nodeGroup, mockSecondBeforeAction)

	resultActions = compositeUpdateBuilder.nodeGroupBeforeActions[nodeGroup]
	assert.Equal(t, 1, len(resultActions))
	assert.Equal(t, mockFirstBeforeAction, resultActions[0])
}
