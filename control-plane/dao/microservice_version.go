package dao

import (
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
)

func (d *InMemRepo) FindMicroserviceVersionByNameAndInitialVersion(name string, namespace msaddr.Namespace, initialVersion string) (*domain.MicroserviceVersion, error) {
	return FindById[domain.MicroserviceVersion](d, domain.MicroserviceVersionTable, name, namespace.GetNamespace(), initialVersion)
}

func (d *InMemRepo) FindMicroserviceVersionsByVersion(version *domain.DeploymentVersion) ([]*domain.MicroserviceVersion, error) {
	if version == nil || version.Version == "" {
		return nil, errors.New("dao: FindMicroserviceVersionsByVersion got invalid empty version argument")
	}
	msVersions, err := d.FindAllMicroserviceVersions()
	result := make([]*domain.MicroserviceVersion, 0, 10)
	if err != nil {
		return nil, err
	}
	for _, msVersion := range msVersions {
		if msVersion.DeploymentVersion == version.Version {
			result = append(result, msVersion)
		}
	}
	return result, nil
}

func (d *InMemRepo) FindMicroserviceVersionsByNameAndNamespace(name string, namespace msaddr.Namespace) ([]*domain.MicroserviceVersion, error) {
	msVersions, err := d.FindAllMicroserviceVersions()
	result := make([]*domain.MicroserviceVersion, 0, 10)
	if err != nil {
		return nil, err
	}
	namespaceString := namespace.GetNamespace()
	for _, msVersion := range msVersions {
		if msVersion.Name == name && msVersion.Namespace == namespaceString {
			result = append(result, msVersion)
		}
	}
	return result, nil
}

func (d *InMemRepo) SaveMicroserviceVersion(msVersion *domain.MicroserviceVersion) error {
	return d.SaveEntity(domain.MicroserviceVersionTable, msVersion)
}

func (d *InMemRepo) DeleteMicroserviceVersion(name string, namespace msaddr.Namespace, initialVersion string) error {
	return d.DeleteById(domain.MicroserviceVersionTable, name, namespace.GetNamespace(), initialVersion)
}

func (d *InMemRepo) FindAllMicroserviceVersions() ([]*domain.MicroserviceVersion, error) {
	return FindAll[domain.MicroserviceVersion](d, domain.MicroserviceVersionTable)
}
