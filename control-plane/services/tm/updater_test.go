package tm

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	envoy "github.com/netcracker/qubership-core-control-plane/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/envoy/grpc"
	"github.com/netcracker/qubership-core-control-plane/ram"
	clientCache "github.com/netcracker/qubership-core-control-plane/services/cache"
	entitySrv "github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"sync/atomic"
	"testing"
)

var (
	testVal  = create()
	client   = clientCache.NewCacheClient()
	testData = map[string]string{
		"uid-1": "ns-1",
		"uid-2": "ns-2",
		"uid-3": "ns-3",
		"uid-4": "ns-4",
	}
	expectedListeners = []*domain.Listener{
		domain.NewListener("internal-gateway-service-listener", "::", "8080",
			"internal-gateway-service", "internal-gateway-service-routes"),
		domain.NewListener("public-gateway-service-listener", "::", "8080",
			"public-gateway-service", "public-gateway-service-routes"),
		domain.NewListener("private-gateway-service-listener", "::", "8080",
			"private-gateway-service", "private-gateway-service-routes"),
	}
	expectedNsString       = "tenantsNS = {[\"uid-1\"] = \"ns-1\", [\"uid-2\"] = \"ns-2\", [\"uid-3\"] = \"ns-3\", [\"uid-4\"] = \"ns-4\"}"
	inMemDao               = dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
	snapshotCache          = cache.NewSnapshotCache(false, grpc.ClusterHash{}, logger)
	envoyConfigBuilder     = envoy.DefaultEnvoyConfigurationBuilder(inMemDao, entitySrv.NewService("v1"), versionAliasProviderMock{})
	updateManager          = envoy.DefaultUpdateManager(inMemDao, snapshotCache, envoyConfigBuilder)
	tenantNamespaceUpdater = NewTenantNamespaceUpdater(inMemDao, updateManager)
)

func create() int {
	os.Create("application.yaml")
	configloader.Init(configloader.YamlPropertySource())
	return 1
}

func TestUpdateWithEmptyCache(t *testing.T) {
	assert.NotPanics(t, notPanicFunc)
}

func TestPrepareCache(t *testing.T) {
	for key, value := range testData {
		client.Set(key, value)
	}
	selectedItems := client.GetAll()
	assert.Equal(t, len(testData), len(selectedItems))
}

func TestPrepareListeners(t *testing.T) {

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		for _, listener := range expectedListeners {
			assert.Nil(t, dao.SaveListener(listener))
		}
		return nil
	})
	assert.Nil(t, err)

	foundListeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, len(expectedListeners), len(foundListeners))
}

func TestConvertTenantsToUpdate(t *testing.T) {
	assert.NotEmpty(t, client.GetAll())
	actualResult := tenantNamespaceUpdater.convertTenantsToUpdate(client.GetAll())
	assert.NotEmpty(t, actualResult)
	assert.Equal(t, testData, actualResult)
}

func TestConvertToNamespaceMappings(t *testing.T) {
	actualResult := tenantNamespaceUpdater.convertToNamespaceMappings(testData)
	assert.NotEmpty(t, actualResult)
	assert.Equal(t, expectedNsString, actualResult)
}

func TestUpdateWithCache(t *testing.T) {
	assert.NotPanics(t, notPanicFunc)
}

func notPanicFunc() {
	tenantNamespaceUpdater.UpdateAllListeners(client.GetAll())
}

// mocks
type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

type aliasesProviderMock struct {
}

func (a aliasesProviderMock) GetVersionAliases() string {
	return ""
}
