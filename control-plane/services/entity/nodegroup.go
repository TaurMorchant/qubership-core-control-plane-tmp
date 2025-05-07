package entity

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (srv *Service) CreateOrUpdateNodeGroup(dao dao.Repository, nodeGroup domain.NodeGroup) (domain.NodeGroup, error) {
	existingNodeGroup, err := dao.FindNodeGroupByName(nodeGroup.Name)
	if err != nil {
		return domain.NodeGroup{}, err
	}
	if existingNodeGroup == nil {
		return nodeGroup, dao.SaveNodeGroup(&nodeGroup)
	}
	if nodeGroup.GatewayType == "" {
		return nodeGroup, nil
	}
	if nodeGroup.GatewayType != existingNodeGroup.GatewayType || nodeGroup.ForbidVirtualHosts != existingNodeGroup.ForbidVirtualHosts {
		return nodeGroup, dao.SaveNodeGroup(&nodeGroup)
	}
	return *existingNodeGroup, nil
}

func (srv *Service) GenerateEnvoyEntityVersions(ctx context.Context, repo dao.Repository, gateway string, entities ...string) error {
	for _, entity := range entities {
		if err := srv.generateEnvoyEntityVersion(ctx, repo, gateway, entity); err != nil {
			return err
		}
	}
	return nil
}

func (srv *Service) generateEnvoyEntityVersion(ctx context.Context, repo dao.Repository, gateway, entity string) error {
	version := domain.NewEnvoyConfigVersion(gateway, entity)
	if err := repo.SaveEnvoyConfigVersion(version); err != nil {
		logger.ErrorC(ctx, "Saving new envoy config version for %s in %s has failed:\n %v", entity, gateway, err)
		return err
	}
	logger.InfoC(ctx, "Saved new envoyConfigVersion for nodeGroup %s and entity %s: %+v", gateway, entity, *version)
	return nil
}
