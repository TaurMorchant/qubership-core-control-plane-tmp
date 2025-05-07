package entity

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

func (srv *Service) PutHealthCheck(dao dao.Repository, healthCheck *domain.HealthCheck) error {
	existingHealthChecks, err := dao.FindHealthChecksByClusterId(healthCheck.ClusterId)
	if err != nil {
		return err
	}
	if existingHealthChecks != nil {
		if _, err := dao.DeleteHealthChecksByClusterId(healthCheck.ClusterId); err != nil {
			return err
		}
	}
	return dao.SaveHealthCheck(healthCheck)
}

func (srv *Service) FindHealthChecksByClusterId(dao dao.Repository, clusterId int32) ([]*domain.HealthCheck, error) {
	healthChecks, err := dao.FindHealthChecksByClusterId(clusterId)
	if err != nil {
		logger.Errorf("Failed to find healthchecks by cluster id %v: %v", clusterId, err)
		return nil, err
	}
	return healthChecks, nil
}
