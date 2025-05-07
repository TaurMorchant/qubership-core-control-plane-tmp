package bluegreen

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/business"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/routekey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
)

type blueGreenPipelineChange struct {
	operation   blueGreenOperation
	versionFrom *domain.DeploymentVersion
	versionTo   *domain.DeploymentVersion
	// here we will store endpoints of current ACTIVE version that should be moved
	endpointsToMove []*domain.Endpoint
	// here we will store MicroserviceVersions of current ACTIVE version that should be moved
	microservicesToMove []*domain.MicroserviceVersion
	// here we will store non-bg cluster keys (will be used later to move routes)
	nonBgClusters map[string]bool
}

func (s *Service) calculateBlueGreenPipelineChange(
	ctx context.Context,
	repo dao.Repository,
	operation blueGreenOperation,
	versionFrom, versionTo *domain.DeploymentVersion) (*blueGreenPipelineChange, error) {

	// load all active version microservices
	microservicesInActive, err := s.versionsRegistry.GetMicroservicesByVersionAsMap(ctx, repo, versionFrom)
	if err != nil {
		return nil, err
	}

	// load all endpoints in active version
	endpointsInActive, err := repo.FindEndpointsByDeploymentVersion(versionFrom.Version)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "MoveMicroservices failed to dao.FindEndpointsByDeploymentVersion %s", versionFrom.Version)
	}

	change := blueGreenPipelineChange{
		operation:           operation,
		versionFrom:         versionFrom,
		versionTo:           versionTo,
		endpointsToMove:     make([]*domain.Endpoint, 0, len(endpointsInActive)),
		microservicesToMove: make([]*domain.MicroserviceVersion, 0, len(microservicesInActive)),
		nonBgClusters:       make(map[string]bool),
	}

	for _, endpointInActive := range endpointsInActive { // collect endpoints to be moved to future active version
		if err := s.addEndpointChangeIfNeeded(ctx, repo, &change, microservicesInActive, endpointInActive); err != nil {
			return nil, err
		}
	}

	// now check MicroserviceVersions in current ACTIVE version that still left in the map
	// (means their had no endpoints, and we did not check them yet)
	for _, microserviceVersion := range microservicesInActive {
		if operation == promote || microserviceVersion.InitialDeploymentVersion != versionFrom.Version {
			isMicroservicePresentInCandidate, err := s.versionsRegistry.IsMicroservicePresentInVersion(ctx, repo, microserviceVersion.GetMicroserviceKey(), versionTo)
			if err != nil {
				return nil, err
			}
			if !isMicroservicePresentInCandidate {
				change.microservicesToMove = append(change.microservicesToMove, microserviceVersion)
			}
		}
	}
	return &change, nil
}

// addEndpointChangeIfNeeded function is used to find out whether provided endpoint belongs to blue-green service:
// does it have any endpoints or MicroserviceVersion entries for future ACTIVE version;
// in case not - endpoint should be "moved" to future ACTIVE version during promote
// and endpoint's cluster should be considered non-blue-green cluster.
// This function also affects microservicesInActive map for optimization: it removes entries for microservices
// that have been already processed by the function, so later we do not double-check these microservices.
//
// Note, we only check future ACTIVE Endpoint (or MicroserviceVersion) existence and do not check routes!
// If you also want to check routes, make sure that 'ext-authz' endpoint is moved to new version
// because 'ext-authz' does not have any routes.
func (s *Service) addEndpointChangeIfNeeded(ctx context.Context, repo dao.Repository, change *blueGreenPipelineChange,
	microservicesInActive map[domain.MicroserviceKey]*domain.MicroserviceVersion, endpointInActive *domain.Endpoint) error {

	endpointsInCandidate, err := repo.FindEndpointsByClusterIdAndDeploymentVersion(endpointInActive.ClusterId, change.versionTo)
	if err != nil {
		return log.ErrorC(ctx, err, "MoveMicroservices failed to dao.FindEndpointsByClusterIdAndDeploymentVersion")
	}

	if len(endpointsInCandidate) > 0 { // service has endpoints in future active version, don't need to move endpoint
		return nil
	}
	// check if microservice candidate version is stored in registry, should not move endpoint in such case
	cluster, err := repo.FindClusterById(endpointInActive.ClusterId)
	if err != nil {
		return log.ErrorC(ctx, err, "MoveMicroservices failed to dao.FindClusterById")
	}
	msKey := cluster.GetMicroserviceKey()
	microservicePresentInCandidate, err := s.versionsRegistry.IsMicroservicePresentInVersion(ctx, repo, msKey, change.versionTo)
	if err != nil {
		return err
	}
	if !microservicePresentInCandidate {
		// need to move endpoint and microservice version (if exists); and now we consider this cluster as non-bg cluster

		// but we don't need to move endpoint if it was created in the version being rolled back right now
		if change.operation == promote || endpointInActive.InitialDeploymentVersion != change.versionFrom.Version {
			change.endpointsToMove = append(change.endpointsToMove, endpointInActive)
			change.nonBgClusters[cluster.Name] = true
		}

		// now same for the MicroserviceVersion
		if microserviceVersion, exists := microservicesInActive[msKey]; exists {
			if change.operation == promote || microserviceVersion.InitialDeploymentVersion != change.versionFrom.Version {
				change.microservicesToMove = append(change.microservicesToMove, microserviceVersion)
				// remove from local active microserviceVersions map, so we don't check this microservice twice
				delete(microservicesInActive, msKey)
			}
		}
	}
	return nil
}

