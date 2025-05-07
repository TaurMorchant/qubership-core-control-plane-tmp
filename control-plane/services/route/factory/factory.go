package factory

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/business"
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
