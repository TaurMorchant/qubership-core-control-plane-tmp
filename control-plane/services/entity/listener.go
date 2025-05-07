package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (srv *Service) PutListener(dao dao.Repository, listener *domain.Listener) error {
	if listener.Id == 0 {
		existing, err := dao.FindListenerByNodeGroupIdAndName(listener.NodeGroupId, listener.Name)
		if err != nil {
			logger.Errorf("Error while trying to find existing listener for node group %v and name %v: %v", listener.NodeGroupId, listener.Name, err)
			return err
		}
		if existing != nil {
			listener.Id = existing.Id
		}
	}
	return dao.SaveListener(listener)
}

func (srv *Service) FindListenersByVirtualHostId(dao dao.Repository, virtualHostId int32) ([]*domain.Listener, error) {
	if virtualHostId == 0 {
		return nil, nil
	}
	routeConfig, err := srv.FindRouteConfigurationByVirtualHostId(dao, virtualHostId)
	if err != nil {
		logger.Errorf("Error while trying to find route config for virtual host id %v:\n %v", virtualHostId, err)
		return nil, err
	}
	return srv.FindListenersByRouteConfiguration(dao, routeConfig)
}

func (srv *Service) FindListenersByRouteConfiguration(dao dao.Repository, routeConfig *domain.RouteConfiguration) ([]*domain.Listener, error) {
	if routeConfig == nil {
		return nil, nil
	}
	listeners, err := dao.FindListenersByNodeGroupId(routeConfig.NodeGroupId)
	if err != nil {
		logger.Errorf("Error while trying to find listeners for node group id %v:\n %v", routeConfig.NodeGroupId, err)
		return nil, err
	}
	result := make([]*domain.Listener, 0, 1)
	for _, listener := range listeners {
		if listener.RouteConfigurationName == routeConfig.Name {
			result = append(result, listener)
		}
	}
	return result, nil
}
