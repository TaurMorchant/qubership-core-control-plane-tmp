package entity

import (
	"errors"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/util"
)

func (srv *Service) PutVirtualHost(dao dao.Repository, virtualHost *domain.VirtualHost) error {
	var existing *domain.VirtualHost
	var err error
	if virtualHost.Id == 0 {
		existing, err = dao.FindFirstVirtualHostByNameAndRouteConfigurationId(virtualHost.Name, virtualHost.RouteConfigurationId)
		if err != nil {
			logger.Errorf("Error while searching for existing virtual host by RouteConfigurationId: %v", err)
			return err
		}
		if existing != nil {
			virtualHost.Id = existing.Id
		}
	} else {
		existing, err = dao.FindVirtualHostById(virtualHost.Id)
		if err != nil {
			logger.Errorf("Error while searching for existing virtual host by id: %v", err)
			return err
		}
	}
	if existing != nil {
		if err := srv.mergeVirtualHosts(dao, virtualHost, existing); err != nil {
			logger.Errorf("Error while merging virtual host domains: %v", err)
			return err
		}
	}
	if err := dao.SaveVirtualHost(virtualHost); err != nil {
		logger.Errorf("Error while saving virtual host: %v", err)
		return err
	}
	if len(virtualHost.Domains) > 0 {
		return srv.SaveVirtualHostDomains(dao, virtualHost.Domains, virtualHost.Id)
	}
	return nil
}

func (srv *Service) mergeVirtualHosts(dao dao.Repository, newVirtualHost *domain.VirtualHost, existingVirtualHost *domain.VirtualHost) error {
	existingDomains, err := dao.FindVirtualHostDomainByVirtualHostId(existingVirtualHost.Id)
	if err != nil {
		return err
	}

	for _, vhDomain := range newVirtualHost.Domains {
		vhDomain.VirtualHostId = existingVirtualHost.Id
	}
	newVirtualHost.Domains = util.MergeVirtualHostDomainsSlices(existingDomains, newVirtualHost.Domains)
	return nil
}

func (srv *Service) DeleteVirtualHostDomains(dao dao.Repository, domains []*domain.VirtualHostDomain) error {
	for _, domainToDelete := range domains {
		if err := dao.DeleteVirtualHostsDomain(domainToDelete); err != nil {
			return err
		}
	}
	return nil
}

func (srv *Service) DeleteVirtualHostDomainsByVirtualHost(dao dao.Repository, virtualHost *domain.VirtualHost) error {
	virtualHostDomains, err := dao.FindVirtualHostDomainByVirtualHostId(virtualHost.Id)
	if err != nil {
		return err
	}
	for _, domainToDelete := range virtualHostDomains {
		if err := dao.DeleteVirtualHostsDomain(domainToDelete); err != nil {
			return err
		}
	}
	return nil
}

func (srv *Service) SaveVirtualHostDomains(dao dao.Repository, domains []*domain.VirtualHostDomain, virtualHostId int32) error {
	for _, domainToSave := range domains {
		domainToSave.VirtualHostId = virtualHostId

		if err := srv.validateVirtualHostDomain(dao, domainToSave, virtualHostId); err != nil {
			logger.Errorf("Virtual host domain %s did not pass validation: %v", domainToSave.Domain, err)
			return err
		}

		if err := dao.SaveVirtualHostDomain(domainToSave); err != nil {
			return err
		}
	}
	return nil
}

// validateVirtualHostDomain validates that domain host is unique inside the nodeGroup
func (srv *Service) validateVirtualHostDomain(dao dao.Repository, domainToSave *domain.VirtualHostDomain, virtualHostId int32) error {
	existingDomains, err := dao.FindVirtualHostDomainsByHost(domainToSave.Domain)
	if err != nil {
		return err
	}
	for _, existingDomain := range existingDomains {
		if domainToSave.Domain == existingDomain.Domain && domainToSave.VirtualHostId != existingDomain.VirtualHostId {
			routeConfig, err := srv.FindRouteConfigurationByVirtualHostId(dao, virtualHostId)
			if err != nil {
				return err
			}
			anotherRouteConfig, err := srv.FindRouteConfigurationByVirtualHostId(dao, existingDomain.VirtualHostId)
			if err != nil {
				return err
			}
			if routeConfig.NodeGroupId == anotherRouteConfig.NodeGroupId {
				return errors.New(fmt.Sprintf("entity: node group %s already contains another virtual service with domain %s", routeConfig.NodeGroupId, domainToSave.Domain))
			}
		}
	}
	return nil
}

