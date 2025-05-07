package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
)

func (d *InMemRepo) FindHashPolicyByRouteId(routeId int32) ([]*domain.HashPolicy, error) {
	return FindByIndex[domain.HashPolicy](d, domain.HashPolicyTable, "routeId", routeId)
}

func (d *InMemRepo) FindHashPolicyByClusterAndVersions(clusterName string, versions ...string) ([]*domain.HashPolicy, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	endpoints, err := d.FindEndpointsByClusterName(clusterName)
	if err != nil {
		return nil, err
	}
	endpointsIds := extractIdWithVersion(endpoints, versions)
	hashPolicies := make([]*domain.HashPolicy, 0)
	for _, endpointId := range endpointsIds {
		if hashPolicy, err := d.storage.FindByIndex(txCtx.tx, domain.HashPolicyTable, "endpointId", endpointId); err == nil {
			if hashPolicy != nil {
				hashPolicies = append(hashPolicies, hashPolicy.([]*domain.HashPolicy)...)
			}
		} else {
			return nil, err
		}
	}
	return hashPolicies, nil
}

func extractIdWithVersion(endpoints []*domain.Endpoint, versions []string) []interface{} {
	endpointsIds := make([]interface{}, 0)
	for _, endpoint := range endpoints {
		if util.SliceContainsElement(versions, endpoint.DeploymentVersion) {
			endpointsIds = append(endpointsIds, endpoint.Id)
		}
	}
	return endpointsIds
}

func (d *InMemRepo) SaveHashPolicy(hashPolicy *domain.HashPolicy) error {
	return d.SaveUnique(domain.HashPolicyTable, hashPolicy)
}

func (d *InMemRepo) DeleteHashPolicyById(id int32) error {
	return d.DeleteById(domain.HashPolicyTable, id)
}

func (d *InMemRepo) DeleteHashPolicyByRouteId(routeId int32) (int, error) {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return txCtx.tx.DeleteAll(domain.HashPolicyTable, "routeId", routeId)
}

func (d *InMemRepo) DeleteHashPolicyByEndpointId(endpointId int32) (int, error) {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return txCtx.tx.DeleteAll(domain.HashPolicyTable, "endpointId", endpointId)
}

func (d *InMemRepo) FindHashPolicyByEndpointId(endpointId int32) ([]*domain.HashPolicy, error) {
	return FindByIndex[domain.HashPolicy](d, domain.HashPolicyTable, "endpointId", endpointId)
}
