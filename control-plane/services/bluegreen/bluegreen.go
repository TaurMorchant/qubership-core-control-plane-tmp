package bluegreen

import (
	"context"
	"errors"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/envoy"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/loadbalance"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/version"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"strings"
)

type Service struct {
	entityService      *entity.Service
	loadBalanceService *loadbalance.LoadBalanceService
	dao                dao.Dao
	eventBus           bus.BusPublisher
	versionsRegistry   VersionsRegistry[dto.ServicesVersionPayload]
}

type blueGreenOperation int

const (
	promote blueGreenOperation = iota
	rollback
)

var (
	logger logging.Logger
	ctx    context.Context

	ErrNoEndpoint                = errors.New("could not find endpoint with the specified host")
	ErrLegacyVersionEmpty        = errors.New("legacy version is empty")
	ErrCurrentActiveVersionEmpty = errors.New("there is no current active version")
)

func init() {
	logger = logging.GetLogger("blue-green-service")
	ctx = context.Background()
}

func NewService(entityService *entity.Service, loadBalanceService *loadbalance.LoadBalanceService, dao dao.Dao, bus bus.BusPublisher, versionsRegistry VersionsRegistry[dto.ServicesVersionPayload]) *Service {
	return &Service{
		entityService:      entityService,
		loadBalanceService: loadBalanceService,
		dao:                dao,
		eventBus:           bus,
		versionsRegistry:   versionsRegistry,
	}
}

// GetMicroserviceVersion resolve actual b/g version for microservice to be set into 'X-Version' header
// in any outgoing requests from this microservice.
//
// Returns empty string in case microservice is located in ACTIVE version
// and does not need to set any 'X-Version' header at all.
func (s *Service) GetMicroserviceVersion(ctx context.Context, microserviceHost string) (string, error) {
	serviceName, initialVersion := s.parseMicroserviceHost(microserviceHost)
	if initialVersion != "" {
		versions, err := s.versionsRegistry.GetMicroserviceCurrentVersion(ctx, s.dao, serviceName, msaddr.CurrentNamespace(), initialVersion)
		if err != nil {
			_ = log.ErrorC(ctx, err, "Could not get microservice current version via versionRegistry (will fall back to legacy function)")
		} else if len(versions) == 1 {
			currentVersion := versions[0]
			if currentVersion.Stage == domain.ActiveStage {
				return "", nil
			} else {
				return currentVersion.Version, nil
			}
		}
	}
	log.InfoC(ctx, "Falling back to getMicroserviceVersionLegacy for %s", microserviceHost)
	return s.getMicroserviceVersionLegacy(ctx, microserviceHost)
}

func (s *Service) parseMicroserviceHost(host string) (string, string) {
	if idx := strings.LastIndex(host, "-"); idx != -1 {
		return host[:idx], host[idx+1:]
	} else {
		return host, ""
	}
}

func (s *Service) getMicroserviceVersionLegacy(ctx context.Context, microserviceHost string) (string, error) {
	// assuming this is a rare operation so we can fetch all endpoints from in-memory storage and avoid extra indexes
	endpoints, err := s.dao.FindAllEndpoints()
	if err != nil {
		logger.ErrorC(ctx, "Failed to load all endpoints from DAO:\n %v", err)
		return "", errors.New("failed to fetch endpoints data from control-plane storage")
	}
	for _, endpoint := range endpoints {
		if endpoint.Address == microserviceHost {
			dVersion, err := s.dao.FindDeploymentVersion(endpoint.DeploymentVersion)
			if err != nil {
				logger.ErrorC(ctx, "Failed to load endpoint deployment version %s from DAO:\n %v", endpoint.DeploymentVersion, err)
				return "", errors.New("failed to fetch b/g version data from control-plane storage")
			}
			logger.InfoC(ctx, "Resolved b/g version %+v for microservice %s", dVersion, microserviceHost)
			if dVersion.Stage == domain.ActiveStage {
				return "", nil
			} else {
				return endpoint.DeploymentVersion, nil
			}
		}
	}
	logger.DebugC(ctx, "Could not find endpoint with host %s", microserviceHost)
	return "", errorcodes.NewCpError(errorcodes.NotFoundEntityError, ErrNoEndpoint.Error(), nil)
}

