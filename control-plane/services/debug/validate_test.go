package debug

import (
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/composite"
	"github.com/netcracker/qubership-core-control-plane/domain"
	mock_dao "github.com/netcracker/qubership-core-control-plane/test/mock/dao"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateConfigOK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()

	statusConfig, err := ValidateConfig(mockDao, nil)
	assert.Nil(t, err)

	assert.Equal(t, "ok", statusConfig.Status)
	assert.Equal(t, 0, len(statusConfig.Problems))
}

func TestValidateConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	testValidateConfigVHConflictStar(mockDao)
	testValidateConfigDuplicatedClusters(mockDao)
	testValidateConfigBGD1Clusters(mockDao)
	testValidateConfigOrphanedClusters(mockDao)
	testValidateConfigSlashPrefix(mockDao)
	testValidateConfigLoop(mockDao)

	service := composite.NewService("", composite.BaselineMode, nil, nil, nil, nil)

	statusConfig, err := ValidateConfig(mockDao, service)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)

	assert.Equal(t, 6, len(statusConfig.Problems))

	star1, star2 := false, false
	for _, problem := range statusConfig.Problems {
		switch problem.ProblemType {
		case VHostsConflict.String():
			assert.Equal(t, Critical.String(), problem.Severity)
			assert.Equal(t, VHostsConflict.getMessage(), problem.Message)
			assert.Equal(t, 2, len(problem.Details))
			for _, detail := range problem.Details {
				switch detail.Gateway {
				case "test-node-group-star1":
					star1 = true
					assert.Equal(t, 1, len(detail.VirtualServices))
					assert.Equal(t, "test-gateway-service-star1", detail.VirtualServices[0].Name)
					assert.Equal(t, 2, len(detail.VirtualServices[0].Hosts))
					for _, host := range statusConfig.Problems[0].Details[0].VirtualServices[0].Hosts {
						assert.True(t, host == "*" || host == "test-gateway-service-star1")
					}
				case "test-node-group-star2":
					star2 = true
					assert.Equal(t, 1, len(detail.VirtualServices))
					assert.Equal(t, "test-gateway-service-star2", detail.VirtualServices[0].Name)
					assert.Equal(t, 1, len(detail.VirtualServices[0].Hosts))
					assert.Equal(t, "*", detail.VirtualServices[0].Hosts[0])
				}
			}
		case ClusterDuplicate.String():
			assert.Equal(t, Major.String(), problem.Severity)
			assert.Equal(t, ClusterDuplicate.String(), problem.ProblemType)
			assert.Equal(t, ClusterDuplicate.getMessage(), problem.Message)
			assert.Equal(t, 1, len(problem.Details))
			assert.Nil(t, problem.Details[0].VirtualServices)
			assert.Equal(t, 2, len(problem.Details[0].Clusters))
			for _, cluster := range problem.Details[0].Clusters {
				assert.Equal(t, 1, len(cluster.Endpoints))
				assert.True(t, (cluster.Name == "test-cluster-duplicated1||test-cluster-duplicated1||8080" || cluster.Name == "test-cluster-duplicated2||test-cluster-duplicated2||8080") || cluster.Endpoints[0] == "test-endpoint-duplicated")
			}
		case Bgd1Cluster.String():
			assert.Equal(t, Bgd1Cluster.String(), problem.ProblemType)
			assert.Equal(t, Bgd1Cluster.getMessage(), problem.Message)
			assert.Equal(t, 1, len(problem.Details))
			assert.Nil(t, problem.Details[0].VirtualServices)
			assert.Equal(t, 1, len(problem.Details[0].Clusters))
			assert.Equal(t, "test-cluster-bgd1||test-cluster-bgd1||8080", problem.Details[0].Clusters[0].Name)
			assert.Equal(t, 2, len(problem.Details[0].Clusters[0].Endpoints))
			for _, endpoint := range problem.Details[0].Clusters[0].Endpoints {
				assert.True(t, endpoint == "test-endpoint-bgd1:8080" || endpoint == "test-endpoint-bgd1second:8080")
			}
		case OrphanedCluster.String():
			assert.Equal(t, Warning.String(), problem.Severity)
			assert.Equal(t, OrphanedCluster.String(), problem.ProblemType)
			assert.Equal(t, OrphanedCluster.getMessage(), problem.Message)
			assert.Equal(t, 1, len(problem.Details))
			assert.Nil(t, problem.Details[0].VirtualServices)
			assert.Equal(t, 1, len(problem.Details[0].Clusters))
			assert.Equal(t, "test-cluster-orphaned||test-cluster-orphaned||8080", problem.Details[0].Clusters[0].Name)
			assert.Equal(t, 1, len(problem.Details[0].Clusters[0].Endpoints))
			assert.Equal(t, "test-endpoint-orphaned:8080", problem.Details[0].Clusters[0].Endpoints[0])
		case PrefixSlash.String():
			assert.Equal(t, Critical.String(), problem.Severity)
			assert.Equal(t, PrefixSlash.String(), problem.ProblemType)
			assert.Equal(t, PrefixSlash.getMessage(), problem.Message)
			assert.Equal(t, 1, len(problem.Details))
			assert.Equal(t, domain.PublicGateway, problem.Details[0].Gateway)
			assert.Nil(t, problem.Details[0].VirtualServices)
			assert.Equal(t, 1, len(problem.Details[0].Clusters))
			assert.Equal(t, "test-cluster-slash||test-cluster-slash||8080", problem.Details[0].Clusters[0].Name)
			assert.Equal(t, 1, len(problem.Details[0].Clusters[0].Endpoints))
			assert.Equal(t, "test-endpoint-slash:8080", problem.Details[0].Clusters[0].Endpoints[0])
		case Loop.String():
			assert.Equal(t, Critical.String(), problem.Severity)
			assert.Equal(t, Loop.String(), problem.ProblemType)
			assert.Equal(t, Loop.getMessage(), problem.Message)
			assert.Equal(t, 1, len(problem.Details))
			assert.Equal(t, "test-node-group-loop", problem.Details[0].Gateway)
			assert.Equal(t, 1, len(problem.Details[0].VirtualServices))
			assert.Equal(t, 1, len(problem.Details[0].Clusters))
			assert.Equal(t, "test-gateway-service-loop", problem.Details[0].VirtualServices[0].Name)
			assert.Equal(t, 1, len(problem.Details[0].VirtualServices[0].Hosts))
			assert.Equal(t, "test-gateway-service-loop:8080", problem.Details[0].VirtualServices[0].Hosts[0])
			assert.Equal(t, 1, len(problem.Details[0].Clusters[0].Endpoints))
			assert.Equal(t, "test-cluster-loop||test-cluster-loop||8080", problem.Details[0].Clusters[0].Name)
			assert.Equal(t, "test-gateway-service-loop:8080", problem.Details[0].Clusters[0].Endpoints[0])
		}
	}
	assert.True(t, star1)
	assert.True(t, star2)
}

