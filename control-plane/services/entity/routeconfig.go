package entity

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (srv *Service) PutRouteConfig(dao dao.Repository, routeConfig *domain.RouteConfiguration) error {
	if routeConfig.Id == 0 {
		existing, err := dao.FindRouteConfigByNodeGroupIdAndName(routeConfig.NodeGroupId, routeConfig.Name)
		if err != nil {
			logger.Errorf("Error while searching for existing route config with node group %v and name %v: %v", routeConfig.NodeGroupId, routeConfig.Name, err)
			return err
		}
		if existing != nil {
			routeConfig.Id = existing.Id
		}
	}
	return dao.SaveRouteConfig(routeConfig)
}

func (srv *Service) GetRouteConfigurationsWithRelations(dao dao.Repository) ([]*domain.RouteConfiguration, error) {
	routeConfigs, err := dao.FindAllRouteConfigs()
	if err != nil {
		logger.Errorf("Failed to find all route configs %v", err)
		return nil, err
	}
	for _, routeConfig := range routeConfigs {
		routeConfig.VirtualHosts, err = srv.FindVirtualHostsByRouteConfig(dao, routeConfig.Id)
		if err != nil {
			logger.Errorf("Failed to load virtual hosts by route config id %v: %v", routeConfig.Id, err)
			return nil, err
		}

	}
	return routeConfigs, nil
}

func (srv *Service) FindRouteConfigurationByVirtualHostId(dao dao.Repository, virtualHostId int32) (*domain.RouteConfiguration, error) {
	vhost, err := dao.FindVirtualHostById(virtualHostId)
	if err != nil {
		logger.Errorf("Failed to load virtual host by id %d during FindRouteConfigurationByVirtualHostId: %v", virtualHostId, err)
		return nil, err
	}
	routeConfig, err := dao.FindRouteConfigById(vhost.RouteConfigurationId)
	if err != nil {
		logger.Errorf("Failed to load route config by id %d during FindRouteConfigurationByVirtualHostId: %v", vhost.RouteConfigurationId, err)
		return nil, err
	}
	return routeConfig, nil
}
