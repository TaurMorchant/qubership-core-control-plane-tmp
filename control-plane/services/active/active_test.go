package active

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/constancy"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
)

const (
	TestPublicGwHostDc1 = "test.public.dc-1.gw"
	TestPublicGwHostDc2 = "test.public.dc-2.gw"
	TestPublicGwHostDc3 = "test.public.dc-3.gw"

	TestPrivateGwHostDc1 = "test.private.dc-1.gw"
	TestPrivateGwHostDc2 = "test.private.dc-2.gw"
	TestPrivateGwHostDc3 = "test.private.dc-3.gw"

	LocalPublicGwHost = TestPublicGwHostDc1
	LocalRivateGwHost = TestPrivateGwHostDc1
)

var TEST_CONTEXT = context.Background()

func prepare(t *testing.T) (*dao.InMemDao, *entity.Service, *bus.EventBusAggregator) {
	storage := getInMemRepo()
	entityService := entity.NewService("v1")
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(storage, internalBus, internalBus, nil, nil)
	dv1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	saveDeploymentVersions(t, storage, dv1)
	return storage, entityService, eventBus
}

func Test_ApplyActiveDCsConfig(t *testing.T) {
	storage, entityService, eventBus := prepare(t)
	activeDCsService := NewActiveDCsService(storage, entityService, eventBus, LocalPublicGwHost, LocalRivateGwHost)

	publicGwHosts := []string{TestPublicGwHostDc1, TestPublicGwHostDc2, TestPublicGwHostDc3}
	privateGwHosts := []string{TestPrivateGwHostDc1, TestPrivateGwHostDc2, TestPrivateGwHostDc3}

	protocols := []string{"https", "http"}
	for _, protocol := range protocols {
		activeDCsV3 := &dto.ActiveDCsV3{
			Protocol:       protocol,
			PublicGwHosts:  publicGwHosts,
			PrivateGwHosts: privateGwHosts,
		}
		err := activeDCsService.ApplyActiveDCsConfig(TEST_CONTEXT, activeDCsV3)
		assert.Nil(t, err)

		// verify clusters:
		// balancing cluster
		publicAACluster, err := storage.FindClusterByName(GetActiveActiveClusterName(PublicGwName))
		assert.Nil(t, err)
		verifyActiveActiveCluster(t, publicAACluster, len(publicGwHosts))
		privateAACluster, err := storage.FindClusterByName(GetActiveActiveClusterName(PrivateGwName))
		assert.Nil(t, err)
		verifyActiveActiveCluster(t, privateAACluster, len(privateGwHosts))

		// clusters for public gws from external DCs
		verifyExternalClusters(t, storage, protocol, PublicGwName, publicGwHosts, LocalPublicGwHost)

		// clusters for private gws from external DCs
		verifyExternalClusters(t, storage, protocol, PrivateGwName, privateGwHosts, LocalRivateGwHost)

		// verify balancing route and other config is present on default virtual host
		verifyActiveActiveConfigForDefaultVirtualHost(t, storage, PublicGwName)
		verifyActiveActiveConfigForDefaultVirtualHost(t, storage, PrivateGwName)
	}
}

func Test_ApplyActiveDCsConfigWithEmptyProtocol(t *testing.T) {
	storage, entityService, eventBus := prepare(t)
	activeDCsService := NewActiveDCsService(storage, entityService, eventBus, LocalPublicGwHost, LocalRivateGwHost)

	publicGwHosts := []string{TestPublicGwHostDc1, TestPublicGwHostDc2, TestPublicGwHostDc3}
	privateGwHosts := []string{TestPrivateGwHostDc1, TestPrivateGwHostDc2, TestPrivateGwHostDc3}

	activeDCsV3 := &dto.ActiveDCsV3{
		PublicGwHosts:  publicGwHosts,
		PrivateGwHosts: privateGwHosts,
	}
	err := activeDCsService.ApplyActiveDCsConfig(TEST_CONTEXT, activeDCsV3)
	assert.NotNil(t, err)
}