// hosts(domains) on two virtual services: on one "*" and another usual
func TestValidateConfigVHConflictStar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	testValidateConfigVHConflictStar(mockDao)

	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()

	statusConfig, err := ValidateConfig(mockDao, nil)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Critical.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, VHostsConflict.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, VHostsConflict.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 2, len(statusConfig.Problems[0].Details))

	star1, star2 := false, false
	for _, detail := range statusConfig.Problems[0].Details {
		switch detail.Gateway {
		case "test-node-group-star1":
			star1 = true
			assert.Equal(t, 1, len(detail.VirtualServices))
			assert.Equal(t, "test-gateway-service-star1", detail.VirtualServices[0].Name)
			assert.Equal(t, 2, len(detail.VirtualServices[0].Hosts))
			for _, host := range statusConfig.Problems[0].Details[0].VirtualServices[0].Hosts {
				assert.True(t, host == "*" || host == "test-gateway-service-star1")
			}
		case "test-node-group-star2":
			star2 = true
			assert.Equal(t, 1, len(detail.VirtualServices))
			assert.Equal(t, "test-gateway-service-star2", detail.VirtualServices[0].Name)
			assert.Equal(t, 1, len(detail.VirtualServices[0].Hosts))
			assert.Equal(t, "*", detail.VirtualServices[0].Hosts[0])
		}
	}
	assert.True(t, star1)
	assert.True(t, star2)
}

