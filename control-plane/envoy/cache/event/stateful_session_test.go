package event

import (
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	mock_builder "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/builder"
	"testing"
)

func TestChangeEventParserImpl_ProcessStatefulSessionChanges_Route(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(&domain.Route{Id: int32(1), VirtualHostId: int32(1)}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig).Times(1).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(1)

	changeEventParser.processStatefulSessionChanges(actions, entityVersions, nodeGroup, changes)
}

func TestChangeEventParserImpl_ProcessStatefulSessionChanges_Endpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(&domain.Endpoint{Id: int32(1), ClusterId: int32(1), StatefulSessionId: int32(1), DeploymentVersion: "v1"}, nil)
	mockDao.EXPECT().FindClusterById(int32(1)).Return(&domain.Cluster{Id: int32(1), Name: "test-cluster||test-cluster||8080"}, nil)
	mockDao.EXPECT().FindRoutesByClusterNameAndDeploymentVersion("test-cluster||test-cluster||8080", "v1").
		Return([]*domain.Route{{Id: int32(1), DeploymentVersion: "v1", VirtualHostId: int32(1)}}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig).Times(1).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(1)

	changeEventParser.processStatefulSessionChanges(actions, entityVersions, nodeGroup, changes)
}

func TestChangeEventParserImpl_ProcessStatefulSessionChanges_Cluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	changeEventParser := NewChangeEventParser(mockDao, mockUpdateAction, mockBuilder)

	actions := getMockActionsMap(ctrl)
	entityVersions := map[string]string{domain.RouteConfigurationTable: "test1"}
	nodeGroup := "nodeGroup"
	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindRoutesByClusterNamePrefix("test-cluster||test-cluster||").
		Return([]*domain.Route{{Id: int32(1), DeploymentVersion: "v1", VirtualHostId: int32(1)}}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	granularUpdate := action.GranularEntityUpdate{}
	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, entityVersions[domain.RouteConfigurationTable], routeConfig).Times(1).Return(granularUpdate)
	actions.EXPECT().Put(action.EnvoyRouteConfig, &granularUpdate).Times(1)

	changeEventParser.processStatefulSessionChanges(actions, entityVersions, nodeGroup, changes)
}

func TestCompositeUpdateBuilder_ProcessStatefulSessionChangesRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroup := "nodeGroup"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroup: {
			action.EnvoyRouteConfig: "test-version",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(&domain.Route{Id: int32(1), VirtualHostId: int32(1)}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, "test-version", routeConfig)

	compositeUpdateBuilder.processStatefulSessionCookieChanges(changes)
}

func TestCompositeUpdateBuilder_ProcessStatefulSessionChangesEndpoint(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroup := "nodeGroup"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroup: {
			action.EnvoyRouteConfig: "test-version",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(&domain.Endpoint{Id: int32(1), ClusterId: int32(1), StatefulSessionId: int32(1), DeploymentVersion: "v1"}, nil)
	mockDao.EXPECT().FindClusterById(int32(1)).Return(&domain.Cluster{Id: int32(1), Name: "test-cluster||test-cluster||8080"}, nil)
	mockDao.EXPECT().FindRoutesByClusterNameAndDeploymentVersion("test-cluster||test-cluster||8080", "v1").
		Return([]*domain.Route{{Id: int32(1), DeploymentVersion: "v1", VirtualHostId: int32(1)}}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, "test-version", routeConfig)

	compositeUpdateBuilder.processStatefulSessionCookieChanges(changes)
}

func TestCompositeUpdateBuilder_ProcessStatefulSessionChangesCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	nodeGroup := "nodeGroup"
	nodeGroupEntityVersions := map[string]map[action.EnvoyEntity]string{
		nodeGroup: {
			action.EnvoyRouteConfig: "test-version",
		},
	}
	mockUpdateAction := getMockUpdateAction(ctrl)
	mockBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	compositeUpdateBuilder := newCompositeUpdateBuilder(mockDao, nodeGroupEntityVersions, mockBuilder, mockUpdateAction)

	changes := []memdb.Change{
		{
			After: &domain.StatefulSession{
				Id:                       int32(1),
				CookieName:               "",
				CookiePath:               "",
				Enabled:                  false,
				ClusterName:              "test-cluster",
				Namespace:                "default",
				Gateways:                 []string{"nodeGroup"},
				DeploymentVersion:        "v1",
				InitialDeploymentVersion: "v1",
			},
		},
	}

	routeConfig := &domain.RouteConfiguration{Id: int32(1), NodeGroupId: "nodeGroup"}

	mockDao.EXPECT().FindRouteByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindEndpointByStatefulSession(int32(1)).Return(nil, nil)
	mockDao.EXPECT().FindRoutesByClusterNamePrefix("test-cluster||test-cluster||").
		Return([]*domain.Route{{Id: int32(1), DeploymentVersion: "v1", VirtualHostId: int32(1)}}, nil)
	mockDao.EXPECT().FindVirtualHostById(int32(1)).Return(&domain.VirtualHost{Id: int32(1), RouteConfigurationId: int32(1)}, nil)
	mockDao.EXPECT().FindRouteConfigById(int32(1)).Return(routeConfig, nil)

	mockUpdateAction.EXPECT().RouteConfigUpdate(nodeGroup, "test-version", routeConfig)

	compositeUpdateBuilder.processStatefulSessionCookieChanges(changes)
}