func Test_ApplyActiveDCsConfigWithEmptyHosts(t *testing.T) {
	storage, entityService, eventBus := prepare(t)
	activeDCsService := NewActiveDCsService(storage, entityService, eventBus, LocalPublicGwHost, LocalRivateGwHost)

	activeDCsV3 := &dto.ActiveDCsV3{
		Protocol: "http",
	}
	err := activeDCsService.ApplyActiveDCsConfig(TEST_CONTEXT, activeDCsV3)
	assert.NotNil(t, err)
}

func Test_ApplyActiveDCsConfigWithHostsNotMatchingLocalGwHost(t *testing.T) {
	storage, entityService, eventBus := prepare(t)
	localPublicGwHost := "local.public.dc1.host"
	localPrivateGwHost := "local.private.dc1.host"
	activeDCsService := NewActiveDCsService(storage, entityService, eventBus, localPublicGwHost, localPrivateGwHost)

	activeDCsV3 := &dto.ActiveDCsV3{
		Protocol:       "http",
		PublicGwHosts:  []string{"public.dc1.host", "public.dc2.host"},
		PrivateGwHosts: []string{"private.dc1.host", "private.dc2.host"},
	}
	err := activeDCsService.ApplyActiveDCsConfig(TEST_CONTEXT, activeDCsV3)
	assert.NotNil(t, err)
}

func Test_DeleteActiveDCsConfig(t *testing.T) {
	storage, entityService, eventBus := prepare(t)
	activeDCsService := NewActiveDCsService(storage, entityService, eventBus, LocalPublicGwHost, LocalRivateGwHost)

	publicGwHosts := []string{TestPublicGwHostDc1, TestPublicGwHostDc2, TestPublicGwHostDc3}
	privateGwHosts := []string{TestPrivateGwHostDc1, TestPrivateGwHostDc2, TestPrivateGwHostDc3}

	protocol := "https"
	activeDCsV3 := &dto.ActiveDCsV3{
		Protocol:       protocol,
		PublicGwHosts:  publicGwHosts,
		PrivateGwHosts: privateGwHosts,
	}
	err := activeDCsService.ApplyActiveDCsConfig(TEST_CONTEXT, activeDCsV3)
	assert.Nil(t, err)

	err = activeDCsService.DeleteActiveDCsConfig(TEST_CONTEXT)
	assert.Nil(t, err)

	// make sure active-active clusters deleted:
	publicAACluster, err := storage.FindClusterByName(GetActiveActiveClusterName(PublicGwName))
	assert.Nil(t, err)
	assert.Nil(t, publicAACluster)

	privateAACluster, err := storage.FindClusterByName(GetActiveActiveClusterName(PrivateGwName))
	assert.Nil(t, err)
	assert.Nil(t, privateAACluster)

	publicRouteConfig, err := storage.FindRouteConfigByNodeGroupIdAndName(PublicGwName, GetRouteConfigName(PublicGwName))
	assert.Nil(t, err)
	assert.NotNil(t, publicRouteConfig)
	assert.NotEmpty(t, publicRouteConfig.Id)
	privateRouteConfig, err := storage.FindRouteConfigByNodeGroupIdAndName(PrivateGwName, GetRouteConfigName(PrivateGwName))
	assert.Nil(t, err)
	assert.NotNil(t, privateRouteConfig)
	assert.NotEmpty(t, privateRouteConfig.Id)

	// clusters for public gws from external DCs should be deleted
	for i := range publicGwHosts {
		alias := GetDcAlias(i)
		clusterNameForExternalGw := GetClusterNameForExternalGw(PublicGwName, alias)
		cluster, err := storage.FindClusterByName(clusterNameForExternalGw)
		assert.Nil(t, err)
		assert.Nil(t, cluster)
		// check virtual host for GW from external DC
		virtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(GetExternalVirtualHostName(PublicGwName, alias), publicRouteConfig.Id)
		assert.Nil(t, err)
		assert.Nil(t, virtualHost)
	}
	for i := range privateGwHosts {
		alias := GetDcAlias(i)
		clusterNameForExternalGw := GetClusterNameForExternalGw(PrivateGwName, alias)
		cluster, err := storage.FindClusterByName(clusterNameForExternalGw)
		assert.Nil(t, err)
		assert.Nil(t, cluster)
		// check virtual host for GW from external DC
		virtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(GetExternalVirtualHostName(PrivateGwName, alias), privateRouteConfig.Id)
		assert.Nil(t, err)
		assert.Nil(t, virtualHost)
	}

	// default virtualHost for public GW should have no routes
	publicVirtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(PublicGwName, publicRouteConfig.Id)
	assert.Nil(t, err)
	assert.NotNil(t, publicVirtualHost)
	// load routes for this virtualHost
	publicRoutes, err := storage.FindRoutesByVirtualHostId(publicVirtualHost.Id)
	assert.Empty(t, publicRoutes)

	// default virtualHost for private GW should have no routes
	privateVirtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(PrivateGwName, privateRouteConfig.Id)
	assert.Nil(t, err)
	assert.NotNil(t, privateVirtualHost)
	// load routes for this virtualHost
	privateRoutes, err := storage.FindRoutesByVirtualHostId(privateVirtualHost.Id)
	assert.Empty(t, privateRoutes)
}

