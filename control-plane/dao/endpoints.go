package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

func (d *InMemRepo) DeleteEndpoint(endpoint *domain.Endpoint) error {
	return d.Delete(domain.EndpointTable, endpoint)
}

func (d *InMemRepo) FindEndpointById(endpointId int32) (*domain.Endpoint, error) {
	return FindById[domain.Endpoint](d, domain.EndpointTable, endpointId)
}

func (d *InMemRepo) FindEndpointsByClusterName(clusterName string) ([]*domain.Endpoint, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	cluster, err := d.FindClusterByName(clusterName)
	if err != nil {
		return nil, err
	}
	if cluster != nil {
		return d.FindEndpointsByClusterId(cluster.Id)
	}
	return nil, nil
}

func (d *InMemRepo) FindEndpointsByDeploymentVersion(version string) ([]*domain.Endpoint, error) {
	return FindByIndex[domain.Endpoint](d, domain.EndpointTable, "dVersion", version)
}

func (d *InMemRepo) FindEndpointsByAddressAndPortAndDeploymentVersion(address string, port int32, version string) ([]*domain.Endpoint, error) {
	return FindByIndex[domain.Endpoint](d, domain.EndpointTable, "addressAndPortAndDVersion", address, port, version)
}

func (d *InMemRepo) SaveEndpoint(endpoint *domain.Endpoint) error {
	return d.SaveUnique(domain.EndpointTable, endpoint)
}

func (d *InMemRepo) FindAllEndpoints() ([]*domain.Endpoint, error) {
	return FindAll[domain.Endpoint](d, domain.EndpointTable)
}

func (d *InMemRepo) FindEndpointsByClusterId(clusterId int32) ([]*domain.Endpoint, error) {
	return FindByIndex[domain.Endpoint](d, domain.EndpointTable, "clusterId", clusterId)
}

func (d *InMemRepo) FindEndpointByStatefulSession(statefulSessionId int32) (*domain.Endpoint, error) {
	return FindFirstByIndex[domain.Endpoint](d, domain.EndpointTable, "statefulSessionId", statefulSessionId)
}

func (d *InMemRepo) FindEndpointsByClusterIdAndDeploymentVersion(clusterId int32, dVersion *domain.DeploymentVersion) ([]*domain.Endpoint, error) {
	return FindByIndex[domain.Endpoint](d, domain.EndpointTable, "clusterIdAndDVersion", clusterId, dVersion.Version)
}

func (d *InMemRepo) FindEndpointsByDeploymentVersionsIn(dVersions []*domain.DeploymentVersion) ([]*domain.Endpoint, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	endpoints := make([]*domain.Endpoint, 0)
	for _, dVersion := range dVersions {
		if foundedEndpoints, err := d.FindEndpointsByDeploymentVersion(dVersion.Version); err == nil {
			if foundedEndpoints != nil {
				endpoints = append(endpoints, foundedEndpoints...)
			}
		} else {
			return nil, err
		}
	}
	return endpoints, nil
}
