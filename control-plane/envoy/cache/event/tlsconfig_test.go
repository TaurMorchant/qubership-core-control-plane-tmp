package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	"github.com/netcracker/qubership-core-control-plane/services/provider"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	mock_provider "github.com/netcracker/qubership-core-control-plane/test/mock/services/provider"
	"testing"
)

func TestChangeEventParser_processTlsConfigChanges(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	nodeGroup := "nodeGroup"
	tlsService := mock_provider.NewMockTlsService(ctrl)
	tlsService.EXPECT().GetGlobalTlsConfigs(gomock.Any(), nodeGroup).AnyTimes().Return(nil, nil)
	provider.Init(tlsService)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{
		domain.ClusterTable: "test",
	}
	tlsConfig1 := &domain.TlsConfig{
		Id: int32(1),
	}
	tlsConfig2 := &domain.TlsConfig{
		Id: int32(2),
	}
	changes := []memdb.Change{
		{
			After: &domain.TlsConfig{
				Id: tlsConfig1.Id,
			},
		},
		{
			Before: &domain.TlsConfig{
				Id: tlsConfig2.Id,
			},
		},
	}

	clusters := []*domain.Cluster{
		{
			TLSId: tlsConfig1.Id,
		},
		{
			TLSId: tlsConfig2.Id,
		},
	}
	granularUpdate := action.GranularEntityUpdate{}

	mockDao.EXPECT().FindAllClusters().Return(clusters, nil).Times(2)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate)
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[1]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate).Times(2)

	changeEventParser.processTlsConfigChanges(actions, entityVersions, nodeGroup, changes)
}