func (srv *Service) FindVirtualHostsByRouteConfig(dao dao.Repository, routeConfigId int32) ([]*domain.VirtualHost, error) {
	virtualHosts, err := dao.FindVirtualHostsByRouteConfigurationId(routeConfigId)
	if err != nil {
		logger.Errorf("Failed to find virtual hosts by route config id %d : %v", routeConfigId, err)
		return nil, err
	}
	for _, virtualHost := range virtualHosts {
		if virtualHost, err = srv.LoadVirtualHostRelations(dao, virtualHost); err != nil {
			logger.Errorf("Failed to load virtual host relations: %v", err)
			return nil, err
		}
	}

	return virtualHosts, nil
}

func (srv *Service) LoadVirtualHostRelations(dao dao.Repository, virtualHost *domain.VirtualHost) (*domain.VirtualHost, error) {
	var err error
	virtualHost.Domains, err = dao.FindVirtualHostDomainByVirtualHostId(virtualHost.Id)
	if err != nil {
		logger.Errorf("Failed to find virtual host domain by virtual host id %d : %v", virtualHost.Id, err)
		return nil, err
	}
	virtualHost.Routes, err = dao.FindRoutesByVirtualHostId(virtualHost.Id)
	if err != nil {
		logger.Errorf("Failed to find routes by virtual host id %d : %v", virtualHost.Id, err)
		return nil, err
	}
	for _, route := range virtualHost.Routes {
		if route, err = srv.LoadRouteRelations(dao, route); err != nil {
			logger.Errorf("Failed to load route relations: %v", err)
			return nil, err
		}
	}
	virtualHost.RateLimit, err = dao.FindRateLimitByNameWithHighestPriority(virtualHost.RateLimitId)
	if err != nil {
		logger.Errorf("Failed to find rate limit by virtual host:\n %v", err)
		return nil, err
	}
	return virtualHost, nil
}

func (srv *Service) FindVirtualHostByNameAndNodeGroup(dao dao.Repository, nodeGroup, virtualService string) (*domain.VirtualHost, error) {
	if ng, err := dao.FindNodeGroupByName(nodeGroup); err != nil {
		logger.Errorf("Failed to find node group %s", nodeGroup)
		return nil, err
	} else if ng == nil {
		return nil, errors.New(fmt.Sprintf("Nodegroup %s not found", nodeGroup))
	}
	routeConfigs, err := dao.FindRouteConfigsByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Errorf("Failed to find route configuration for node group %s", nodeGroup)
		return nil, err
	}
	for _, routeConfig := range routeConfigs {
		virtualHosts, err := dao.FindVirtualHostsByRouteConfigurationId(routeConfig.Id)
		if err != nil {
			logger.Errorf("Failed to find virtual hosts for route configuration id %d and node group %s", routeConfig.Id, nodeGroup)
			return nil, err
		}
		for _, virtualHost := range virtualHosts {
			if virtualHost.Name == virtualService {
				return virtualHost, nil
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("virtualService %s for node group %s not found", virtualService, nodeGroup))
}

func (srv *Service) LoadVirtualHostRelationByNameAndNodeGroup(dao dao.Repository, nodeGroup, virtualService string) (*domain.VirtualHost, error) {
	virtualHost, err := srv.FindVirtualHostByNameAndNodeGroup(dao, nodeGroup, virtualService)
	if err != nil {
		logger.Errorf("Failed to find virtual host %s for node group %s", virtualService, nodeGroup)
		return nil, err
	}
	return srv.LoadVirtualHostRelations(dao, virtualHost)
}

func (srv *Service) FindVirtualHostsByNodeGroup(dao dao.Repository, nodeGroup string) ([]*domain.VirtualHost, error) {
	routeConfigs, err := dao.FindRouteConfigsByNodeGroupId(nodeGroup)
	if err != nil {
		logger.Errorf("Failed to find route configs for node group %s:\n %v", nodeGroup, err)
		return nil, err
	}
	result := make([]*domain.VirtualHost, 0, len(routeConfigs))
	for _, routeConfig := range routeConfigs {
		virtualHosts, err := srv.FindVirtualHostsByRouteConfig(dao, routeConfig.Id)
		if err != nil {
			logger.Errorf("Failed to find virtual hosts for route config id %d:\n %v", routeConfig.Id, err)
			return nil, err
		}
		result = append(result, virtualHosts...)
	}

	return result, nil
}