func verifyActiveActiveCluster(t *testing.T, cluster *domain.Cluster, numOfEndpoints int) {
	assert.NotNil(t, cluster)
	assert.Equal(t, domain.LbPolicyMaglev, cluster.LbPolicy)
	assert.Equal(t, domain.DISCOVERY_TYPE_STRICT_DNS, cluster.DiscoveryType)
	assert.Equal(t, domain.DISCOVERY_TYPE_STRICT_DNS, cluster.DiscoveryTypeOld)

	assert.NotNil(t, cluster.Endpoints)
	assert.Equal(t, numOfEndpoints, len(cluster.Endpoints))
	for i := 0; i < numOfEndpoints; i++ {
		endpoint := getEndpointByOrderId(cluster.Endpoints, int32(i))
		assert.NotNil(t, endpoint, "failed to find endpoint by order id %d", i)
		alias := GetDcAlias(i)
		assert.Equal(t, alias, endpoint.Address)
		assert.Equal(t, alias, endpoint.Hostname)
		assert.Equal(t, int32(8080), endpoint.Port)

	}
	assert.NotNil(t, cluster.DnsResolvers)
	assert.Equal(t, 1, len(cluster.DnsResolvers))
	dnsResolver := cluster.DnsResolvers[0]
	assert.NotNil(t, dnsResolver)
	socketAddress := dnsResolver.SocketAddress
	assert.NotNil(t, socketAddress)
	assert.Equal(t, "UDP", socketAddress.Protocol)
	assert.Equal(t, "127.0.0.1", socketAddress.Address)
	assert.Equal(t, uint32(1053), socketAddress.Port)
	assert.Equal(t, true, socketAddress.IPv4_compat)
	// common load balancing config
	commonLbConfig := cluster.CommonLbConfig
	assert.NotNil(t, commonLbConfig)
	consistentHashingLbConfig := commonLbConfig.ConsistentHashingLbConfig
	assert.NotNil(t, consistentHashingLbConfig)
	assert.Equal(t, true, consistentHashingLbConfig.UseHostnameForHashing)
	// health checks
	healthChecks := cluster.HealthChecks
	assert.NotNil(t, healthChecks)
	assert.Equal(t, 1, len(healthChecks))
	healthCheck := healthChecks[0]
	assert.NotNil(t, healthCheck)
	assert.Equal(t, int64(1000), healthCheck.Timeout)
	assert.Equal(t, int64(1000), healthCheck.Interval)
	assert.Equal(t, true, healthCheck.ReuseConnection)
	assert.Equal(t, int64(1000), healthCheck.NoTrafficInterval)
	assert.Equal(t, int64(1000), healthCheck.UnhealthyInterval)
	assert.Equal(t, uint32(10), healthCheck.HealthyThreshold)
	assert.Equal(t, uint32(10), healthCheck.UnhealthyThreshold)
	assert.Equal(t, false, healthCheck.AlwaysLogHealthCheckFailures)
}