func (c *blueGreenPipelineChange) apply(ctx context.Context, repo dao.Repository) error {
	if err := c.moveEndpointsToNewActiveVersion(ctx, repo); err != nil {
		logger.ErrorC(ctx, "Failed to process endpoints: %v", err)
		return err
	}
	if err := c.moveMicroserviceVersionsToNewActiveVersion(ctx, repo); err != nil {
		logger.ErrorC(ctx, "Failed to process microserviceVersions: %v", err)
		return err
	}
	if err := c.moveStatefulSessionPerClusterConfigs(ctx, repo); err != nil {
		logger.ErrorC(ctx, "Failed to process per-cluster stateful session configs: %v", err)
		return err
	}
	return c.processRoutes(ctx, repo)
}

func (c *blueGreenPipelineChange) moveEndpointsToNewActiveVersion(ctx context.Context, repo dao.Repository) error {
	for _, activeEndpoint := range c.endpointsToMove {
		clonedEndpoint := activeEndpoint.Clone()
		if clonedEndpoint.StatefulSessionId != 0 {
			sessionToMove, err := repo.FindStatefulSessionConfigById(clonedEndpoint.StatefulSessionId)
			if err != nil {
				logger.ErrorC(ctx, "Can't find stateful session to move \n %v", err)
				return err
			}
			if err := c.moveStatefulSession(ctx, repo, sessionToMove); err != nil {
				logger.ErrorC(ctx, "Could not move stateful session with endpoint during %v:\n %v", c.operation, err)
				return err
			}
		}
		clonedEndpoint.DeploymentVersion = c.versionTo.Version
		if err := repo.SaveEndpoint(clonedEndpoint); err != nil {
			logger.ErrorC(ctx, "Can't save endpoint %v \n %v", clonedEndpoint, err)
			return err
		}
		logger.InfoC(ctx, "Moved endpoint to future ACTIVE version: %v", clonedEndpoint)
	}
	return nil
}

func (c *blueGreenPipelineChange) moveMicroserviceVersionsToNewActiveVersion(ctx context.Context, repo dao.Repository) error {
	for _, msVersion := range c.microservicesToMove {
		clonedMsVersion := msVersion.Clone()
		clonedMsVersion.DeploymentVersion = c.versionTo.Version
		if err := repo.SaveMicroserviceVersion(clonedMsVersion); err != nil {
			logger.ErrorC(ctx, "Can't save microservice version %v \n %v", *clonedMsVersion, err)
			return err
		}
		logger.InfoC(ctx, "Moved microservice version %+v to future ACTIVE version %s", *clonedMsVersion, c.versionTo.Version)
	}
	return nil
}

func (c *blueGreenPipelineChange) moveStatefulSession(ctx context.Context, repo dao.Repository, session *domain.StatefulSession) error {
	sessionToSave := session.Clone()
	sessionToSave.DeploymentVersion = c.versionTo.Version
	if err := repo.SaveStatefulSessionConfig(sessionToSave); err != nil {
		logger.ErrorC(ctx, "Failed to save stateful session %+v moved to target version:\n %v", *sessionToSave, err)
		return err
	}
	return nil
}

