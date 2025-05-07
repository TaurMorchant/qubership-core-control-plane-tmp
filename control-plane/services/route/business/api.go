package business

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

type VersionService interface {
	GetOrCreateDeploymentVersion(dao dao.Repository, deploymentVersion string) (*domain.DeploymentVersion, error)
	GetActiveDeploymentVersion(dao dao.Repository) (*domain.DeploymentVersion, error)
}
