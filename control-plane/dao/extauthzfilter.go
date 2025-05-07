package dao

import "github.com/netcracker/qubership-core-control-plane/domain"

func (d *InMemRepo) SaveExtAuthzFilter(filter *domain.ExtAuthzFilter) error {
	return d.SaveEntity(domain.ExtAuthzFilterTable, filter)
}

func (d *InMemRepo) FindExtAuthzFilterByNodeGroup(nodeGroup string) (*domain.ExtAuthzFilter, error) {
	return FindFirstByIndex[domain.ExtAuthzFilter](d, domain.ExtAuthzFilterTable, "nodeGroup", nodeGroup)
}

func (d *InMemRepo) FindAllExtAuthzFilters() ([]*domain.ExtAuthzFilter, error) {
	return FindAll[domain.ExtAuthzFilter](d, domain.ExtAuthzFilterTable)
}

func (d *InMemRepo) DeleteExtAuthzFilter(extAuthzFilterName string) error {
	return d.DeleteById(domain.ExtAuthzFilterTable, extAuthzFilterName)
}

func (d *InMemRepo) FindExtAuthzFilterByName(id string) (*domain.ExtAuthzFilter, error) {
	return FindById[domain.ExtAuthzFilter](d, domain.ExtAuthzFilterTable, id)
}