// hosts(domains) on one virtual service: *, another AND
// hosts(domains) on two virtual services: on one "*" and another usual (we return only with star)
func testValidateConfigVHConflictStar(mockDao *mock_dao.MockDao) {
	nodeGroups := []*domain.NodeGroup{
		domain.NewNodeGroup("test-node-group-star1"),
		domain.NewNodeGroup("test-node-group-star2"),
	}

	mockDao.EXPECT().FindAllNodeGroups().Return(nodeGroups, nil)

	routeConfigStar1 := &domain.RouteConfiguration{
		Id:          99,
		Name:        "test-node-group-routes-star1",
		NodeGroupId: "test-node-group-star1",
	}

	routeConfigStar2 := &domain.RouteConfiguration{
		Id:          98,
		Name:        "test-node-group-routes-star2",
		NodeGroupId: "test-node-group-star2",
	}

	mockDao.EXPECT().FindRouteConfigsByNodeGroupId("test-node-group-star1").Return([]*domain.RouteConfiguration{routeConfigStar1}, nil)
	mockDao.EXPECT().FindRouteConfigsByNodeGroupId("test-node-group-star2").Return([]*domain.RouteConfiguration{routeConfigStar2}, nil)

	virtualHostStar1 := &domain.VirtualHost{
		Id:                   99,
		Name:                 "test-gateway-service-star1",
		RouteConfigurationId: 99,
	}

	mockDao.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(99)).Return([]*domain.VirtualHost{virtualHostStar1}, nil)

	virtualHostStar2 := &domain.VirtualHost{
		Id:                   98,
		Name:                 "test-gateway-service-star2",
		RouteConfigurationId: 98,
	}

	virtualHostStar2Without := &domain.VirtualHost{
		Id:                   97,
		Name:                 "test-gateway-service-star2Without",
		RouteConfigurationId: 98,
	}

	mockDao.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(98)).Return([]*domain.VirtualHost{virtualHostStar2, virtualHostStar2Without}, nil)

	virtualHostDomainsStar1WithStar := &domain.VirtualHostDomain{
		Domain:        "*",
		VirtualHostId: 99,
	}

	virtualHostDomainsStar1WithoutStar := &domain.VirtualHostDomain{
		Domain:        "test-gateway-service-star1",
		Version:       1,
		VirtualHostId: 99,
	}

	mockDao.EXPECT().FindVirtualHostDomainByVirtualHostId(int32(99)).Return([]*domain.VirtualHostDomain{virtualHostDomainsStar1WithStar, virtualHostDomainsStar1WithoutStar}, nil)

	virtualHostDomainsStar2WithStar := &domain.VirtualHostDomain{
		Domain:        "*",
		VirtualHostId: 98,
	}

	virtualHostDomainsStar2WithoutStar := &domain.VirtualHostDomain{
		Domain:        "test-gateway-service-star2",
		VirtualHostId: 97,
	}

	mockDao.EXPECT().FindVirtualHostDomainByVirtualHostId(int32(98)).Return([]*domain.VirtualHostDomain{virtualHostDomainsStar2WithStar}, nil)
	mockDao.EXPECT().FindVirtualHostDomainByVirtualHostId(int32(97)).Return([]*domain.VirtualHostDomain{virtualHostDomainsStar2WithoutStar}, nil)
}

func TestValidateConfigDuplicatedClusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	testValidateConfigDuplicatedClusters(mockDao)

	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()

	statusConfig, err := ValidateConfig(mockDao, nil)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Major.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, ClusterDuplicate.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, ClusterDuplicate.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details))
	assert.Nil(t, statusConfig.Problems[0].Details[0].VirtualServices)
	assert.Equal(t, 2, len(statusConfig.Problems[0].Details[0].Clusters))
	for _, cluster := range statusConfig.Problems[0].Details[0].Clusters {
		assert.Equal(t, 1, len(cluster.Endpoints))
		assert.True(t, (cluster.Name == "test-cluster-duplicated1||test-cluster-duplicated1||8080" || cluster.Name == "test-cluster-duplicated2||test-cluster-duplicated2||8080") || cluster.Endpoints[0] == "test-endpoint-duplicated")
	}
}

func testValidateConfigDuplicatedClusters(mockDao *mock_dao.MockDao) {
	cluster1 := &domain.Cluster{
		Id:   88,
		Name: "test-cluster-duplicated1||test-cluster-duplicated1||8080",
	}

	cluster2 := &domain.Cluster{
		Id:   87,
		Name: "test-cluster-duplicated2||test-cluster-duplicated2||8080",
	}

	mockDao.EXPECT().FindAllClusters().Return([]*domain.Cluster{cluster1, cluster2}, nil)

	endpoint1 := &domain.Endpoint{
		Id:        88,
		Address:   "test-endpoint-duplicated",
		Port:      8080,
		ClusterId: 88,
	}

	endpoint2 := &domain.Endpoint{
		Id:        87,
		Address:   "test-endpoint-duplicated",
		Port:      8080,
		ClusterId: 87,
	}

	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-duplicated1||test-cluster-duplicated1||8080").Return([]*domain.Endpoint{endpoint1}, nil)
	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-duplicated2||test-cluster-duplicated2||8080").Return([]*domain.Endpoint{endpoint2}, nil)
}

