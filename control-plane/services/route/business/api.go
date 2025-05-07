package business

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

type VersionService interface {
	GetOrCreateDeploymentVersion(dao dao.Repository, deploymentVersion string) (*domain.DeploymentVersion, error)
	GetActiveDeploymentVersion(dao dao.Repository) (*domain.DeploymentVersion, error)
}
