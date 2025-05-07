package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"

	"github.com/go-errors/errors"
)

func (srv *Service) PutCluster(dao dao.Repository, cluster *domain.Cluster) error {
	if cluster.Id == 0 {
		err := RebaseCluster(dao, cluster)
		if err != nil {
			logger.Errorf("Error while rebasing cluster with name %v: %v", cluster.Name, err)
			return err
		}
	}
	return dao.SaveCluster(cluster)
}

func (srv *Service) UpdateExistingClusterWithTcpKeepalive(dao dao.Repository, existing *domain.Cluster, newKeepalive *domain.TcpKeepalive) error {
	resKeepalive, err := srv.updateExistingClusterTcpKeepalive(dao, existing, newKeepalive)
	if err != nil {
		return err
	}
	if resKeepalive == nil {
		existing.TcpKeepalive = nil
		existing.TcpKeepaliveId = 0
	} else {
		existing.TcpKeepalive = resKeepalive
		existing.TcpKeepaliveId = resKeepalive.Id
	}
	return dao.SaveCluster(existing)
}

func (srv *Service) updateExistingClusterTcpKeepalive(dao dao.Repository, existing *domain.Cluster, newKeepalive *domain.TcpKeepalive) (*domain.TcpKeepalive, error) {
	if existing.TcpKeepaliveId != 0 {
		existingTcpKeepalive, err := dao.FindTcpKeepaliveById(existing.TcpKeepaliveId)
		if err != nil {
			logger.Errorf("Error while searching for existing tcp keepalive for cluster with name %v:\n %v", existing.Name, err)
			return nil, err
		}
		if existingTcpKeepalive == nil {
			err = errors.New("cluster tcp keepalive for provided ID is missing")
			logger.Errorf("Error while loading existing tcp keepalive for cluster with name %v:\n %v", existing.Name, err)
			return nil, err
		}

		if newKeepalive == nil {
			if err = dao.DeleteTcpKeepaliveById(existing.TcpKeepaliveId); err != nil {
				logger.Errorf("Error while deleting existing tcp keepalive for cluster with name %v: %v", existing.Name, err)
				return nil, err
			}
			return nil, nil
		} else {
			existingTcpKeepalive.Probes = newKeepalive.Probes
			existingTcpKeepalive.Time = newKeepalive.Time
			existingTcpKeepalive.Interval = newKeepalive.Interval
			if err = dao.SaveTcpKeepalive(existingTcpKeepalive); err != nil {
				logger.Errorf("Error while updating existing tcp keepalive for cluster with name %v: %v", existing.Name, err)
				return nil, err
			}
			return existingTcpKeepalive, nil
		}
	} else {
		if newKeepalive != nil {
			if err := dao.SaveTcpKeepalive(newKeepalive); err != nil {
				logger.Errorf("Error while saving tcp keepalive for cluster with name %v:\n %v", existing.Name, err)
				return nil, err
			}
			return newKeepalive, nil
		}
		return nil, nil
	}
}

func (srv *Service) UpdateClusterTcpKeepalive(dao dao.Repository, cluster *domain.Cluster) error {
	existing, err := dao.FindClusterByName(cluster.Name)
	if err != nil {
		logger.Errorf("Error while searching for existing cluster with name %v: %v", cluster.Name, err)
		return err
	}
	if existing == nil {
		if cluster.TcpKeepalive == nil {
			return nil
		}
		if err = srv.saveNewTcpKeepaliveForCluster(dao, cluster); err != nil {
			return err
		}
	} else {
		resKeepalive, err := srv.updateExistingClusterTcpKeepalive(dao, existing, cluster.TcpKeepalive)
		if err != nil {
			return err
		}
		if resKeepalive == nil {
			cluster.TcpKeepalive = nil
			cluster.TcpKeepaliveId = 0
		} else {
			cluster.TcpKeepalive = resKeepalive
			cluster.TcpKeepaliveId = resKeepalive.Id
		}
	}
	return nil
}