func (s *Service) DeleteCandidate(ctx context.Context, dVersion *domain.DeploymentVersion) error {
	changes, err := s.dao.WithWTx(func(storage dao.Repository) error {
		err := s.processCandidates(ctx, storage, dVersion)
		if err != nil {
			logger.ErrorC(ctx, "Can't process candidate for deployment version %v \n %v", dVersion, err)
			return err
		}
		logger.InfoC(ctx, "Deleting deployment version %v", dVersion)
		return storage.DeleteDeploymentVersion(dVersion)
	})
	if err != nil {
		logger.ErrorC(ctx, "Can't delete candidate with version %v \n %v", dVersion, err)
		return err
	}

	logger.InfoC(ctx, "Sending event to update envoy-configuration")
	err = s.eventBus.Publish(bus.TopicReload, events.NewReloadEvent(changes))
	if err != nil {
		logger.ErrorC(ctx, "Can't publish reload event, cause: %v", err)
		return err
	}
	return nil
}

func (s *Service) Promote(ctx context.Context, originPromoteVersion *domain.DeploymentVersion, historySize int) ([]*domain.DeploymentVersion, error) {
	logger.InfoC(ctx, "Starting Promote version '%s' with archive size '%d'", originPromoteVersion.Version, historySize)
	changes, err := s.dao.WithWTx(func(dao dao.Repository) error {
		promoteVersion := originPromoteVersion.Clone()
		nodeGroups, err := s.dao.FindAllNodeGroups()
		if err != nil {
			logger.ErrorC(ctx, "Can't get node groups %v", err)
			return err
		}

		deploymentVersions, err := s.dao.FindAllDeploymentVersions()
		if err != nil {
			logger.ErrorC(ctx, "Can't find deployment %v", err)
			return err
		}
		versionState := version.NewVersionState(deploymentVersions)
		versionState.SetVersionToPromote(promoteVersion)

		// Change stage of current ACTIVE, LEGACY
		currentActiveVersion := versionState.GetActive()
		if currentActiveVersion == nil {
			logger.ErrorC(ctx, "Can't get active deployment version. %v", ErrCurrentActiveVersionEmpty)
			return errorcodes.NewCpError(errorcodes.BlueGreenConflictError, ErrCurrentActiveVersionEmpty.Error(), nil)
		}

		optionalLegacy := versionState.GetLegacy()
		currentActiveVersion.Stage = domain.LegacyStage

		if optionalLegacy != nil {
			optionalLegacy.Stage = domain.ArchivedStage
		}

		// Change stage of promoting version
		promoteVersion.Stage = domain.ActiveStage

		//STORE candidates
		versionsToDelete := versionState.GetCandidates()
		if versionState.GetHistorySize() > historySize && versionState.GetOldestArchivedVersion() != nil {
			versionsToDelete = append(versionsToDelete, versionState.GetArchivedVersionsToDelete(versionState.GetHistorySize()-historySize)...)
		}

		logger.DebugC(ctx, "Deleting autogenerated routes for %s version...", promoteVersion.Version)
		err = s.entityService.DeleteRoutesByAutoGeneratedAndDeploymentVersion(dao, true, promoteVersion.Version)
		if err != nil {
			logger.ErrorC(ctx, "Can't delete routes by deployment version %s \n %v", promoteVersion.Version, err)
			return err
		}

		err = s.deleteCandidates(dao, versionsToDelete...)
		if err != nil {
			logger.ErrorC(ctx, "Can't delete clusters by deployment version %v", err)
			return err
		}
		if err := dao.DeleteDeploymentVersions(versionsToDelete); err != nil {
			logger.ErrorC(ctx, "Can't delete deployment versions %v", err)
			return err
		}

		if err := s.processEndpointsAndRoutes(ctx, dao, currentActiveVersion, promoteVersion, promote); err != nil {
			logger.ErrorC(ctx, "Can't process routes and endpoints: %v", err)
		}

		err = s.loadBalanceService.ApplyLoadBalanceForAllClusters(ctx, dao)
		if err != nil {
			logger.ErrorC(ctx, "Can't configure load balancing %v", err)
			return err
		}
		for _, nodeGroup := range nodeGroups {
			err = s.generateAndSaveEnvoyConfig(dao, nodeGroup.Name)
			if err != nil {
				logger.ErrorC(ctx, "Can't generate and save envoy configuration for node group %s", nodeGroup.Name)
				return err
			}
		}
		for _, dVersion := range []*domain.DeploymentVersion{currentActiveVersion, promoteVersion, optionalLegacy} {
			if dVersion != nil {
				err := s.entityService.SaveDeploymentVersion(dao, dVersion)
				if err != nil {
					logger.ErrorC(ctx, "Can't save deployment version %v \n %v", dVersion, err)
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.InfoC(ctx, "Sending event to update envoy-configuration")
	err = s.eventBus.Publish(bus.TopicReload, events.NewReloadEvent(changes))
	if err != nil {
		logger.ErrorC(ctx, "Can't publish reload event, cause: %v", err)
	}

	logger.InfoC(ctx, "Promoting of version %v has finished.", originPromoteVersion)

	return s.dao.FindAllDeploymentVersions()
}

func (s *Service) Rollback(ctx context.Context) ([]*domain.DeploymentVersion, error) {
	changes, err := s.dao.WithWTx(func(dao dao.Repository) error {
		nodeGroups, err := s.dao.FindAllNodeGroups()
		if err != nil {
			logger.ErrorC(ctx, "Can't get node groups %v", err)
			return err
		}
		deploymentVersions, err := s.dao.FindAllDeploymentVersions()
		if err != nil {
			logger.ErrorC(ctx, "Can't find deployment %v", err)
			return err
		}
		versionState := version.NewVersionState(deploymentVersions)

		//Validate
		futureCandidateVersion := versionState.GetActive()
		futureActiveVersion := versionState.GetLegacy()

		if futureActiveVersion == nil {
			logger.ErrorC(ctx, "rollback is not possible: %v", err)
			return errorcodes.NewCpError(errorcodes.BlueGreenConflictError, ErrLegacyVersionEmpty.Error(), nil)
		}

		versionsToDelete := versionState.GetCandidates()

		//delete candidates
		logger.InfoC(ctx, "Starting rollback")
		err = s.processCandidates(ctx, dao, versionsToDelete...)
		if err != nil {
			logger.ErrorC(ctx, "Can't process candidate deployment version %v \n %v", versionsToDelete, err)
			logger.ErrorC(ctx, "Can't rollback %v", err)
			return err
		}
		if err != nil {
			logger.ErrorC(ctx, "Can't delete candidates %v \n %v", versionsToDelete, err)
			return err
		}

		err = dao.DeleteDeploymentVersions(versionsToDelete)
		if err != nil {
			logger.ErrorC(ctx, "Can't rollback %v", err)
			return err
		}

		futureActiveVersion.Stage = domain.ActiveStage
		futureCandidateVersion.Stage = domain.CandidateStage

		err = s.entityService.DeleteRoutesByAutoGeneratedAndDeploymentVersion(dao, true, futureActiveVersion.Version)
		if err != nil {
			logger.ErrorC(ctx, "Can't delete routes by deployment version %s \n %v", futureActiveVersion.Version, err)
			return err
		}

		if err := s.processEndpointsAndRoutes(ctx, dao, futureCandidateVersion, futureActiveVersion, rollback); err != nil {
			logger.ErrorC(ctx, "Can't process routes and endpoints: %v", err)
			return err
		}

		err = s.loadBalanceService.ApplyLoadBalanceForAllClusters(ctx, dao)
		if err != nil {
			logger.ErrorC(ctx, "Can't configure load balancing \n %v", err)
			return err
		}

		for _, dVersion := range []*domain.DeploymentVersion{futureActiveVersion, futureCandidateVersion} {
			if dVersion != nil {
				err := s.entityService.SaveDeploymentVersion(dao, dVersion)
				if err != nil {
					logger.ErrorC(ctx, "Can't save deployment version %v \n %v", dVersion, err)
					return err
				}
			}
		}

		for _, nodeGroup := range nodeGroups {
			err = s.generateAndSaveEnvoyConfig(dao, nodeGroup.Name)
			if err != nil {
				logger.ErrorC(ctx, "Can't generate and save envoy configuration for node group %s", nodeGroup.Name)
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.InfoC(ctx, "Sending event to update envoy-configuration")
	log.DebugC(ctx, "Publish changes: %+v", changes)
	err = s.eventBus.Publish(bus.TopicReload, events.NewReloadEvent(changes))
	if err != nil {
		logger.ErrorC(ctx, "Can't publish reload event, cause: %v", err)
	}

	logger.InfoC(ctx, "Rollback has finished.")

	return s.dao.FindAllDeploymentVersions()
}

func (s *Service) deleteRoutes(storage dao.Repository, dVersions []*domain.DeploymentVersion) error {
	routes, err := storage.FindRoutesByDeploymentVersions(dVersions...)
	if err != nil {
		logger.ErrorC(ctx, "Can't get routes by cluster name and deployment version %v", err)
		return err
	}
	err = s.entityService.DeleteRoutesByUUID(storage, routes)
	if err != nil {
		logger.ErrorC(ctx, "Can't delete routes by UUID %v", err)
		return err
	}
	return nil
}

func (s *Service) deleteCandidates(storage dao.Repository, dVersions ...*domain.DeploymentVersion) error {
	//find clusters with endpoints
	endpoints, err := storage.FindEndpointsByDeploymentVersionsIn(dVersions)
	if err != nil {
		logger.ErrorC(ctx, "Can't find endpoints by deployment version %v", err)
		return err
	}
	clusterIdsMap := make(map[int32]bool)
	for _, endpoint := range endpoints {
		logger.InfoC(ctx, "Processing endpoint %v", endpoint)
		err = s.entityService.DeleteEndpointCascade(storage, endpoint)
		if err != nil {
			logger.ErrorC(ctx, "Can't delete endpoint:\n %v", err)
			return err
		}
		clusterIdsMap[endpoint.ClusterId] = true
	}

	err = s.deleteRoutes(storage, dVersions)
	if err != nil {
		logger.ErrorC(ctx, "Can't delete routes for versions %s \n %v", dVersions, err)
		return err
	}

	for clusterId, _ := range clusterIdsMap {
		clusterEndpoints, err := storage.FindEndpointsByClusterId(clusterId)
		if err != nil {
			logger.ErrorC(ctx, "Can't find endpoints for cluster id %d \n %v", clusterId, err)
			return err
		}
		if len(clusterEndpoints) > 0 {
			continue
		}
		cluster, err := storage.FindClusterById(clusterId)
		if err != nil {
			logger.ErrorC(ctx, "Can't load cluster by id %d \n  %v", clusterId, err)
			return err
		}
		if err := s.entityService.DeleteClusterCascade(storage, cluster); err != nil {
			logger.ErrorC(ctx, "Can't delete cluster \n  %v", err)
			return err
		}
	}
	statefulSessions, err := storage.FindAllStatefulSessionConfigs()
	if err != nil {
		logger.ErrorC(ctx, "Can't load all stateful sessions from DAO during candidate deletion \n  %v", err)
		return err
	}
	for _, session := range statefulSessions {
		for _, dVersion := range dVersions {
			if dVersion.Version == session.DeploymentVersion {
				logger.InfoC(ctx, "Deleting %+v", *session)
				if err := storage.DeleteStatefulSessionConfig(session.Id); err != nil {
					logger.ErrorC(ctx, "Can't delete statefulSession \n  %v", err)
					return err
				}
				break
			}
		}
	}

	return s.versionsRegistry.DeleteVersions(ctx, storage, dVersions...)
}

func (s *Service) generateAndSaveEnvoyConfig(dao dao.Repository, nodeGroupName string) error {
	envoyConfigService := envoy.NewEnvoyConfigService(dao)
	if err := envoyConfigService.GenerateAndSave(nodeGroupName, domain.ListenerTable); err != nil {
		logger.ErrorC(ctx, "Can't save envoy configuration version for %s and node group %s: \n %v", domain.ListenerTable, nodeGroupName, err)
		return err
	}
	if err := envoyConfigService.GenerateAndSave(nodeGroupName, domain.ClusterTable); err != nil {
		logger.ErrorC(ctx, "Can't save envoy configuration version for %s and node group %s: \n %v", domain.ClusterTable, nodeGroupName, err)
		return err
	}
	if err := envoyConfigService.GenerateAndSave(nodeGroupName, domain.RouteConfigurationTable); err != nil {
		logger.ErrorC(ctx, "Can't save envoy configuration version for %s and node group %s: \n %v", domain.RouteConfigurationTable, nodeGroupName, err)
		return err
	}
	return nil
}

func (s *Service) processEndpointsAndRoutes(ctx context.Context, repo dao.Repository, versionFrom, versionTo *domain.DeploymentVersion, operation blueGreenOperation) error {
	change, err := s.calculateBlueGreenPipelineChange(ctx, repo, operation, versionFrom, versionTo)
	if err != nil {
		return err
	}
	return change.apply(ctx, repo)
}

func (s *Service) processCandidates(ctx context.Context, dao dao.Repository, dVersions ...*domain.DeploymentVersion) error {
	nodeGroups, err := dao.FindAllNodeGroups()
	if err != nil {
		logger.ErrorC(ctx, "Can't get node groups %v", err)
		return err
	}
	logger.InfoC(ctx, "Processing candidate for deployment versions %v", dVersions)
	err = s.deleteCandidates(dao, dVersions...)
	if err != nil {
		logger.ErrorC(ctx, "Can't delete clusters by deployment version %v", err)
		return err
	}

	err = s.loadBalanceService.ApplyLoadBalanceForAllClusters(ctx, dao)
	if err != nil {
		logger.ErrorC(ctx, "Can't configure load balancing %v", err)
		return err
	}

	for _, nodeGroup := range nodeGroups {
		err = s.generateAndSaveEnvoyConfig(dao, nodeGroup.Name)
		if err != nil {
			logger.ErrorC(ctx, "Can't generate and save envoy configuration for node group %s", nodeGroup.Name)
			return err
		}
	}
	return nil
}