func getEndpointByOrderId(endpoints []*domain.Endpoint, orderId int32) *domain.Endpoint {
	for _, endpoint := range endpoints {
		if endpoint.OrderId == orderId {
			return endpoint
		}
	}
	return nil
}

func verifyExternalClusters(t *testing.T, storage *dao.InMemDao, protocol string, gwName string, gwHosts []string, localGwHost string) {
	for i, gwHost := range gwHosts {
		alias := GetDcAlias(i)
		clusterNameForExternalGw := GetClusterNameForExternalGw(gwName, alias)
		cluster, err := storage.FindClusterByName(clusterNameForExternalGw)
		assert.Nil(t, err)
		if gwHost == localGwHost {
			// there should be no cluster for current DC gw
			assert.Nil(t, cluster)
			continue
		}
		assert.NotNil(t, cluster)
		assert.Equal(t, domain.LbPolicyLeastRequest, cluster.LbPolicy)
		assert.Equal(t, domain.DISCOVERY_TYPE_STRICT_DNS, cluster.DiscoveryType)
		assert.Equal(t, domain.DISCOVERY_TYPE_STRICT_DNS, cluster.DiscoveryTypeOld)
		assert.NotNil(t, cluster.Endpoints)
		assert.Equal(t, 1, len(cluster.Endpoints))
		endpoint := cluster.Endpoints[0]
		assert.NotNil(t, endpoint)
		assert.Equal(t, gwHost, endpoint.Address)
		assert.Equal(t, gwHost, endpoint.Hostname)

		var expectedPort int32
		if protocol == ProtocolHttp {
			expectedPort = int32(80)
		} else if protocol == ProtocolHttps {
			expectedPort = int32(443)
		} else {
			assert.Fail(t, "Unsupported protocol - "+protocol)
		}
		assert.Equal(t, expectedPort, endpoint.Port)

		// verify virtual host to external dc
		routeConfig, err := storage.FindRouteConfigByNodeGroupIdAndName(gwName, GetRouteConfigName(gwName))
		assert.Nil(t, err)
		assert.NotNil(t, routeConfig)
		assert.NotEmpty(t, routeConfig.Id)
		virtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(GetExternalVirtualHostName(gwName, alias), routeConfig.Id)
		assert.Nil(t, err)
		assert.NotNil(t, virtualHost)
		assert.NotNil(t, virtualHost.Routes)
		assert.Equal(t, 1, len(virtualHost.Routes))
		route := virtualHost.Routes[0]
		assert.NotNil(t, route)
		assert.Equal(t, "/", route.Prefix)
		assert.Equal(t, "", route.Path)
		assert.Equal(t, domain.NewNullBool(true), route.HostAutoRewrite)
		assert.Equal(t, clusterNameForExternalGw, route.ClusterName)
	}
}