func (srv *Service) saveNewTcpKeepaliveForCluster(dao dao.Repository, cluster *domain.Cluster) error {
	if err := dao.SaveTcpKeepalive(cluster.TcpKeepalive); err != nil {
		logger.Errorf("Error while saving tcp keepalive for cluster with name %v:\n %v", cluster.Name, err)
		return err
	}
	cluster.TcpKeepaliveId = cluster.TcpKeepalive.Id
	return nil
}

func RebaseCluster(dao dao.Repository, cluster *domain.Cluster) error {
	existing, err := dao.FindClusterByName(cluster.Name)
	if err != nil {
		logger.Errorf("Error while searching for existing cluster with name %v: %v", cluster.Name, err)
		return err
	}
	if existing != nil {
		cluster.Id = existing.Id
		cluster.CircuitBreakerId = existing.CircuitBreakerId
		if existing.Version == cluster.Version {
			cluster.LbPolicy = existing.LbPolicy
			cluster.DiscoveryType = existing.DiscoveryType
			cluster.DiscoveryTypeOld = existing.DiscoveryTypeOld
		}
		if cluster.HttpVersion == nil {
			cluster.HttpVersion = existing.HttpVersion
		}
	}
	return nil
}

func (srv *Service) PutClustersNodeGroupIfAbsent(dao dao.Repository, relation *domain.ClustersNodeGroup) error {
	clustersNodeGroup, err := dao.FindClustersNodeGroup(relation)
	if err != nil {
		logger.Errorf("Error while get clusters Node Group %v: %v", relation, err)
		return err
	}
	if clustersNodeGroup != nil {
		logger.Infof("Node groups %s is already connected with cluster id %d.", relation.NodegroupsName, relation.ClustersId)
		return nil
	}
	err = dao.SaveClustersNodeGroup(relation)
	if err != nil {
		logger.Errorf("Error while saving clusters Node Group %v: %v", relation, err)
		return err
	}
	return nil
}

func (srv *Service) GetClustersWithRelations(dao dao.Repository) ([]*domain.Cluster, error) {
	clusters, err := dao.FindAllClusters()
	if err != nil {
		logger.Errorf("Failed to find all clusters %v", err)
		return nil, err
	}
	for _, cluster := range clusters {
		cluster, err = srv.GetClusterWithRelations(dao, cluster.Name)
		if err != nil {
			logger.Errorf("Failed to load cluster %v: %v", cluster, err)
			return nil, err
		}
	}
	return clusters, nil
}

func (srv *Service) GetClusterWithRelations(dao dao.Repository, clusterName string) (*domain.Cluster, error) {
	cluster, err := dao.FindClusterByName(clusterName)
	if err != nil {
		logger.Errorf("Failed to find cluster by name %v: %v", clusterName, err)
		return nil, err
	}
	if cluster == nil {
		logger.Infof("Cluster with name %s is not found", clusterName)
		return nil, nil
	}
	cluster.NodeGroups, err = dao.FindNodeGroupsByCluster(cluster)
	if err != nil {
		logger.Errorf("Failed to load node groups by cluster %v: %v", clusterName, err)
		return nil, err
	}
	cluster.Endpoints, err = srv.FindEndpointsByClusterId(dao, cluster.Id)
	if err != nil {
		logger.Errorf("Failed to load endpoints by cluster id %v: %v", cluster.Id, err)
		return nil, err
	}
	for _, endpoint := range cluster.Endpoints {
		endpoint.HashPolicies, err = dao.FindHashPolicyByEndpointId(endpoint.Id)
		if err != nil {
			logger.Errorf("Failed to load hashPolicies by endpoint id %v: %v", endpoint.Id, err)
		}
	}
	cluster.HealthChecks, err = srv.FindHealthChecksByClusterId(dao, cluster.Id)
	if err != nil {
		logger.Errorf("Failed to load healthChecks by cluster id %v: %v", cluster.Id, err)
		return nil, err
	}
	return cluster, nil
}

