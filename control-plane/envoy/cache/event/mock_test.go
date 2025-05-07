package event

import (
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	mock_dao "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/dao"
	mock_action "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/envoy/cache/action"
	"strconv"
)

func getMockDao(ctrl *gomock.Controller) *mock_dao.MockDao {
	mockDao := mock_dao.NewMockDao(ctrl)
	mockDao.EXPECT().WithWTx(gomock.Any()).AnyTimes().Return(nil, nil)
	return mockDao
}

func getMockUpdateAction(ctrl *gomock.Controller) *mock_action.MockUpdateActionFactory {
	mockUpdateAction := mock_action.NewMockUpdateActionFactory(ctrl)
	return mockUpdateAction
}

func getMockSnapshotUpdateAction(ctrl *gomock.Controller) *mock_action.MockSnapshotUpdateAction {
	mock := mock_action.NewMockSnapshotUpdateAction(ctrl)
	return mock
}

func getMockActionsMap(ctrl *gomock.Controller) *mock_action.MockActionsMap {
	mockActionsMap := mock_action.NewMockActionsMap(ctrl)
	return mockActionsMap
}

func getAndExpectRoutesWithVirtualHostId(mockDao *mock_dao.MockDao, ids ...int32) []*domain.Route {
	routes := make([]*domain.Route, len(ids))
	for i, id := range ids {
		routes[i] = &domain.Route{
			Id:            id,
			VirtualHostId: id,
		}
		mockDao.EXPECT().FindRouteById(id).Return(routes[i], nil)
	}

	return routes
}

func getAndExpectEndpointsWithClusterId(mockDao *mock_dao.MockDao, ids ...int32) []*domain.Endpoint {
	endpoints := make([]*domain.Endpoint, len(ids))
	for i, id := range ids {
		endpoints[i] = &domain.Endpoint{
			Id:        id,
			ClusterId: id,
		}
		mockDao.EXPECT().FindEndpointById(id).Return(endpoints[i], nil)
	}

	return endpoints
}

func getAndExpectVirtualHostWithRouteConfigurationId(mockDao *mock_dao.MockDao, ids ...int32) []*domain.VirtualHost {
	virtualHosts := make([]*domain.VirtualHost, len(ids))
	for i, id := range ids {
		virtualHosts[i] = &domain.VirtualHost{
			Id:                   id,
			RouteConfigurationId: id,
		}
		mockDao.EXPECT().FindVirtualHostById(id).Return(virtualHosts[i], nil)
	}

	return virtualHosts
}

func getAndExpectRouteConfigurationWithNodeGroupId(mockDao *mock_dao.MockDao, ids ...int32) []*domain.RouteConfiguration {
	routeConfigurations := make([]*domain.RouteConfiguration, len(ids))
	for i, id := range ids {
		routeConfigurations[i] = &domain.RouteConfiguration{
			Id:          id,
			NodeGroupId: strconv.Itoa(int(id)),
		}
		mockDao.EXPECT().FindRouteConfigById(id).Return(routeConfigurations[i], nil)
	}

	return routeConfigurations
}

func getAndExpectClusters(mockDao *mock_dao.MockDao, ids ...int32) []*domain.Cluster {
	clusters := make([]*domain.Cluster, len(ids))
	for i, id := range ids {
		clusters[i] = &domain.Cluster{
			Id: id,
		}
		mockDao.EXPECT().FindClusterById(id).Return(clusters[i], nil)
	}

	return clusters
}

func getClusters(ids ...int32) []*domain.Cluster {
	clusters := make([]*domain.Cluster, len(ids))
	for i, id := range ids {
		clusters[i] = &domain.Cluster{
			Id: id,
		}
	}
	return clusters
}

func getCircuitBreakers(ids ...int32) []*domain.CircuitBreaker {
	circuitBreakers := make([]*domain.CircuitBreaker, len(ids))
	for i, id := range ids {
		circuitBreakers[i] = &domain.CircuitBreaker{
			Id: id,
		}
	}
	return circuitBreakers
}
