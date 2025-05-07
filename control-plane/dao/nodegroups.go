package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	errUtil "github.com/pkg/errors"
)

func (d *InMemRepo) FindAllNodeGroups() ([]*domain.NodeGroup, error) {
	return FindAll[domain.NodeGroup](d, domain.NodeGroupTable)
}

func (d *InMemRepo) FindNodeGroupByName(name string) (*domain.NodeGroup, error) {
	return FindById[domain.NodeGroup](d, domain.NodeGroupTable, name)
}

func (d *InMemRepo) SaveNodeGroup(nodeGroup *domain.NodeGroup) error {
	return d.SaveEntity(domain.NodeGroupTable, nodeGroup)
}

func (d *InMemRepo) FindNodeGroupsByCluster(cluster *domain.Cluster) ([]*domain.NodeGroup, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	clusterNodeGroups, err := d.storage.FindByIndex(txCtx.tx, domain.ClusterNodeGroupTable, "clustersId", cluster.Id)
	if err == nil {
		clusters := make([]*domain.NodeGroup, 0)
		for _, clusterNodeGroup := range clusterNodeGroups.([]*domain.ClustersNodeGroup) {
			nodeGroup, err := d.FindNodeGroupByName(clusterNodeGroup.NodegroupsName)
			if err != nil {
				return nil, errUtil.Wrapf(err, "Finding NodeGroup by name '%s' caused error", clusterNodeGroup.NodegroupsName)
			}
			clusters = append(clusters, nodeGroup)
		}
		return clusters, nil
	} else {
		return nil, errUtil.Wrapf(err, "Finding NodeGroup relations for Cluster '%v' caused error", cluster)
	}
}

func (d *InMemRepo) SaveClustersNodeGroup(relation *domain.ClustersNodeGroup) error {
	return d.SaveEntity(domain.ClusterNodeGroupTable, relation)
}

func (d *InMemRepo) FindClustersNodeGroup(relation *domain.ClustersNodeGroup) (*domain.ClustersNodeGroup, error) {
	return FindById[domain.ClustersNodeGroup](d, domain.ClusterNodeGroupTable, relation.ClustersId, relation.NodegroupsName)
}

func (d *InMemRepo) FindAllClusterWithNodeGroup() ([]*domain.ClustersNodeGroup, error) {
	return FindAll[domain.ClustersNodeGroup](d, domain.ClusterNodeGroupTable)
}

func (d *InMemRepo) DeleteClustersNodeGroupByClusterId(clusterId int32) (int, error) {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	return txCtx.tx.DeleteAll(domain.ClusterNodeGroupTable, "clustersId", clusterId)
}

func (d *InMemRepo) DeleteClustersNodeGroup(relation *domain.ClustersNodeGroup) error {
	return d.Delete(domain.ClusterNodeGroupTable, relation)
}

func (d *InMemRepo) DeleteNodeGroupByName(nodeGroupName string) error {
	return d.DeleteById(domain.NodeGroupTable, nodeGroupName)
}
