package entity

import (
	"errors"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (srv *Service) PutEndpoint(dao dao.Repository, endpoint *domain.Endpoint) error {
	if endpoint.Id == 0 {
		existingEndpoints, err := dao.FindEndpointsByClusterId(endpoint.ClusterId)
		if err != nil {
			logger.Errorf("Error while searching for existing endpoint by cluster id: %v", err)
			return err
		}
		for _, existingEndpoint := range existingEndpoints {
			if endpoint.DeploymentVersion == existingEndpoint.InitialDeploymentVersion {
				endpoint.Id = existingEndpoint.Id
				endpoint.DeploymentVersion = existingEndpoint.DeploymentVersion
				endpoint.InitialDeploymentVersion = existingEndpoint.InitialDeploymentVersion
				break
			} else if endpoint.DeploymentVersion == existingEndpoint.DeploymentVersion {
				endpoint.Id = existingEndpoint.Id
				endpoint.InitialDeploymentVersion = existingEndpoint.InitialDeploymentVersion
				break
			}
		}
	}
	return dao.SaveEndpoint(endpoint)
}

func (srv *Service) PutEndpoints(dao dao.Repository, clusterId int32, endpoints []*domain.Endpoint) error {
	deploymentVersion := ""
	for _, endpoint := range endpoints {
		if deploymentVersion == "" {
			deploymentVersion = endpoint.DeploymentVersion
		} else if deploymentVersion != endpoint.DeploymentVersion {
			return errors.New(fmt.Sprintf("endpoints must have the same deployment version. Endpoints: %v", endpoints))
		}
	}
	existingEndpoints, err := dao.FindEndpointsByClusterIdAndDeploymentVersion(clusterId, &domain.DeploymentVersion{Version: deploymentVersion})
	if err != nil {
		logger.Errorf("Error while searching for existing endpoint by cluster id: %v", err)
		return err
	}
	// need to delete existing endpoints to replace them with new ones
	for _, endpoint := range existingEndpoints {
		err := srv.DeleteEndpointCascade(dao, endpoint)
		if err != nil {
			return err
		}
	}
	for _, endpoint := range endpoints {
		err := dao.SaveEndpoint(endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func (srv *Service) DeleteEndpointCascade(dao dao.Repository, endpoint *domain.Endpoint) error {
	if _, err := dao.DeleteHashPolicyByEndpointId(endpoint.Id); err != nil {
		logger.Errorf("Failed to delete endpoints hash policy during endpoint cascade deletion: %v", err)
		return err
	}
	if endpoint.StatefulSessionId != 0 {
		if err := dao.DeleteStatefulSessionConfig(endpoint.StatefulSessionId); err != nil {
			logger.Errorf("Failed to delete endpoints stateful session config during endpoint cascade deletion: %v", err)
			return err
		}
	}
	return dao.DeleteEndpoint(endpoint)
}

func (srv *Service) DeleteEndpointsCascade(dao dao.Repository, endpointsToDelete []*domain.Endpoint) error {
	for _, endpointToDelete := range endpointsToDelete {
		err := srv.DeleteEndpointCascade(dao, endpointToDelete)
		if err != nil {
			logger.Errorf("Can't delete endpoint %v: error:%v", endpointToDelete, err)
			return err
		}
		logger.Infof("Endpoint %v is deleted", endpointToDelete)
	}
	return nil
}

func (srv *Service) FindEndpointsByClusterId(dao dao.Repository, clusterId int32) ([]*domain.Endpoint, error) {
	endpoints, err := dao.FindEndpointsByClusterId(clusterId)
	if err != nil {
		logger.Errorf("Failed to find endpoints by cluster id %v: %v", clusterId, err)
		return nil, err
	}
	for _, endpoint := range endpoints {
		if endpoint, err = srv.LoadEndpointRelations(dao, endpoint); err != nil {
			logger.Errorf("Failed to load endpoint relations: %v", err)
			return nil, err
		}
	}
	return endpoints, nil
}

func (srv *Service) LoadEndpointRelations(dao dao.Repository, endpoint *domain.Endpoint) (*domain.Endpoint, error) {
	var err error
	endpoint.DeploymentVersionVal, err = dao.FindDeploymentVersion(endpoint.DeploymentVersion)
	if err != nil {
		logger.Errorf("Failed to load endpoint version %v from DAO: %v", endpoint.DeploymentVersion, err)
		return nil, err
	}
	endpoint.HashPolicies, err = dao.FindHashPolicyByEndpointId(endpoint.Id)
	if err != nil {
		logger.Errorf("Failed to load endpoint hash policies from DAO: %v", err)
		return nil, err
	}
	return endpoint, nil
}