func TestValidateConfigBGD1Clusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	mockDao.EXPECT().FindAllClusters().Return(nil, nil)

	testValidateConfigBGD1Clusters(mockDao)

	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()

	statusConfig, err := ValidateConfig(mockDao, nil)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Major.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, Bgd1Cluster.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, Bgd1Cluster.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details))
	assert.Nil(t, statusConfig.Problems[0].Details[0].VirtualServices)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters))
	assert.Equal(t, "test-cluster-bgd1||test-cluster-bgd1||8080", statusConfig.Problems[0].Details[0].Clusters[0].Name)
	assert.Equal(t, 2, len(statusConfig.Problems[0].Details[0].Clusters[0].Endpoints))
	for _, endpoint := range statusConfig.Problems[0].Details[0].Clusters[0].Endpoints {
		assert.True(t, endpoint == "test-endpoint-bgd1:8080" || endpoint == "test-endpoint-bgd1second:8080")
	}
}

func testValidateConfigBGD1Clusters(mockDao *mock_dao.MockDao) {
	cluster := &domain.Cluster{
		Id:   77,
		Name: "test-cluster-bgd1||test-cluster-bgd1||8080",
	}

	mockDao.EXPECT().FindAllClusters().Return([]*domain.Cluster{cluster}, nil)

	endpoint1 := &domain.Endpoint{
		Id:        77,
		Address:   "test-endpoint-bgd1",
		Port:      8080,
		ClusterId: 77,
	}

	endpoint2 := &domain.Endpoint{
		Id:        76,
		Address:   "test-endpoint-bgd1second",
		Port:      8080,
		ClusterId: 77,
	}

	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-bgd1||test-cluster-bgd1||8080").Return([]*domain.Endpoint{endpoint1, endpoint2}, nil)
}

func TestValidateConfigOrphanedClusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	mockDao.EXPECT().FindAllClusters().Return(nil, nil).Times(2)

	testValidateConfigOrphanedClusters(mockDao)

	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()

	statusConfig, err := ValidateConfig(mockDao, nil)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Warning.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, OrphanedCluster.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, OrphanedCluster.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details))
	assert.Nil(t, statusConfig.Problems[0].Details[0].VirtualServices)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters))
	assert.Equal(t, "test-cluster-orphaned||test-cluster-orphaned||8080", statusConfig.Problems[0].Details[0].Clusters[0].Name)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters[0].Endpoints))
	assert.Equal(t, "test-endpoint-orphaned:8080", statusConfig.Problems[0].Details[0].Clusters[0].Endpoints[0])
}

func testValidateConfigOrphanedClusters(mockDao *mock_dao.MockDao) {
	cluster := &domain.Cluster{
		Id:   66,
		Name: "test-cluster-orphaned||test-cluster-orphaned||8080",
	}

	mockDao.EXPECT().FindAllClusters().Return([]*domain.Cluster{cluster}, nil)
	mockDao.EXPECT().FindAllRoutes().Return(nil, nil)

	endpoint := &domain.Endpoint{
		Id:        66,
		Address:   "test-endpoint-orphaned",
		Port:      8080,
		ClusterId: 66,
	}

	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-orphaned||test-cluster-orphaned||8080").Return([]*domain.Endpoint{endpoint}, nil)
}

func TestValidateConfigSlashPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	testValidateConfigSlashPrefix(mockDao)

	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllRouteConfigs().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()

	service := composite.NewService("", composite.BaselineMode, nil, nil, nil, nil)

	statusConfig, err := ValidateConfig(mockDao, service)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Critical.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, PrefixSlash.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, PrefixSlash.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details))
	assert.Equal(t, domain.PublicGateway, statusConfig.Problems[0].Details[0].Gateway)
	assert.Nil(t, statusConfig.Problems[0].Details[0].VirtualServices)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters))
	assert.Equal(t, "test-cluster-slash||test-cluster-slash||8080", statusConfig.Problems[0].Details[0].Clusters[0].Name)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters[0].Endpoints))
	assert.Equal(t, "test-endpoint-slash:8080", statusConfig.Problems[0].Details[0].Clusters[0].Endpoints[0])
}