func (srv *Service) DeleteClusterCascade(dao dao.Repository, cluster *domain.Cluster) error {
	endpoints, err := dao.FindEndpointsByClusterId(cluster.Id)
	if err != nil {
		logger.Errorf("Failed to load cluster endpoints during cluster deletion: %v", err)
		return err
	}
	for _, endpoint := range endpoints {
		if err := srv.DeleteEndpointCascade(dao, endpoint); err != nil {
			logger.Errorf("Failed to delete endpoint during cluster cascade deletion: %v", err)
			return err
		}
	}

	routes, err := dao.FindRoutesByClusterName(cluster.Name)
	if err != nil {
		logger.Errorf("Failed to load cluster routes during cluster deletion: %v", err)
		return err
	}
	for _, route := range routes {
		if err := srv.DeleteRouteCascade(dao, route); err != nil {
			logger.Errorf("Failed to delete route during cluster cascade deletion: %v", err)
			return err
		}
	}
	if _, err = dao.DeleteHealthChecksByClusterId(cluster.Id); err != nil {
		logger.Errorf("Failed to delete health checks during cluster cascade deletion: %v", err)
		return err
	}
	_, err = dao.DeleteClustersNodeGroupByClusterId(cluster.Id)
	if err != nil {
		logger.Errorf("Failed to delete cluster node groups by cluster id: %v", err)
		return err
	}

	if err := srv.DeleteOrphanedStatefulSessionByCluster(dao, cluster); err != nil {
		logger.Errorf("Error during cluster stateful session deletion: %v", err)
		return err
	}

	if cluster.CircuitBreakerId != 0 {
		if err := srv.DeleteCircuitBreakerCascadeById(dao, cluster.CircuitBreakerId); err != nil {
			logger.Errorf("Error during cascade CircuitBreaker deletion: %v", err)
			return err
		}
	}

	if cluster.TcpKeepaliveId != 0 {
		if err := dao.DeleteTcpKeepaliveById(cluster.TcpKeepaliveId); err != nil {
			logger.Errorf("Error during cascade TcpKeepalive deletion: %v", err)
			return err
		}
	}

	if err := dao.DeleteCluster(cluster); err != nil {
		logger.Errorf("Error during cluster deletion: %v", err)
		return err
	}
	return err
}

func (srv *Service) DeleteClusterCascadeByName(dao dao.Repository, clusterName string) error {
	cluster, err := dao.FindClusterByName(clusterName)
	if err != nil {
		return err
	}
	return srv.DeleteClusterCascade(dao, cluster)
}

func (srv *Service) DeleteOrphanedStatefulSessionByCluster(dao dao.Repository, cluster *domain.Cluster) error {
	sessions, err := dao.FindStatefulSessionConfigsByCluster(cluster)
	if err != nil {
		logger.Errorf("DeleteOrphanedStatefulSessionByCluster failed to load stateful sessions:\n: %v", err)
		return err
	}
	for _, session := range sessions {
		if route, err := dao.FindRouteByStatefulSession(session.Id); err != nil {
			logger.Errorf("DeleteOrphanedStatefulSessionByCluster failed to load route:\n: %v", err)
			return err
		} else if route != nil {
			continue
		}
		if endpoint, err := dao.FindEndpointByStatefulSession(session.Id); err != nil {
			logger.Errorf("DeleteOrphanedStatefulSessionByCluster failed to load endpoint:\n: %v", err)
			return err
		} else if endpoint != nil {
			continue
		}
		clusters, err := dao.FindClustersByFamilyNameAndNamespace(session.ClusterName, msaddr.Namespace{Namespace: session.Namespace})
		if err != nil {
			logger.Errorf("DeleteOrphanedStatefulSessionByCluster failed to load clusters:\n: %v", err)
			return err
		}
		usedByAnotherCluster := false
		for _, existingCluster := range clusters {
			if existingCluster.Id != cluster.Id {
				usedByAnotherCluster = true
				break
			}
		}
		if !usedByAnotherCluster {
			if err := dao.DeleteStatefulSessionConfig(session.Id); err != nil {
				logger.Errorf("DeleteOrphanedStatefulSessionByCluster failed to delete stateful session %+v:\n: %v", session, err)
				return err
			}
		}
	}
	return nil
}
