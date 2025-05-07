package factory

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/route/business"
)

type ComponentsFactory struct {
	entityService *entity.Service
}

func NewComponentsFactory(entityService *entity.Service) *ComponentsFactory {
	return &ComponentsFactory{entityService: entityService}
}

func (f *ComponentsFactory) GetRoutesAutoGenerator(dao dao.Repository) *business.RoutesAutoGenerator {
	return business.NewRoutesAutoGenerator(dao, f.entityService)
}
