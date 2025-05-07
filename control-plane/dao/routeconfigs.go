package dao

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (d *InMemRepo) FindRouteConfigById(routeConfigurationId int32) (*domain.RouteConfiguration, error) {
	return FindById[domain.RouteConfiguration](d, domain.RouteConfigurationTable, routeConfigurationId)
}

func (d *InMemRepo) SaveRouteConfig(routeConfig *domain.RouteConfiguration) error {
	return d.SaveUnique(domain.RouteConfigurationTable, routeConfig)
}

func (d *InMemRepo) FindAllRouteConfigs() ([]*domain.RouteConfiguration, error) {
	return FindAll[domain.RouteConfiguration](d, domain.RouteConfigurationTable)
}

func (d *InMemRepo) FindRouteConfigByNodeGroupIdAndName(nodeGroupId, name string) (*domain.RouteConfiguration, error) {
	return FindFirstByIndex[domain.RouteConfiguration](d, domain.RouteConfigurationTable, "nodeGroupAndName", nodeGroupId, name)
}

func (d *InMemRepo) FindRouteConfigsByNodeGroupId(nodeGroupId string) ([]*domain.RouteConfiguration, error) {
	return FindByIndex[domain.RouteConfiguration](d, domain.RouteConfigurationTable, "nodeGroup", nodeGroupId)
}

func (d *InMemRepo) FindRouteConfigsByRouteDeploymentVersion(deploymentVersion string) ([]*domain.RouteConfiguration, error) {
	routes, err := d.FindRoutesByDeploymentVersion(deploymentVersion)
	if err != nil {
		return nil, err
	}

	var foundRouteConfigs []*domain.RouteConfiguration
	// used for go-like Set implementation:
	// RouteConfiguration ID presence in map as key indicates that this RouteConfiguration is already present in resulting slice
	foundRouteConfigsSet := make(map[int32]bool)
	for _, r := range routes {
		vhost, err := d.FindVirtualHostById(r.VirtualHostId)
		if err != nil {
			return nil, err
		}
		if vhost == nil {
			return nil, fmt.Errorf("can not find VirtualHost by ID: VirtualHostId=%d", r.VirtualHostId)
		}
		rc, err := d.FindRouteConfigById(vhost.RouteConfigurationId)
		if err != nil {
			return nil, err
		}
		if rc == nil {
			return nil, fmt.Errorf("can not find RouteConfig by ID: RouteConfigId=%d", vhost.RouteConfigurationId)
		}
		if _, value := foundRouteConfigsSet[rc.Id]; !value {
			foundRouteConfigsSet[rc.Id] = true
			foundRouteConfigs = append(foundRouteConfigs, rc)
		}
	}

	return foundRouteConfigs, err
}

func (d *InMemRepo) FindRouteConfigsByEndpoint(endpoint *domain.Endpoint) ([]*domain.RouteConfiguration, error) {
	cluster, err := d.FindClusterById(endpoint.ClusterId)
	if err != nil || cluster == nil {
		return nil, err
	}

	routes, err := d.FindRoutesByClusterNameAndDeploymentVersion(cluster.Name, endpoint.DeploymentVersion)
	if err != nil {
		return nil, err
	}

	var routeConfigs []*domain.RouteConfiguration
	// used for go-like Set implementation:
	// RouteConfiguration ID presence in map as key indicates that this RouteConfiguration is already present in resulting slice
	foundRouteConfigsSet := make(map[int32]bool)
	for _, route := range routes {
		vhost, err := d.FindVirtualHostById(route.VirtualHostId)
		if err != nil {
			return nil, err
		}
		if vhost == nil {
			return nil, fmt.Errorf("can not find VirtualHost by ID: VirtualHostId=%d", route.VirtualHostId)
		}
		routeConfig, err := d.FindRouteConfigById(vhost.RouteConfigurationId)
		if err != nil {
			return nil, err
		}
		if routeConfig == nil {
			return nil, fmt.Errorf("can not find RouteConfig by ID: RouteConfigId=%d", vhost.RouteConfigurationId)
		}
		if _, value := foundRouteConfigsSet[routeConfig.Id]; !value {
			foundRouteConfigsSet[routeConfig.Id] = true
			routeConfigs = append(routeConfigs, routeConfig)
		}
	}

	return routeConfigs, nil
}

func (d *InMemRepo) DeleteRouteConfigById(id int32) error {
	return d.DeleteById(domain.RouteConfigurationTable, id)
}
