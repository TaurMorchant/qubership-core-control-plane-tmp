package tm

import (
	"encoding/json"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/google/uuid"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	envoy "github.com/netcracker/qubership-core-control-plane/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/envoy/grpc"
	"github.com/netcracker/qubership-core-control-plane/ram"
	cpCache "github.com/netcracker/qubership-core-control-plane/services/cache"
	entitySrv "github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/tm/entity"
	go_stomp_websocket "github.com/netcracker/qubership-core-lib-go-stomp-websocket/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTenantsNamespacesUpdate(t *testing.T) {
	cacheClient := cpCache.NewCacheClient()
	tenantWatcherClient := TenantWatcherClient{cache: cacheClient}
	testTanant1 := &entity.Tenant{"uid-1", "uid-1", "", entity.TenantSuspended}
	testTanant2 := &entity.Tenant{"uid-1", "uid-1", "", entity.TenantActive}
	testTanant3 := &entity.Tenant{"uid-1", "uid-1", "", entity.TenantSuspended}
	testTanant4 := &entity.Tenant{"uid-2", "", "ns-5", entity.TenantAwaitingApproval}
	testTanant5 := &entity.Tenant{"uid-3", "uid-3", "ns-3", entity.TenantActive}
	testTanant6 := &entity.Tenant{"uid-3", "uid-3", "ns-3", entity.TenantSuspended}
	testTanant7 := &entity.Tenant{"uid-3", "uid-3", "ns-3", entity.TenantActive}
	testTanant8 := &entity.Tenant{"uid-4", "uid-4", "ns-4", ""}
	assert.False(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant1))
	assert.False(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant2))
	assert.False(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant3))
	assert.False(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant4))
	assert.True(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant5))
	assert.True(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant6))
	assert.True(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant7))
	assert.False(t, tenantWatcherClient.tenantsNamespacesUpdate(testTanant8))
}

func TestTenantWatcherClient_Watch(t *testing.T) {
	testSub := &go_stomp_websocket.Subscription{
		FrameCh: make(chan *go_stomp_websocket.Frame),
	}
	cacheClient := cpCache.NewCacheClient()
	memDao := dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
	snapshotCache := cache.NewSnapshotCache(false, grpc.ClusterHash{}, logger)
	envoyConfigBuilder := envoy.DefaultEnvoyConfigurationBuilder(memDao, entitySrv.NewService("v1"), versionAliasProviderMock{})
	updateManager := envoy.DefaultUpdateManager(memDao, snapshotCache, envoyConfigBuilder)
	prepareListeners(t, memDao)
	tenantWatcherClient := TenantWatcherClient{cache: cacheClient, updateTenantNamespace: NewTenantNamespaceUpdater(memDao, updateManager)}
	go runTenantWatcherTests(t, cacheClient, testSub)
	assert.NotPanics(t, func() {
		tenantWatcherClient.Watch(testSub)
	})
}

func prepareListeners(t *testing.T, memDao *dao.InMemDao) {
	listeners := []*domain.Listener{
		domain.NewListener("internal-gateway-service-listener", "::", "8080",
			"internal-gateway-service", "internal-gateway-service-routes"),
		domain.NewListener("public-gateway-service-listener", "::", "8080",
			"public-gateway-service", "public-gateway-service-routes"),
		domain.NewListener("private-gateway-service-listener", "::", "8080",
			"private-gateway-service", "private-gateway-service-routes"),
	}
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		for _, listener := range listeners {
			assert.Nil(t, dao.SaveListener(listener))
		}
		return nil
	})
	assert.Nil(t, err)

	actualListeners, err := memDao.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, len(listeners), len(actualListeners))
	assert.ObjectsAreEqualValues(listeners, actualListeners)
}

func runTenantWatcherTests(t *testing.T, cacheClient *cpCache.CacheClient, testSub *go_stomp_websocket.Subscription) {
	firstTenantId := uuid.New().String()
	sendFrameAndCheckCache(t, testSub, cacheClient, firstTenantId, entity.TenantActive, "test", 1)
	sendFrameAndCheckCache(t, testSub, cacheClient, firstTenantId, entity.TenantSuspended, "test", 0)
	sendFrameAndCheckCache(t, testSub, cacheClient, firstTenantId, entity.TenantActive, "test", 1)
	secondTenantId := uuid.New().String()
	sendFrameAndCheckCache(t, testSub, cacheClient, secondTenantId, entity.TenantActive, "test2", 2)
	thirdTenantId := uuid.New().String()
	sendFrameAndCheckCache(t, testSub, cacheClient, thirdTenantId, entity.TenantActive, "test3", 3)
	fourthTenantId := uuid.New().String()
	sendFrameAndCheckCache(t, testSub, cacheClient, fourthTenantId, entity.TenantAwaitingApproval, "test4", 3)
	sendFrameAndCheckCache(t, testSub, cacheClient, fourthTenantId, entity.TenantActive, "test4", 4)
	sendFrameAndCheckCache(t, testSub, cacheClient, fourthTenantId, entity.TenantSuspended, "test4", 3)

	testSub.FrameCh <- &go_stomp_websocket.Frame{Body: ""}
	time.Sleep(200 * time.Millisecond)
	checkCache(t, cacheClient, 3)
	close(testSub.FrameCh)
}

func sendFrameAndCheckCache(t *testing.T, testSub *go_stomp_websocket.Subscription,
	cacheClient *cpCache.CacheClient, tenantId, tenantStatus, namespace string, expectedCacheSize int) {
	testSub.FrameCh <- getFrameWithTenant(t, tenantId, tenantStatus, namespace)
	time.Sleep(200 * time.Millisecond)
	actualNamespace, ok := cacheClient.Get(tenantId)
	if tenantStatus == entity.TenantActive {
		assert.True(t, ok)
		assert.Equal(t, namespace, actualNamespace)
	} else {
		assert.False(t, ok)
		assert.Equal(t, nil, actualNamespace)
	}
	checkCache(t, cacheClient, expectedCacheSize)
}

func checkCache(t *testing.T, cacheClient *cpCache.CacheClient, expectedCacheSize int) {
	entireCache := cacheClient.GetAll()
	assert.NotNil(t, entireCache)
	assert.Equal(t, expectedCacheSize, len(entireCache))
}

func getFrameWithTenant(t *testing.T, id, tenantStatus, namespace string) *go_stomp_websocket.Frame {
	tenants := entity.WatchApiTenant{
		Tenants: []*entity.Tenant{
			{
				ExternalId: id,
				ObjectId:   id,
				Status:     tenantStatus,
				Namespace:  namespace,
			},
		},
	}
	marshaledTenants, err := json.Marshal(tenants)
	assert.Nil(t, err)
	assert.NotEmpty(t, marshaledTenants)
	return &go_stomp_websocket.Frame{Body: string(marshaledTenants)}
}

type versionAliasProviderMock struct{}

func (v versionAliasProviderMock) GetVersionAliases() string {
	return ""
}
