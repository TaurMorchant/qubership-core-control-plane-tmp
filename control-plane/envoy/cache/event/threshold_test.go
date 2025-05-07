package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_processThresholdChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.ThresholdTable: "test2"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			Before: &domain.Threshold{
				Id:             1,
				MaxConnections: 2,
			},
		},
		{
			After: &domain.Threshold{
				Id:             1,
				MaxConnections: 1,
			},
		},
	}

	clusters := getClusters(int32(1), int32(2))
	clusters[0].CircuitBreakerId = 4
	circuitBreakers := getCircuitBreakers(int32(4))
	circuitBreakers[0].ThresholdId = 1

	granularUpdate := action.GranularEntityUpdate{}
	mockDao.EXPECT().FindAllClusters().Return(clusters, nil).Times(2)
	mockDao.EXPECT().FindAllCircuitBreakers().Return(circuitBreakers, nil).Times(2)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate).Times(2)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate).Times(2)

	changeEventParser.processThresholdChanges(actions, entityVersions, nodeGroup, changes)
}