func verifyActiveActiveConfigForDefaultVirtualHost(t *testing.T, storage *dao.InMemDao, gwName string) {
	routeConfig, err := storage.FindRouteConfigByNodeGroupIdAndName(gwName, GetRouteConfigName(gwName))
	assert.Nil(t, err)
	assert.NotNil(t, routeConfig)
	assert.NotEmpty(t, routeConfig.Id)
	virtualHost, err := storage.FindFirstVirtualHostByNameAndRouteConfigurationId(gwName, routeConfig.Id)
	assert.Nil(t, err)
	assert.NotNil(t, virtualHost)
	assert.NotNil(t, virtualHost.Routes)
	// check route
	assert.Equal(t, 1, len(virtualHost.Routes))
	route := virtualHost.Routes[0]
	assert.NotNil(t, route)
	assert.Equal(t, "/", route.Prefix)
	assert.Equal(t, "", route.Path)
	assert.Equal(t, domain.NewNullBool(true), route.HostAutoRewrite)
	assert.Equal(t, GetActiveActiveClusterName(gwName), route.ClusterName)
	// check hashPolicy
	assert.NotNil(t, route.HashPolicies)
	assert.Equal(t, 1, len(route.HashPolicies))
	hashPolicy := route.HashPolicies[0]
	assert.NotNil(t, hashPolicy)
	assert.Equal(t, AnchorHeaderName, hashPolicy.HeaderName)
	// check retryPolicy
	retryPolicy := route.RetryPolicy
	assert.NotNil(t, retryPolicy)
	assert.Equal(t, "gateway-error,reset,connect-failure", retryPolicy.RetryOn)
	assert.Equal(t, uint32(30), retryPolicy.NumRetries)
	assert.Equal(t, int64(2000), retryPolicy.PerTryTimeout)
	assert.Equal(t, []uint32{200}, retryPolicy.RetriableStatusCodes)
	retryBackOff := retryPolicy.RetryBackOff
	assert.NotNil(t, retryBackOff)
	assert.Equal(t, int64(2000), retryBackOff.BaseInterval)
	assert.Equal(t, int64(2000), retryBackOff.MaxInterval)
	// check matcher
	assert.NotNil(t, route.HeaderMatchers)
	var achorMatcher *domain.HeaderMatcher
	var balancingMatcher *domain.HeaderMatcher
	for _, matcher := range route.HeaderMatchers {
		if matcher.Name == AnchorHeaderName {
			achorMatcher = matcher
		} else if matcher.Name == BalanceTrackHeaderName {
			balancingMatcher = matcher
		}
	}
	assert.NotNil(t, achorMatcher)
	assert.Equal(t, domain.NewNullBool(true), achorMatcher.PresentMatch)
	assert.Equal(t, false, achorMatcher.InvertMatch)

	assert.NotNil(t, balancingMatcher)
	assert.Equal(t, domain.NewNullBool(true), balancingMatcher.PresentMatch)
	assert.Equal(t, true, balancingMatcher.InvertMatch)

	// check requestHeadersToAdd
	var balancingTrackingHeader *domain.Header
	for _, header := range route.RequestHeadersToAdd {
		if header.Name == BalanceTrackHeaderName {
			balancingTrackingHeader = &header
		}
	}
	assert.NotNil(t, balancingTrackingHeader)
	assert.Equal(t, "-", balancingTrackingHeader.Value)
}

func getInMemRepo() *dao.InMemDao {
	return dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, []func([]memdb.Change) error{flushChanges})
}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

func flushChanges(changes []memdb.Change) error {
	flusher := &constancy.Flusher{BatchTm: &batchTransactionManagerMock{}}
	return flusher.Flush(changes)
}

type batchTransactionManagerMock struct{}

func (tm *batchTransactionManagerMock) WithTxBatch(_ func(tx constancy.BatchStorage) error) error {
	return nil
}

func (tm *batchTransactionManagerMock) IsCurrentPodDefinedAsMaster() (bool, error) {
	return true, nil
}

func saveDeploymentVersions(t *testing.T, storage *dao.InMemDao, dVs ...*domain.DeploymentVersion) {
	_, err := storage.WithWTx(func(dao dao.Repository) error {
		for _, dV := range dVs {
			assert.Nil(t, dao.SaveDeploymentVersion(dV))
		}
		return nil
	})
	assert.Nil(t, err)
}
