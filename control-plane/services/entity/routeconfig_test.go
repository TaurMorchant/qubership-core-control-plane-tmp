package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindRouteConfigurationByVirtualHostId_shouldPanic_whenHostNotFound(t *testing.T) {
	entityService, inMemDao := getService(t)
	virtualHostId := int32(-1)
	assert.Panics(t, func() { entityService.FindRouteConfigurationByVirtualHostId(inMemDao, virtualHostId) }, "The code should panic")
}

func TestFindRouteConfigurationByVirtualHostId_shouldReturnConfig_whenConfigExist(t *testing.T) {
	entityService, inMemDao := getService(t)

	routeConfig := domain.NewRouteConfiguration("test-routeconfig", "test-nodegroup")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutRouteConfig(dao, routeConfig)
	})

	virtualHost := domain.NewVirtualHost("test-virtualhost", routeConfig.Id)
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutVirtualHost(dao, virtualHost)
	})

	foundRouteConfigs, err := entityService.FindRouteConfigurationByVirtualHostId(inMemDao, virtualHost.Id)
	assert.Nil(t, err)
	assert.NotNil(t, foundRouteConfigs)
	assert.Equal(t, routeConfig, foundRouteConfigs)
}

func TestService_PutRouteConfigIfDoesNotExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedRouteConfig := domain.NewRouteConfiguration("test-routeconfig", "test-nodegroup")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutRouteConfig(dao, expectedRouteConfig)
	})
	assert.Nil(t, err)

	actualRouteConfigs, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRouteConfigs)
	assert.Equal(t, 1, len(actualRouteConfigs))
	assert.EqualValues(t, expectedRouteConfig, actualRouteConfigs[0])
}

func TestService_PutRouteConfigIfExists(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedRouteConfig := domain.NewRouteConfiguration("test-routeconfig", "test-nodegroup")

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return dao.SaveRouteConfig(expectedRouteConfig)
	})
	assert.Nil(t, err)
	actualRouteConfigs, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRouteConfigs)
	assert.EqualValues(t, expectedRouteConfig, actualRouteConfigs[0])

	expectedRouteConfig = domain.NewRouteConfiguration("test-routeconfig", "test-nodegroup")
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutRouteConfig(dao, expectedRouteConfig)
	})
	assert.Nil(t, err)

	actualRouteConfigs, err = inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRouteConfigs)
	assert.Equal(t, 1, len(actualRouteConfigs))
	assert.EqualValues(t, expectedRouteConfig, actualRouteConfigs[0])
}

func TestService_PutRouteConfigsWhichDoNotExist(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedFirstRouteConfig := domain.NewRouteConfiguration("test-routeconfig1", "test-nodegroup")
	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutRouteConfig(dao, expectedFirstRouteConfig)
	})
	assert.Nil(t, err)

	expectedSecondRouteConfig := domain.NewRouteConfiguration("test-routeconfig2", "test-nodegroup")
	_, err = inMemDao.WithWTx(func(dao dao.Repository) error {
		return entityService.PutRouteConfig(dao, expectedSecondRouteConfig)
	})
	assert.Nil(t, err)

	actualRouteConfigs, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRouteConfigs)
	assert.Equal(t, 2, len(actualRouteConfigs))
	assert.Contains(t, actualRouteConfigs, expectedFirstRouteConfig)
	assert.Contains(t, actualRouteConfigs, expectedSecondRouteConfig)
}

func TestService_GetRouteConfigurationsWithRelations(t *testing.T) {
	entityService, inMemDao := getService(t)
	expectedRouteConfig := prepareRouteConfigData(t, inMemDao)
	actualRouteConfigs, err := entityService.GetRouteConfigurationsWithRelations(inMemDao)
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRouteConfigs)
	assert.Equal(t, 1, len(actualRouteConfigs))
	assert.NotEmpty(t, actualRouteConfigs[0].VirtualHosts)
	assert.Equal(t, 1, len(actualRouteConfigs[0].VirtualHosts))
	AssertDeepEqual(t, expectedRouteConfig, actualRouteConfigs[0], domain.RouteConfigurationTable)
}

func TestService_GetRouteConfigurationsWithRelationsWhereNoData(t *testing.T) {
	entityService, inMemDao := getService(t)
	routeConfigs, err := entityService.GetRouteConfigurationsWithRelations(inMemDao)
	assert.Nil(t, err)
	assert.Empty(t, routeConfigs)
}

func prepareRouteConfigData(t *testing.T, memDao *dao.InMemDao) *domain.RouteConfiguration {
	routeConfig := domain.NewRouteConfiguration("test-cluster", "test-nodegroup")
	virtualHost := domain.NewVirtualHost("some-vh", 0)
	routeConfig.Id, virtualHost.Id = saveRouteConfigData(t, memDao, *routeConfig, *virtualHost)
	virtualHost.RouteConfigurationId = routeConfig.Id
	routeConfig.VirtualHosts = []*domain.VirtualHost{virtualHost}
	return routeConfig
}

func saveRouteConfigData(t *testing.T, memDao *dao.InMemDao, routeConfig domain.RouteConfiguration, virtualHost domain.VirtualHost) (int32, int32) {
	_, err := memDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveRouteConfig(&routeConfig))
		virtualHost.RouteConfigurationId = routeConfig.Id
		assert.Nil(t, dao.SaveVirtualHost(&virtualHost))
		return nil
	})
	assert.Nil(t, err)
	return routeConfig.Id, virtualHost.Id
}

func RouteConfigsEqual(expected, actual *domain.RouteConfiguration) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if expected.Id != actual.Id ||
		expected.Name != actual.Name ||
		expected.NodeGroupId != actual.NodeGroupId ||
		expected.Version != actual.Version {
		return false
	}

	// compare virtual hosts
	if len(expected.VirtualHosts) != len(actual.VirtualHosts) {
		return false
	}
	for _, expectedRoute := range expected.VirtualHosts {
		presentsInBothLists := false
		for _, actualRoute := range actual.VirtualHosts {
			if VirtualHostsEqual(expectedRoute, actualRoute) {
				presentsInBothLists = true
				break
			}
		}
		if !presentsInBothLists {
			return false
		}
	}
	for _, actualRoute := range actual.VirtualHosts {
		presentsInBothLists := false
		for _, expectedRoute := range expected.VirtualHosts {
			if VirtualHostsEqual(expectedRoute, actualRoute) {
				presentsInBothLists = true
				break
			}
		}
		if !presentsInBothLists {
			return false
		}
	}
	return true
}
