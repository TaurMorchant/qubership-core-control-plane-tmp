package dao

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"

func (d *InMemRepo) SaveCompositeSatellite(satellite *domain.CompositeSatellite) error {
	return d.SaveEntity(domain.CompositeSatelliteTable, satellite)
}

func (d *InMemRepo) DeleteCompositeSatellite(namespace string) error {
	return d.Delete(domain.CompositeSatelliteTable, &domain.CompositeSatellite{Namespace: namespace})
}

func (d *InMemRepo) FindAllCompositeSatellites() ([]*domain.CompositeSatellite, error) {
	return FindAll[domain.CompositeSatellite](d, domain.CompositeSatelliteTable)
}

func (d *InMemRepo) FindCompositeSatelliteByNamespace(namespace string) (*domain.CompositeSatellite, error) {
	return FindFirstByIndex[domain.CompositeSatellite](d, domain.CompositeSatelliteTable, "id", namespace)
}
