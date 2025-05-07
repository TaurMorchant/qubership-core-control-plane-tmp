package entity

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
)

func (srv *Service) DeleteVirtualServiceByNodeGroupAndName(dao dao.Repository, nodeGroup, virtualServiceName string) error {
	virtualHost, err := srv.LoadVirtualHostRelationByNameAndNodeGroup(dao, nodeGroup, virtualServiceName)
	if err != nil {
		logger.Errorf("Failed to find virtual host with name %s and node group %s", virtualServiceName, nodeGroup)
		return err
	}
	clustersMap := make(map[string]bool)
	for _, route := range virtualHost.Routes {
		err := srv.DeleteRouteCascade(dao, route)
		if err != nil {
			logger.Errorf("Failed to delete routes for virtual host with name %s and node group %s", virtualServiceName, nodeGroup)
			return err
		}
		clustersMap[route.ClusterName] = false
	}
	if err := srv.DeleteVirtualHostDomainsByVirtualHost(dao, virtualHost); err != nil {
		logger.Errorf("Failed to delete virtual host domains for virtual host with name %s and node group %s:\n %v", virtualServiceName, nodeGroup, err)
		return err
	}
	if err := dao.DeleteVirtualHost(virtualHost); err != nil {
		logger.Errorf("Failed to delete virtual host with name %s and node group %s:\n %v", virtualServiceName, nodeGroup, err)
		return err
	}
	for clusterName, _ := range clustersMap {
		if routes, err := dao.FindRoutesByClusterName(clusterName); err != nil {
			logger.Errorf("Failed to find routes by cluster name %s:\n %v", clusterName, err)
			return err
		} else if len(routes) > 0 {
			continue
		}
		if err := srv.DeleteClusterCascadeByName(dao, clusterName); err != nil {
			logger.Errorf("Failed to delete cluster with name %s:\n %v", clusterName, err)
			return err
		}
	}
	return nil
}
