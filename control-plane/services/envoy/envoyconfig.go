package envoy

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"time"
)

type EnvoyConfigService struct {
	dao dao.Repository
}

func NewEnvoyConfigService(dao dao.Repository) *EnvoyConfigService {
	return &EnvoyConfigService{
		dao,
	}
}

func (envoy *EnvoyConfigService) GenerateAndSave(nodeGroup, entityType string) error {
	version := &domain.EnvoyConfigVersion{
		NodeGroup:  nodeGroup,
		EntityType: entityType,
		Version:    time.Now().UnixNano(),
	}
	err := envoy.dao.SaveEnvoyConfigVersion(version)
	if err != nil {
		return err
	}
	return nil
}

func UpdateAllResourceVersions(repo dao.Repository, nodeGroup string) error {
	if err := repo.SaveEnvoyConfigVersion(domain.NewEnvoyConfigVersion(nodeGroup, domain.ListenerTable)); err != nil {
		return err
	}
	if err := repo.SaveEnvoyConfigVersion(domain.NewEnvoyConfigVersion(nodeGroup, domain.ClusterTable)); err != nil {
		return err
	}
	return repo.SaveEnvoyConfigVersion(domain.NewEnvoyConfigVersion(nodeGroup, domain.RouteConfigurationTable))
}
