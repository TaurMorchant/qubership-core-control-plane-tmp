package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (d *InMemRepo) DeleteDeploymentVersion(dVersion *domain.DeploymentVersion) error {
	return d.DeleteById(domain.DeploymentVersionTable, dVersion.Version)
}

func (d *InMemRepo) DeleteDeploymentVersions(dVersions []*domain.DeploymentVersion) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	for _, dVersion := range dVersions {
		err := d.DeleteDeploymentVersion(dVersion)
		if err != nil {
			return nil
		}
	}
	return nil
}

func (d *InMemRepo) FindAllDeploymentVersions() ([]*domain.DeploymentVersion, error) {
	return FindAll[domain.DeploymentVersion](d, domain.DeploymentVersionTable)
}

func (d *InMemRepo) FindDeploymentVersionsByStage(stage string) ([]*domain.DeploymentVersion, error) {
	return FindByIndex[domain.DeploymentVersion](d, domain.DeploymentVersionTable, "stage", stage)
}

func (d *InMemRepo) FindDeploymentVersion(version string) (*domain.DeploymentVersion, error) {
	return FindById[domain.DeploymentVersion](d, domain.DeploymentVersionTable, version)
}

func (d *InMemRepo) SaveDeploymentVersion(dVersion *domain.DeploymentVersion) error {
	return d.SaveEntity(domain.DeploymentVersionTable, dVersion)
}
