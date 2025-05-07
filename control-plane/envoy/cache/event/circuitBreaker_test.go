package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_processCircuitBreakerChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.CircuitBreakerTable: "test2"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			Before: &domain.CircuitBreaker{
				Id:          1,
				ThresholdId: 2,
			},
		},
		{
			After: &domain.CircuitBreaker{
				Id:          1,
				ThresholdId: 0,
			},
		},
	}

	clusters := getClusters(int32(1), int32(2))
	clusters[0].CircuitBreakerId = 1

	granularUpdate := action.GranularEntityUpdate{}
	mockDao.EXPECT().FindAllClusters().Return(clusters, nil).Times(2)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate).Times(2)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate).Times(2)

	changeEventParser.processCircuitBreakerChanges(actions, entityVersions, nodeGroup, changes)
}
