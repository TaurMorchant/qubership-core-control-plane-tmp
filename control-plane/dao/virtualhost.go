package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (d *InMemRepo) SaveVirtualHost(virtualHost *domain.VirtualHost) error {
	return d.SaveUnique(domain.VirtualHostTable, virtualHost)
}

func (d *InMemRepo) SaveVirtualHostDomain(virtualHostDomain *domain.VirtualHostDomain) error {
	return d.SaveEntity(domain.VirtualHostDomainTable, virtualHostDomain)
}

func (d *InMemRepo) FindFirstVirtualHostByNameAndRouteConfigurationId(name string, id int32) (*domain.VirtualHost, error) {
	return FindFirstByIndex[domain.VirtualHost](d, domain.VirtualHostTable, "nameAndRouteConfigId", name, id)
}

func (d *InMemRepo) FindFirstVirtualHostByRouteConfigurationId(routeConfigId int32) (*domain.VirtualHost, error) {
	return FindFirstByIndex[domain.VirtualHost](d, domain.VirtualHostTable, "routeConfigId", routeConfigId)
}

func (d *InMemRepo) FindVirtualHostDomainByVirtualHostId(virtualHostId int32) ([]*domain.VirtualHostDomain, error) {
	return FindByIndex[domain.VirtualHostDomain](d, domain.VirtualHostDomainTable, "virtualHostId", virtualHostId)
}

func (d *InMemRepo) FindVirtualHostDomainsByHost(virtualHostDomain string) ([]*domain.VirtualHostDomain, error) {
	return FindByIndex[domain.VirtualHostDomain](d, domain.VirtualHostDomainTable, "domain", virtualHostDomain)
}

func (d *InMemRepo) FindVirtualHostById(virtualHostId int32) (*domain.VirtualHost, error) {
	return FindById[domain.VirtualHost](d, domain.VirtualHostTable, virtualHostId)
}

func (d *InMemRepo) FindAllVirtualHosts() ([]*domain.VirtualHost, error) {
	return FindAll[domain.VirtualHost](d, domain.VirtualHostTable)
}

func (d *InMemRepo) FindVirtualHostsByRouteConfigurationId(routeConfigId int32) ([]*domain.VirtualHost, error) {
	return FindByIndex[domain.VirtualHost](d, domain.VirtualHostTable, "routeConfigId", routeConfigId)
}

func (d *InMemRepo) FindAllVirtualHostsDomain() ([]*domain.VirtualHostDomain, error) {
	return FindAll[domain.VirtualHostDomain](d, domain.VirtualHostDomainTable)
}

func (d *InMemRepo) DeleteVirtualHostsDomain(virtualHostDomain *domain.VirtualHostDomain) error {
	return d.Delete(domain.VirtualHostDomainTable, virtualHostDomain)
}

func (d *InMemRepo) DeleteVirtualHost(vHost *domain.VirtualHost) error {
	return d.Delete(domain.VirtualHostTable, vHost)
}