func (c *blueGreenPipelineChange) moveStatefulSessionPerClusterConfigs(ctx context.Context, repo dao.Repository) error {
	for clusterName := range c.nonBgClusters {
		cluster, err := repo.FindClusterByName(clusterName)
		if err != nil {
			logger.ErrorC(ctx, "Failed to load cluster for moveStatefulSessionPerClusterConfigs by name %s:\n %v", clusterName, err)
			return err
		}
		if cluster != nil {
			if err := c.moveStatefulSessionConfigsForCluster(ctx, repo, cluster); err != nil {
				logger.ErrorC(ctx, "Failed to move per-cluster stateful session configs for %s:\n %v", clusterName, err)
				return err
			}
		}
	}
	return nil
}

func (c *blueGreenPipelineChange) processRoutes(ctx context.Context, dao dao.Repository) error {
	routes, err := dao.FindRoutesByDeploymentVersions(c.versionFrom, c.versionTo)
	if err != nil {
		logger.ErrorC(ctx, "Can't find routes by versions %v and %v \n %v", c.versionFrom, c.versionTo, err)
		return err
	}

	groupedRoutesByVirtualHost := make(map[int32][]*domain.Route)
	for _, route := range routes {
		groupedRoutesByVirtualHost[route.VirtualHostId] = append(groupedRoutesByVirtualHost[route.VirtualHostId], route)
	}

	for _, groupedRoutes := range groupedRoutesByVirtualHost {
		groupedRoutesByRouteMatcher := make(map[string][]*domain.Route)
		for _, route := range groupedRoutes {
			keyMatcher := routekey.GenerateNoVersionKey(*route)
			groupedRoutesByRouteMatcher[keyMatcher] = append(groupedRoutesByRouteMatcher[keyMatcher], route)
		}
		logger.DebugC(ctx, "Grouped routes by matchers: %s", groupedRoutesByRouteMatcher)
		for _, routesByRouteMatcher := range groupedRoutesByRouteMatcher {
			logger.DebugC(ctx, "Processing group routes: \n\t%s", routesByRouteMatcher)
			err := c.processRoutesGroupedByMatcher(ctx, dao, routesByRouteMatcher)
			if err != nil {
				logger.ErrorC(ctx, "Route processing failed for b/g operation %v -> %v \n %v", c.versionFrom, c.versionTo, err)
				return err
			}
		}
	}
	return nil
}

func (c *blueGreenPipelineChange) processRoutesGroupedByMatcher(ctx context.Context, dao dao.Repository, routesByRouteMatcher []*domain.Route) error {
	if len(routesByRouteMatcher) > 1 {
		// route for such matcher presents in both versions, nothing to do
		logger.DebugC(ctx, "Routes exists in active and candidate. Skipping \n\t%s", routesByRouteMatcher)
		return nil
	}
	route := routesByRouteMatcher[0]
	if route.DeploymentVersion == c.versionFrom.Version {
		// route presents only in current ACTIVE version
		logger.DebugC(ctx, "Route exists only in current Active version and needs to be processed: \n\t%s", route)
		nonBgRoute := false
		for clusterName := range c.nonBgClusters {
			if clusterName == route.ClusterName {
				nonBgRoute = true
				break
			}
		}
		if nonBgRoute {
			logger.DebugC(ctx, "Route serves no blue-green cluster %s", route.ClusterName)
			// this is non-blue-green route, so it needs to be moved to a new ACTIVE version

			// but don't move route during rollback if it was originally registered only in future CANDIDATE version
			if c.operation == rollback && route.InitialDeploymentVersion == c.versionFrom.Version {
				return nil
			}

			clonedRoute := route.Clone()
			if clonedRoute.StatefulSessionId != 0 {
				routeStatefulSession, err := dao.FindStatefulSessionConfigById(clonedRoute.StatefulSessionId)
				if err != nil {
					logger.ErrorC(ctx, "Can't load route %+v stateful session:\n %v", *clonedRoute, err)
					return err
				}
				if err := c.moveStatefulSession(ctx, dao, routeStatefulSession); err != nil {
					logger.ErrorC(ctx, "Can't move route %+v stateful session:\n %v", *clonedRoute, err)
					return err
				}
			}
			clonedRoute.DeploymentVersion = c.versionTo.Version
			logger.DebugC(ctx, "Changed deployment version of route to %s", c.versionTo.Version)
			if err := dao.SaveRoute(clonedRoute); err != nil {
				logger.ErrorC(ctx, "Can't move route %v from version %v to version %v:\n %v", *clonedRoute, c.versionFrom.Version, c.versionTo.Version, err)
				return err
			}
		} else {
			logger.DebugC(ctx, "Route belongs to blue-green cluster %s and is missing in future ACTIVE version", route.ClusterName)
			// blue-green cluster route is missing in future ACTIVE version: no need to prohibit anything
		}
	} else {
		logger.DebugC(ctx, "Route exists only in future Active version: \n\t%s", route)
		// route presents only in future ACTIVE version, need to prohibit old route

		notFoundRoute := business.GenerateProhibitRoute(route, msaddr.NewNamespace(""), c.versionFrom.Version)
		notFoundRoute.RouteKey = routekey.GenerateNoVersionKey(*notFoundRoute)
		if err := dao.SaveRoute(notFoundRoute); err != nil {
			logger.ErrorC(ctx, "Can't save prohibited route %v to version %v", *route, c.versionFrom.Version)
			return err
		}
		logger.DebugC(ctx, "Saved prohibit route: \n\t%s", notFoundRoute)
	}
	return nil
}

