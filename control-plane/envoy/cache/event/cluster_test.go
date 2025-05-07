package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_processClusterNodeGroupChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.ClusterTable: "test2"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			Before: &domain.ClustersNodeGroup{
				NodegroupsName: nodeGroup,
				ClustersId:     int32(1),
			},
		},
		{
			After: &domain.ClustersNodeGroup{
				NodegroupsName: nodeGroup,
				ClustersId:     int32(2),
			},
		},
	}

	clusters := getAndExpectClusters(mockDao, int32(1), int32(2))

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().ClusterDelete(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate).Times(2)

	changeEventParser.processClusterNodeGroupChanges(actions, entityVersions, nodeGroup, changes)
}

func TestCompositeUpdateBuilder_processClusterNodeGroupChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		"1": {
			action.EnvoyCluster: "test1",
		},
		"2": {
			action.EnvoyListener: "test2",
		},
	}
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			Before: &domain.ClustersNodeGroup{
				NodegroupsName: nodeGroup,
				ClustersId:     int32(1),
			},
		},
		{
			After: &domain.ClustersNodeGroup{
				NodegroupsName: nodeGroup,
				ClustersId:     int32(2),
			},
		},
	}

	mockDao.EXPECT().FindClusterById(int32(2)).Return(&domain.Cluster{Id: int32(2), Name: "test-cluster||test-cluster||8080"}, nil)
	compositeUpdateBuilder.processClusterNodeGroupChanges(changes)
}
