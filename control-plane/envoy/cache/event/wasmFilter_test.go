package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"testing"
)

func TestCompositeUpdateBuilder_processListenersWasmFilterChangesprocessWasmFilterChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroupId1 := "nodeId1"
	nodeGroupId2 := "nodeId2"
	versionsByNodeGroup := nodeGroupEntityVersions{
		nodeGroupId1: {
			action.EnvoyListener: "test1",
		},
		nodeGroupId2: {
			action.EnvoyListener: "test2",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, versionsByNodeGroup, mockBuilder, mockUpdateAction)

	listeners := []*domain.Listener{
		{
			Id:          int32(1),
			Name:        "listener1",
			NodeGroupId: nodeGroupId1,
		},
		{
			Id:          int32(2),
			Name:        "listener2",
			NodeGroupId: nodeGroupId2,
		},
	}

	changes := []memdb.Change{
		{
			Before: &domain.ListenersWasmFilter{
				ListenerId: listeners[0].Id,
			},
			After: &domain.ListenersWasmFilter{
				ListenerId: listeners[1].Id,
			},
		},
	}

	mockDao.EXPECT().FindListenerById(changes[0].After.(*domain.ListenersWasmFilter).ListenerId).Return(listeners[0], nil)
	mockUpdateAction.EXPECT().ListenerUpdate(nodeGroupId1, versionsByNodeGroup[nodeGroupId1][action.EnvoyListener], gomock.Any())

	compositeUpdateBuilder.processListenersWasmFilterChanges(changes)
}

func TestCompositeUpdateBuilder_processWasmFilterChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroupId1 := "nodeId1"
	nodeGroupId2 := "nodeId2"
	versionsByNodeGroup := nodeGroupEntityVersions{
		nodeGroupId1: {
			action.EnvoyListener: "test1",
		},
		nodeGroupId2: {
			action.EnvoyListener: "test2",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, versionsByNodeGroup, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			Before: &domain.WasmFilter{
				Id:   int32(1),
				Name: "before",
			},
			After: &domain.WasmFilter{
				Id:   int32(2),
				Name: "after",
			},
		},
	}

	wasmFilterIds := []int32{int32(1), int32(2)}
	mockDao.EXPECT().FindListenerIdsByWasmFilterId(changes[0].After.(*domain.WasmFilter).Id).Return(wasmFilterIds, nil)

	listeners := []*domain.Listener{
		{
			Id:          int32(1),
			Name:        "listener1",
			NodeGroupId: nodeGroupId1,
		},
		{
			Id:          int32(2),
			Name:        "listener2",
			NodeGroupId: nodeGroupId2,
		},
	}
	mockDao.EXPECT().FindListenerById(wasmFilterIds[0]).Return(listeners[0], nil)
	mockDao.EXPECT().FindListenerById(wasmFilterIds[1]).Return(listeners[1], nil)
	mockUpdateAction.EXPECT().ListenerUpdate(nodeGroupId1, versionsByNodeGroup[nodeGroupId1][action.EnvoyListener], gomock.Any())
	mockUpdateAction.EXPECT().ListenerUpdate(nodeGroupId2, versionsByNodeGroup[nodeGroupId2][action.EnvoyListener], gomock.Any())

	compositeUpdateBuilder.processWasmFilterChanges(changes)
}

func TestChangeEventParser_processWasmFilterChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{
		domain.ListenerTable: "test",
	}
	nodeGroup := "nodeGroup"
	listeners := []*domain.Listener{{}}

	mockDao.EXPECT().FindListenersByNodeGroupId(nodeGroup).Return(listeners, nil)
	granularEntityUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().ListenerUpdate(nodeGroup, entityVersions[domain.ListenerTable], listeners[0]).Return(granularEntityUpdate)
	actions.EXPECT().Put(action.EnvoyListener, &granularEntityUpdate)

	changeEventParser.processWasmFilterChanges(actions, entityVersions, nodeGroup)
}
