package loadbalance

import (
	"context"
	"errors"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/active"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/envoy"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/registration"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var log logging.Logger

func init() {
	log = logging.GetLogger("load-balance")
}

type LoadBalanceService struct {
	dao           dao.Dao
	entityService *entity.Service
	eventBus      *bus.EventBusAggregator
}

func NewLoadBalanceService(dao dao.Dao, entityService *entity.Service, eventBus *bus.EventBusAggregator) *LoadBalanceService {
	return &LoadBalanceService{dao: dao, entityService: entityService, eventBus: eventBus}
}

func (c *LoadBalanceService) ApplyLoadBalanceConfig(ctx context.Context, clusterName string, version string, policies []*domain.HashPolicy) error {
	activeVersion, err := c.entityService.GetActiveDeploymentVersion(c.dao)
	if err != nil {
		log.ErrorC(ctx, "Failed to get active deployment version: %v", err)
		return err
	}
	_, deploymentVersion, err := registration.ResolveVersions(c.dao, clusterName, version, activeVersion.Version)
	if err != nil {
		return err
	}
	nodeGroups, changes, err := c.dao.WithWTxVal(func(dao dao.Repository) (interface{}, error) {
		return c.applyLoadBalanceConfigInternal(ctx, clusterName, deploymentVersion, policies, dao)
	})
	if err != nil {
		log.ErrorC(ctx, "Error apply load balancing rules: %s", err.Error())
		return err
	}
	log.InfoC(ctx, "Load balancing policies successfully applied: cluster=%+v, policies=%+v", clusterName, policies)
	return c.publishChanges(ctx, nodeGroups, changes)
}

func (c *LoadBalanceService) ConfigureLoadBalanceForAllClusters(ctx context.Context) error {
	nodeGroups, changes, err := c.dao.WithWTxVal(func(dao dao.Repository) (interface{}, error) {
		err := c.ApplyLoadBalanceForAllClusters(ctx, dao)
		if err != nil {
			log.ErrorC(ctx, "Can't apply load balance policy to all clusters %v", err)
			return nil, err
		}
		return dao.FindAllNodeGroups()
	})
	if err != nil {
		log.ErrorC(ctx, "Error apply load balancing: %s", err.Error())
		return err
	}

	return c.publishChanges(ctx, nodeGroups, changes)
}

func (c *LoadBalanceService) ApplyLoadBalanceForAllClusters(ctx context.Context, dao dao.Repository) error {
	clusters, err := c.entityService.GetClustersWithRelations(dao)
	if err != nil {
		log.ErrorC(ctx, "Can't find clusters %v", err)
		return err
	}
	for _, cluster := range clusters {
		isActiveActiveCluster := active.IsActiveActiveCluster(cluster.Name)
		if isActiveActiveCluster {
			// skip active-active clusters
			continue
		}
		err = c.configureLoadBalanceForCluster(ctx, dao, cluster)
		if err != nil {
			log.ErrorC(ctx, "Can't configure load balancing for cluster %v \n %v", cluster, err)
			return err
		}
	}
	log.InfoC(ctx, "Load balancing policies successfully applied: clusters=%v", clusters)
	return nil
}

func (c *LoadBalanceService) applyLoadBalanceConfigInternal(ctx context.Context, clusterName, version string, policies []*domain.HashPolicy, dao dao.Repository) (interface{}, error) {
	if cluster, err := c.entityService.GetClusterWithRelations(dao, clusterName); err != nil {
		return nil, err
	} else if cluster == nil {
		log.ErrorC(ctx, "Can't find cluster by given name: %s", clusterName)
		return nil, errors.New("Cluster with name " + clusterName + " is not present")
	} else {
		return c.processClusterWithHashPolicies(ctx, dao, cluster, version, policies)
	}
}

func (c *LoadBalanceService) configureLoadBalanceForCluster(ctx context.Context, repository dao.Repository, cluster *domain.Cluster) error {
	needBalancing := false
	for _, endpoint := range cluster.Endpoints {
		dVStage := endpoint.DeploymentVersionVal.Stage
		if len(endpoint.HashPolicies) > 0 && (dVStage == domain.ActiveStage || dVStage == domain.CandidateStage) {
			needBalancing = true
			break
		}
	}
	return c.applyLoadBalanceForCluster(ctx, needBalancing, repository, cluster)
}

func (c *LoadBalanceService) processClusterWithHashPolicies(ctx context.Context, dao dao.Repository, cluster *domain.Cluster, version string, policies []*domain.HashPolicy) (interface{}, error) {
	log.DebugC(ctx, "Apply policies to cluster=%+v, policies=%+v", cluster.Name, policies)
	needBalancing := len(policies) > 0
	for _, endpoint := range cluster.Endpoints {
		// use shallow equality
		if version != endpoint.DeploymentVersion {
			// we will not affect this endpoint, but depending on it's version and hash policy we will set cluster lbPolicy
			if !needBalancing && len(endpoint.HashPolicies) > 0 &&
				(endpoint.DeploymentVersionVal.Stage == domain.ActiveStage || endpoint.DeploymentVersionVal.Stage == domain.CandidateStage) {
				needBalancing = true
			}
			continue
		}
		// do it as simple as possible: clean old and add new policies
		for _, policy := range endpoint.HashPolicies {
			if err := dao.DeleteHashPolicyById(policy.Id); err != nil {
				return nil, err
			}
		}

		for _, policy := range policies {
			policy.EndpointId = endpoint.Id
			if err := dao.SaveHashPolicy(policy); err != nil {
				return nil, err
			}
		}
	}
	return cluster.NodeGroups, c.applyLoadBalanceForCluster(ctx, needBalancing, dao, cluster)
}

func (c *LoadBalanceService) applyLoadBalanceForCluster(ctx context.Context, needBalancing bool, repository dao.Repository, cluster *domain.Cluster) error {
	if needBalancing {
		cluster.LbPolicy = domain.LbPolicyRingHash
	} else {
		cluster.LbPolicy = domain.LbPolicyLeastRequest
	}
	if err := repository.SaveCluster(cluster); err != nil {
		log.ErrorC(ctx, "Error during saving cluster with updated load balancing config: %v", err)
		return err
	}
	for _, nodeGroup := range cluster.NodeGroups {
		log.InfoC(ctx, "Saving new EnvoyConfigVersion for NodeGroup '%v' after applying load balancing config", nodeGroup.Name)
		if err := envoy.UpdateAllResourceVersions(repository, nodeGroup.Name); err != nil {
			log.ErrorC(ctx, "Error while saving new envoy config version for node group %v after applying load balancing config: %v", nodeGroup.Name, err)
			return err
		}
	}
	return nil
}

func (c *LoadBalanceService) publishChanges(ctx context.Context, nodeGroup interface{}, changes []memdb.Change) error {
	log.InfoC(ctx, "Load balancing policies successfully applied for all clusters")

	for _, n := range nodeGroup.([]*domain.NodeGroup) {
		event := events.NewChangeEventByNodeGroup(n.Name, changes)
		if err := c.eventBus.Publish(bus.TopicChanges, event); err != nil {
			log.ErrorC(ctx, "Can't publish event to eventBus: topic=%s, event=%v, error: %s", bus.TopicChanges, event, err.Error())
			return err
		}
	}
	return nil
}
