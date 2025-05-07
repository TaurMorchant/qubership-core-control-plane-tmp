package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteVirtualHostsDomain(t *testing.T) {
	testable, _, vHostDomains := initTestData(t)

	_, _ = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteVirtualHostsDomain(vHostDomains[len(vHostDomains)-1]))
		return nil
	})

	foundVHostDomains, err := testable.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Equal(t, len(vHostDomains)-1, len(foundVHostDomains))
	for i, foundDomain := range foundVHostDomains {
		assert.Equal(t, vHostDomains[i], foundDomain)
	}
}

func TestDeleteVirtualHost(t *testing.T) {
	testable, vHosts, _ := initTestData(t)

	_, _ = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteVirtualHost(vHosts[len(vHosts)-1]))
		return nil
	})

	foundVHosts, err := testable.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Equal(t, len(vHosts)-1, len(foundVHosts))
	for i, foundHost := range foundVHosts {
		assert.Equal(t, vHosts[i], foundHost)
	}
}

func TestFindAllVirtualHosts(t *testing.T) {
	testable, vHosts, _ := initTestData(t)

	foundVHosts, err := testable.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Equal(t, len(vHosts), len(foundVHosts))
	for i, foundHost := range foundVHosts {
		assert.Equal(t, vHosts[i], foundHost)
	}
}

func TestFindAllVirtualHostsDomain(t *testing.T) {
	testable, _, vHostDomains := initTestData(t)

	foundVHostDomains, err := testable.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Equal(t, len(vHostDomains), len(foundVHostDomains))
	for i, foundDomain := range foundVHostDomains {
		assert.Equal(t, vHostDomains[i], foundDomain)
	}
}

func TestFindVirtualHostDomainsByHost(t *testing.T) {
	testable, _, vHostDomains := initTestData(t)

	foundVHostDomains, err := testable.FindVirtualHostDomainsByHost(vHostDomains[0].Domain)
	assert.Nil(t, err)
	assert.Equal(t, len(vHostDomains), len(foundVHostDomains))
	for i, domain := range foundVHostDomains {
		assert.Equal(t, vHostDomains[i], domain)
	}

	foundVHostDomains, err = testable.FindVirtualHostDomainsByHost("non-existing")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundVHostDomains))
}

func TestFindFirstVirtualHostByRouteConfigurationId(t *testing.T) {
	testable, vHosts, _ := initTestData(t)

	foundVHost, err := testable.FindFirstVirtualHostByRouteConfigurationId(vHosts[0].RouteConfigurationId)
	assert.Nil(t, err)
	assert.Equal(t, vHosts[0], foundVHost)

	foundVHost, err = testable.FindFirstVirtualHostByRouteConfigurationId(-1)
	assert.Nil(t, err)
	assert.Nil(t, foundVHost)
}

func TestFindFirstVirtualHostByNameAndRouteConfigurationId(t *testing.T) {
	testable, vHosts, _ := initTestData(t)

	foundVHost, err := testable.FindFirstVirtualHostByNameAndRouteConfigurationId(vHosts[0].Name, vHosts[0].RouteConfigurationId)
	assert.Nil(t, err)
	assert.Equal(t, vHosts[0], foundVHost)

	foundVHost, err = testable.FindFirstVirtualHostByNameAndRouteConfigurationId("non-existing", vHosts[0].RouteConfigurationId)
	assert.Nil(t, err)
	assert.Nil(t, foundVHost)
}

func TestInMemDao_FindVirtualHostById(t *testing.T) {
	testable, vHosts, vHostDomains := initTestData(t)
	foundVHosts, err := testable.FindVirtualHostsByRouteConfigurationId(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundVHosts))

	foundVHost, err := testable.FindVirtualHostById(2)
	assert.Nil(t, err)
	assert.Equal(t, vHosts[1], foundVHost)

	foundVHost, err = testable.FindVirtualHostById(-1)
	assert.Nil(t, err)
	assert.Nil(t, foundVHost)

	foundVHostDomains, err := testable.FindVirtualHostDomainByVirtualHostId(2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundVHostDomains))
	assert.Equal(t, vHostDomains[1], foundVHostDomains[0])

	foundVHostDomains, err = testable.FindVirtualHostDomainByVirtualHostId(-1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundVHostDomains))
}

func initTestData(t *testing.T) (*InMemRepo, []*domain.VirtualHost, []*domain.VirtualHostDomain) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	vHosts := []*domain.VirtualHost{
		{
			Id:                   1,
			Name:                 "virtualHost1",
			RouteConfigurationId: 1,
		},
		{
			Id:                   2,
			Name:                 "virtualHost2",
			RouteConfigurationId: 1,
		},
		{
			Id:                   3,
			Name:                 "virtualHost3",
			RouteConfigurationId: 2,
		},
	}
	vHostDomains := []*domain.VirtualHostDomain{
		{
			Domain:        "*",
			VirtualHostId: 1,
		},
		{
			Domain:        "*",
			VirtualHostId: 2,
		},
		{
			Domain:        "*",
			VirtualHostId: 3,
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, vHost := range vHosts {
			assert.Nil(t, dao.SaveVirtualHost(vHost))
		}
		for _, vHostDomain := range vHostDomains {
			assert.Nil(t, dao.SaveVirtualHostDomain(vHostDomain))
		}
		return nil
	})
	assert.Nil(t, err)

	return testable, vHosts, vHostDomains
}