func (c *blueGreenPipelineChange) moveStatefulSessionConfigsForCluster(ctx context.Context, repo dao.Repository, cluster *domain.Cluster) error {
	logger.InfoC(ctx, "Moving stateful session configurations bound to %s from %s to %s", cluster.Name, c.versionFrom.Version, c.versionTo.Version)
	clusterFamilyName := clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(cluster.Name)
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(cluster.Name)
	originalSessions, err := repo.FindStatefulSessionConfigsByClusterAndVersion(clusterFamilyName, namespace, c.versionFrom)
	if err != nil {
		logger.ErrorC(ctx, "Failed to load stateful session configs of the original version for cluster %s:\n %v", cluster.Name, err)
		return err
	}
	if len(originalSessions) == 0 {
		return nil // nothing to do
	}

	// collect only per-cluster configs
	perClusterSessions := make([]*domain.StatefulSession, 0, len(originalSessions))
	for _, originalSession := range originalSessions {
		if isPerCluster, err := c.isPerClusterStatefulSession(ctx, repo, originalSession); err != nil {
			logger.ErrorC(ctx, "isPerClusterStatefulSession for cluster %s failed:\n %v", cluster.Name, err)
			return err
		} else if isPerCluster {
			if c.operation == promote || originalSession.DeploymentVersion != originalSession.InitialDeploymentVersion {
				// do not roll back to the version before InitialDeploymentVersion
				perClusterSessions = append(perClusterSessions, originalSession)
			}
		}
	}
	if len(perClusterSessions) == 0 {
		return nil // nothing to do
	}

	// check that there is no per-cluster pair in target version
	targetVersionSessions, err := repo.FindStatefulSessionConfigsByClusterAndVersion(clusterFamilyName, namespace, c.versionTo)
	if err != nil {
		logger.ErrorC(ctx, "Failed to load stateful session configs of target version for cluster %s:\n %v", cluster.Name, err)
		return err
	}
	for _, targetVerSession := range targetVersionSessions {
		if isPerCluster, err := c.isPerClusterStatefulSession(ctx, repo, targetVerSession); err != nil {
			logger.ErrorC(ctx, "Failed isPerClusterStatefulSession on stateful session config of target version for cluster %s:\n %v", cluster.Name, err)
			return err
		} else if isPerCluster {
			return nil // there is already a per-cluster configuration in target version, nothing to move
		}
	}
	// no pair in target version found, move the stateful session
	for _, sessionToMove := range perClusterSessions {
		if err := c.moveStatefulSession(ctx, repo, sessionToMove); err != nil {
			logger.ErrorC(ctx, "Failed to move stateful session config to target version for cluster %s:\n %v", cluster.Name, err)
			return err
		}
	}
	return nil
}

func (c *blueGreenPipelineChange) isPerClusterStatefulSession(ctx context.Context, repo dao.Repository, session *domain.StatefulSession) (bool, error) {
	route, err := repo.FindRouteByStatefulSession(session.Id)
	if err != nil {
		logger.ErrorC(ctx, "Failed to load route by stateful session:\n %v", err)
		return false, err
	}
	if route != nil {
		return false, nil
	}
	endpoint, err := repo.FindEndpointByStatefulSession(session.Id)
	if err != nil {
		logger.ErrorC(ctx, "Failed to load endpoint by stateful session:\n %v", err)
		return false, err
	}
	if endpoint != nil {
		return false, nil
	}
	return true, nil
}