func testValidateConfigSlashPrefix(mockDao *mock_dao.MockDao) {
	virtualHost := &domain.VirtualHost{
		Id:   3,
		Name: domain.PublicGateway,
	}

	mockDao.EXPECT().FindAllVirtualHosts().Return([]*domain.VirtualHost{virtualHost}, nil)

	route := &domain.Route{
		Id:            555,
		VirtualHostId: 3,
		ClusterName:   "test-cluster-slash||test-cluster-slash||8080",
		Prefix:        "/",
	}

	mockDao.EXPECT().FindRoutesByVirtualHostId(int32(3)).Return([]*domain.Route{route}, nil)

	endpoint := &domain.Endpoint{
		Id:        55,
		Address:   "test-endpoint-slash",
		Port:      8080,
		ClusterId: 55,
	}

	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-slash||test-cluster-slash||8080").Return([]*domain.Endpoint{endpoint}, nil)
}

func TestValidateConfigLoop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockDao := mock_dao.NewMockDao(ctrl)

	testValidateConfigLoop(mockDao)

	mockDao.EXPECT().FindAllRoutes().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllNodeGroups().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllClusters().Return(nil, nil).AnyTimes()
	mockDao.EXPECT().FindAllVirtualHosts().Return(nil, nil).AnyTimes()

	service := composite.NewService("", composite.BaselineMode, nil, nil, nil, nil)

	statusConfig, err := ValidateConfig(mockDao, service)
	assert.Nil(t, err)

	assert.Equal(t, "problem", statusConfig.Status)
	assert.Equal(t, 1, len(statusConfig.Problems))
	assert.Equal(t, Critical.String(), statusConfig.Problems[0].Severity)
	assert.Equal(t, Loop.String(), statusConfig.Problems[0].ProblemType)
	assert.Equal(t, Loop.getMessage(), statusConfig.Problems[0].Message)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details))
	assert.Equal(t, "test-node-group-loop", statusConfig.Problems[0].Details[0].Gateway)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].VirtualServices))
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters))
	assert.Equal(t, "test-gateway-service-loop", statusConfig.Problems[0].Details[0].VirtualServices[0].Name)
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].VirtualServices[0].Hosts))
	assert.Equal(t, "test-gateway-service-loop:8080", statusConfig.Problems[0].Details[0].VirtualServices[0].Hosts[0])
	assert.Equal(t, 1, len(statusConfig.Problems[0].Details[0].Clusters[0].Endpoints))
	assert.Equal(t, "test-cluster-loop||test-cluster-loop||8080", statusConfig.Problems[0].Details[0].Clusters[0].Name)
	assert.Equal(t, "test-gateway-service-loop:8080", statusConfig.Problems[0].Details[0].Clusters[0].Endpoints[0])
}

func testValidateConfigLoop(mockDao *mock_dao.MockDao) {
	routeConfig := &domain.RouteConfiguration{
		Id:          44,
		Name:        "test-node-group-routes-loop",
		NodeGroupId: "test-node-group-loop",
	}

	mockDao.EXPECT().FindAllRouteConfigs().Return([]*domain.RouteConfiguration{routeConfig}, nil)

	virtualHost := &domain.VirtualHost{
		Id:                   44,
		RouteConfigurationId: 44,
		Name:                 "test-gateway-service-loop",
	}

	mockDao.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(44)).Return([]*domain.VirtualHost{virtualHost}, nil)

	virtualHostDomain := &domain.VirtualHostDomain{
		Domain:        "test-gateway-service-loop:8080",
		VirtualHostId: 44,
	}

	mockDao.EXPECT().FindVirtualHostDomainByVirtualHostId(int32(44)).Return([]*domain.VirtualHostDomain{virtualHostDomain}, nil)

	route := &domain.Route{
		Id:            444,
		VirtualHostId: 44,
		ClusterName:   "test-cluster-loop||test-cluster-loop||8080",
		Prefix:        "/test/loop",
	}

	mockDao.EXPECT().FindRoutesByVirtualHostId(int32(44)).Return([]*domain.Route{route}, nil)

	endpoint := &domain.Endpoint{
		Id:        44,
		Address:   "test-gateway-service-loop",
		Port:      8080,
		ClusterId: 44,
	}

	mockDao.EXPECT().FindEndpointsByClusterName("test-cluster-loop||test-cluster-loop||8080").Return([]*domain.Endpoint{endpoint}, nil)
}
