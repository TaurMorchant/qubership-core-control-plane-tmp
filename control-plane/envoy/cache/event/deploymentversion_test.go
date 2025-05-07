package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_processDeploymentVersionChanges(t *testing.T) {
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
			After: &domain.DeploymentVersion{
				Version: "v1",
				Stage:   domain.ActiveStage,
			},
		},
	}

	endpoints := []*domain.Endpoint{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindEndpointsByDeploymentVersion(changes[0].After.(*domain.DeploymentVersion).Version).Return(endpoints, nil)

	clusters := []*domain.Cluster{
		{
			Id:   int32(1),
			Name: "cluster1",
		},
	}
	mockDao.EXPECT().FindClusterByEndpointIn(endpoints).Return(clusters, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().ClusterUpdate(nodeGroup, entityVersions[domain.ClusterTable], clusters[0]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyCluster, &granularUpdate)

	routeConfigs := []*domain.RouteConfiguration{
		{
			Id: int32(1),
		},
	}
	mockDao.EXPECT().FindRouteConfigsByRouteDeploymentVersion(changes[0].After.(*domain.DeploymentVersion).Version).Return(routeConfigs, nil)

	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfigs[0]).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate)

	changeEventParser.processDeploymentVersionChanges(actions, entityVersions, nodeGroup, changes)
}
