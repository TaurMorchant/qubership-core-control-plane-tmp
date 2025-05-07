package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	error_util "github.com/pkg/errors"
)

func (d *InMemRepo) FindClusterById(id int32) (*domain.Cluster, error) {
	return FindById[domain.Cluster](d, domain.ClusterTable, id)
}

func (d *InMemRepo) FindClusterByName(name string) (*domain.Cluster, error) {
	return FindFirstByIndex[domain.Cluster](d, domain.ClusterTable, "name", name)
}

func (d *InMemRepo) FindClustersByFamilyNameAndNamespace(familyName string, namespace msaddr.Namespace) ([]*domain.Cluster, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	clusterKeyPrefix := clusterkey.DefaultClusterKeyGenerator.BuildKeyPrefix(familyName, namespace)
	return FindByIndex[domain.Cluster](d, domain.ClusterTable, "name_prefix", clusterKeyPrefix)
}

func (d *InMemRepo) FindClusterByNodeGroup(group *domain.NodeGroup) ([]*domain.Cluster, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	clusterNodeGroups, err := d.storage.FindByIndex(txCtx.tx, domain.ClusterNodeGroupTable, "nodegroupsName", group.Name)
	if err == nil {
		clusters := make([]*domain.Cluster, 0)
		for _, clusterNodeGroup := range clusterNodeGroups.([]*domain.ClustersNodeGroup) {
			cluster, err := d.FindClusterById(clusterNodeGroup.ClustersId)
			if err != nil {
				return nil, err
			}
			clusters = append(clusters, cluster)
		}
		return clusters, nil
	} else {
		return nil, error_util.Wrapf(err, "Getting clustersNodeGroups caused error")
	}
}

func (d *InMemRepo) FindAllClusters() ([]*domain.Cluster, error) {
	return FindAll[domain.Cluster](d, domain.ClusterTable)
}

func (d *InMemRepo) SaveCluster(cluster *domain.Cluster) error {
	if cluster.HttpVersion == nil {
		var httpVersion int32 = 1
		cluster.HttpVersion = &httpVersion
	}
	return d.SaveUnique(domain.ClusterTable, cluster)
}

// For cascade operations must be special service which make it in transaction
func (d *InMemRepo) DeleteClusterByName(name string) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	_, err := txCtx.tx.DeleteAll(domain.ClusterTable, "name", name)
	if err != nil {
		return error_util.Wrapf(err, "Removing cluster with name '%s' caused error", name)
	}
	return nil
}

// For cascade operations must be special service which make it in transaction
func (d *InMemRepo) DeleteCluster(cluster *domain.Cluster) error {
	return d.Delete(domain.ClusterTable, cluster)
}

func (d *InMemRepo) FindClusterByEndpointIn(endpoints []*domain.Endpoint) ([]*domain.Cluster, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()

	var foundClusters []*domain.Cluster
	for _, e := range endpoints {
		c, err := d.FindClusterById(e.ClusterId)
		if err != nil {
			return nil, err
		}
		if c != nil {
			foundClusters = append(foundClusters, c)
		}
	}

	return foundClusters, nil
}
