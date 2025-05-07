package event

import (
	v3runtime "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParsePartialReloadEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	nodeGroup := "nodeGroup1"
	version := int64(1)
	changeEvent := &events.PartialReloadEvent{
		EnvoyVersions: []*domain.EnvoyConfigVersion{
			{
				NodeGroup:  nodeGroup,
				EntityType: domain.ClusterTable,
				Version:    version,
			},
		},
	}

	clusters := []*domain.Cluster{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindClusterByNodeGroup(&domain.NodeGroup{Name: nodeGroup}).Return(clusters, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, "1", clusters[0]).Return(granularUpdate)

	updateActionMap := changeEventParser.ParsePartialReloadEvent(changeEvent)
	assert.NotNil(t, updateActionMap[nodeGroup])
}

func TestParseChangeEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	//mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, action.NewUpdateActionFactory(mockBuilder), mockBuilder)

	nodeGroup := "nodeGroup1"
	changeEvent := &events.ChangeEvent{
		NodeGroup: nodeGroup,
		Changes: map[string][]memdb.Change{
			domain.NodeGroupTable: {
				{
					After: &domain.NodeGroup{
						Name: nodeGroup,
					},
					Before: nil,
				},
			},
			domain.EnvoyConfigVersionTable: {
				{
					After: &domain.EnvoyConfigVersion{
						NodeGroup:  nodeGroup,
						EntityType: domain.ClusterTable,
						Version:    int64(1),
					},
				},
				{
					After: &domain.EnvoyConfigVersion{
						NodeGroup:  nodeGroup,
						EntityType: domain.NodeGroupTable,
						Version:    int64(2),
					},
				},
			},
		},
	}

	mockBuilder.EXPECT().RegisterGateway(changeEvent.Changes[domain.NodeGroupTable][0].After).Return(nil)

	mockDao.EXPECT().FindRouteConfigsByNodeGroupId(nodeGroup).Return([]*domain.RouteConfiguration{}, nil)
	mockDao.EXPECT().FindListenersByNodeGroupId(nodeGroup).Return([]*domain.Listener{}, nil)

	updateActionMap := changeEventParser.ParseChangeEvent(changeEvent)
	assert.NotNil(t, updateActionMap)

	actions := updateActionMap.CompositeAction()
	assert.NotNil(t, actions)

	originalSnapshot := &cache.Snapshot{}
	version := "2"
	originalSnapshot.Resources[types.Runtime].Version = version
	originalSnapshot.Resources[types.Runtime].Items = map[string]types.ResourceWithTTL{}

	runtimeName := "rtds_layer0"
	res := &v3runtime.Runtime{}
	mockBuilder.EXPECT().BuildRuntime(nodeGroup, runtimeName).Return(res, nil)

	result, err := actions.Perform(originalSnapshot)
	assert.Nil(t, err)
	assert.NotEqual(t, result.Resources[types.Runtime].Version, version)
	assert.Equal(t, result.Resources[types.Runtime].Items[runtimeName].Resource, res)
}

func TestParseMultipleChangeEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	//mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, action.NewUpdateActionFactory(mockBuilder), mockBuilder)

	nodeGroup := "nodeGroup1"
	changeEvent := &events.MultipleChangeEvent{
		Changes: map[string][]memdb.Change{
			domain.NodeGroupTable: {
				{
					After: &domain.NodeGroup{
						Name: nodeGroup,
					},
					Before: nil,
				},
			},
			domain.EnvoyConfigVersionTable: {
				{
					After: &domain.EnvoyConfigVersion{
						NodeGroup:  nodeGroup,
						EntityType: domain.ClusterTable,
						Version:    int64(1),
					},
				},
				{
					After: &domain.EnvoyConfigVersion{
						NodeGroup:  nodeGroup,
						EntityType: domain.NodeGroupTable,
						Version:    int64(2),
					},
				},
			},
		},
	}

	mockBuilder.EXPECT().RegisterGateway(changeEvent.Changes[domain.NodeGroupTable][0].After).Return(nil)

	mockDao.EXPECT().FindRouteConfigsByNodeGroupId(nodeGroup).Return([]*domain.RouteConfiguration{}, nil)
	mockDao.EXPECT().FindListenersByNodeGroupId(nodeGroup).Return([]*domain.Listener{}, nil)

	snapshotUpdateActionMap := changeEventParser.ParseMultipleChangeEvent(changeEvent)
	assert.NotNil(t, snapshotUpdateActionMap)

	actions := snapshotUpdateActionMap[nodeGroup]
	assert.NotNil(t, actions)

	originalSnapshot := &cache.Snapshot{}
	version := "2"
	originalSnapshot.Resources[types.Runtime].Version = version
	originalSnapshot.Resources[types.Runtime].Items = map[string]types.ResourceWithTTL{}

	runtimeName := "rtds_layer0"
	res := &v3runtime.Runtime{}
	mockBuilder.EXPECT().BuildRuntime(nodeGroup, runtimeName).Return(res, nil)

	result, err := actions.Perform(originalSnapshot)
	assert.Nil(t, err)
	assert.NotEqual(t, result.Resources[types.Runtime].Version, version)
	assert.Equal(t, result.Resources[types.Runtime].Items[runtimeName].Resource, res)
}
