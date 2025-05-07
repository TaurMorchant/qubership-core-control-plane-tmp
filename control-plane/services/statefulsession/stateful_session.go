package statefulsession

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var (
	log = logging.GetLogger("stateful_session")

	ErrNoCluster       = errors.New("StatefulSession: no cluster found for specified cluster name and namespace")
	ErrVersionArchived = errors.New("StatefulSession: cannot modify stateful session configuration for ARCHIVED deployment version")
)

type Service interface {
	ApplyStatefulSession(ctx context.Context, spec *dto.StatefulSession) error
	FindAll(ctx context.Context) ([]*dto.StatefulSession, error)
}

type serviceImpl struct {
	dao           dao.Dao
	entityService *entity.Service
	bus           bus.BusPublisher
}

func NewService(dao dao.Dao, entityService *entity.Service, bus bus.BusPublisher) *serviceImpl {
	return &serviceImpl{dao: dao, entityService: entityService, bus: bus}
}

func (srv *serviceImpl) ApplyStatefulSession(ctx context.Context, spec *dto.StatefulSession) error {
	log.InfoC(ctx, "Applying stateful session config %+v", *spec)
	var err error
	var changes []memdb.Change
	if spec.Port == nil && spec.Hostname == "" {
		changes, err = srv.applyStatefulSessionForCluster(ctx, spec)
	} else {
		changes, err = srv.applyStatefulSessionForEndpoint(ctx, spec)
	}
	if err != nil {
		log.ErrorC(ctx, "Could not apply stateful session configuration, write transaction finished with error:\n %v", err)
		return err
	}
	log.InfoC(ctx, "Storing stateful session to DAO finished successfully")
	return srv.sendChangeEvent(ctx, changes)
}

func (srv *serviceImpl) FindAll(ctx context.Context) ([]*dto.StatefulSession, error) {
	log.DebugC(ctx, "Getting all stateful sessions from DAO")
	sessions, err := srv.dao.FindAllStatefulSessionConfigs()
	if err != nil {
		log.ErrorC(ctx, "Could get all stateful session configurations from DAO:\n %v", err)
		return nil, err
	}
	result := make([]*dto.StatefulSession, 0, len(sessions))
	for _, session := range sessions {
		sessionDto, err := srv.statefulSessionToDto(ctx, session)
		if err != nil {
			return nil, err
		}
		result = append(result, sessionDto)
	}
	return result, nil
}

func (srv *serviceImpl) statefulSessionToDto(ctx context.Context, session *domain.StatefulSession) (*dto.StatefulSession, error) {
	sessionDto := dto.StatefulSession{
		Gateways:  session.Gateways,
		Version:   session.DeploymentVersion,
		Namespace: session.Namespace,
		Cluster:   session.ClusterName,
		Enabled:   &session.Enabled,
	}
	if session.CookieName != "" {
		sessionDto.Cookie = &dto.Cookie{
			Name: session.CookieName,
			Ttl:  session.CookieTtl,
			Path: domain.NewNullString(session.CookiePath),
		}
	}

	// try to resolve endpoint
	endpoint, err := srv.dao.FindEndpointByStatefulSession(session.Id)
	if err != nil {
		log.ErrorC(ctx, "Failed to load stateful session endpoint by id, error:\n %v", err)
		return nil, err
	}
	if endpoint != nil {
		sessionDto.Hostname = endpoint.Address
		portInt := int(endpoint.Port)
		sessionDto.Port = &portInt
		return &sessionDto, nil
	}

	// try to resolve route
	route, err := srv.dao.FindRouteByStatefulSession(session.Id)
	if err != nil {
		log.ErrorC(ctx, "Failed to load stateful session route by id using DAO, error:\n %v", err)
		return nil, err
	}
	if route != nil {
		route, err = srv.entityService.LoadRouteRelations(srv.dao, route)
		if err != nil {
			log.ErrorC(ctx, "Failed to load stateful session route relations using entityService, error:\n %v", err)
			return nil, err
		}
		sessionDto.Route = dto.DefaultResponseConverter.ConvertRouteMatcher(route)
		return &sessionDto, nil
	}

	// stateful session config bound to cluster family name & namespace
	return &sessionDto, nil
}

func (srv *serviceImpl) applyStatefulSessionForCluster(ctx context.Context, spec *dto.StatefulSession) ([]memdb.Change, error) {
	return srv.dao.WithWTx(func(repo dao.Repository) error {
		deploymentVersion, err := srv.resolveDeploymentVersion(ctx, repo, spec.Version)
		if err != nil {
			return err
		}
		spec.Version = deploymentVersion.Version

		existingStatefulSessions, err := repo.FindStatefulSessionConfigsByClusterName(spec.Cluster, msaddr.Namespace{Namespace: spec.Namespace})
		if err != nil {
			log.ErrorC(ctx, "Error finding existing stateful sessions by cluster using DAO:\n %v", err)
			return err
		}

		var existingStatefulSession *domain.StatefulSession
		for _, existing := range existingStatefulSessions {
			if existing.InitialDeploymentVersion == spec.Version {
				existingStatefulSession = existing
				break // initialDeploymentVersion version match has higher priority then deploymentVersion match
			} else if existing.DeploymentVersion == spec.Version {
				existingStatefulSession = existing
			}
		}

		if existingStatefulSession == nil {
			domainSession := &domain.StatefulSession{
				ClusterName:              spec.Cluster,
				Namespace:                spec.Namespace,
				InitialDeploymentVersion: spec.Version,
				DeploymentVersion:        spec.Version,
			}
			return srv.applyStatefulSessionNewSpec(ctx, repo, domainSession, spec)
		} else {
			return srv.applyStatefulSessionNewSpec(ctx, repo, existingStatefulSession, spec)
		}
	})
}

func (srv *serviceImpl) applyStatefulSessionForEndpoint(ctx context.Context, spec *dto.StatefulSession) ([]memdb.Change, error) {
	return srv.dao.WithWTx(func(repo dao.Repository) error {
		deploymentVersion, err := srv.resolveDeploymentVersion(ctx, repo, spec.Version)
		if err != nil {
			return err
		}
		spec.Version = deploymentVersion.Version

		portFromRequest := int32(*spec.Port)
		if spec.Hostname != "" {
			endpoints, err := repo.FindEndpointsByAddressAndPortAndDeploymentVersion(spec.Hostname, portFromRequest, spec.Version)
			if err != nil {
				log.ErrorC(ctx, "Error finding endpoints by hostname, port and version using DAO:\n %v", err)
				return err
			}
			for _, endpoint := range endpoints {
				cluster, err := repo.FindClusterById(endpoint.ClusterId)
				if err != nil {
					log.ErrorC(ctx, "Error finding clusters by family name using DAO:\n %v", err)
					return err
				}
				specToApply := spec.Clone()
				specToApply.Cluster = clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(cluster.Name)
				specToApply.Namespace = clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(cluster.Name).Namespace
				if err := srv.applyEndpointStatefulSessionInternal(ctx, repo, endpoint, specToApply); err != nil {
					return err
				}
			}
		} else {
			clusters, err := repo.FindClustersByFamilyNameAndNamespace(spec.Cluster, msaddr.Namespace{Namespace: spec.Namespace})
			if err != nil {
				log.ErrorC(ctx, "Error finding clusters by family name using DAO:\n %v", err)
				return err
			}
			if len(clusters) == 0 {
				return errorcodes.NewCpError(errorcodes.NotFoundEntityError, ErrNoCluster.Error(), nil)
			}
			for _, cluster := range clusters {
				endpoints, err := repo.FindEndpointsByClusterIdAndDeploymentVersion(cluster.Id, deploymentVersion)
				if err != nil {
					log.ErrorC(ctx, "Error cluster endpoints of version %s using DAO:\n %v", deploymentVersion.Version, err)
					return err
				}
				for _, endpoint := range endpoints {
					if endpoint.Port == portFromRequest {
						specToApply := spec.Clone()
						if err := srv.applyEndpointStatefulSessionInternal(ctx, repo, endpoint, specToApply); err != nil {
							return err
						}
					}
				}
			}
		}
		log.InfoC(ctx, "Stateful session config for endpoint applied to control-plane storage successfully")
		return nil
	})
}

func (srv *serviceImpl) applyEndpointStatefulSessionInternal(ctx context.Context, repo dao.Repository, endpoint *domain.Endpoint, spec *dto.StatefulSession) error {
	existingStatefulSession, err := repo.FindStatefulSessionConfigById(endpoint.StatefulSessionId)
	if err != nil {
		log.ErrorC(ctx, "Error finding existing stateful session by endpoint using DAO:\n %v", err)
		return err
	}
	if existingStatefulSession == nil {
		newSession := &domain.StatefulSession{
			ClusterName:              spec.Cluster,
			Namespace:                spec.Namespace,
			DeploymentVersion:        spec.Version,
			InitialDeploymentVersion: spec.Version,
		}
		if err := srv.applyStatefulSessionNewSpec(ctx, repo, newSession, spec); err != nil {
			return err
		}

		endpoint.StatefulSessionId = newSession.Id
		if err := repo.SaveEndpoint(endpoint); err != nil {
			log.ErrorC(ctx, "Error saving endpoint with new stateful session config using DAO:\n %v", err)
			return err
		}
	} else {
		if err := srv.applyStatefulSessionNewSpec(ctx, repo, existingStatefulSession, spec); err != nil {
			return err
		}
	}
	return nil
}

func (srv *serviceImpl) resolveDeploymentVersion(ctx context.Context, repo dao.Repository, versionFromRequest string) (*domain.DeploymentVersion, error) {
	if versionFromRequest == "" {
		if deploymentVersion, err := srv.entityService.GetActiveDeploymentVersion(repo); err != nil {
			log.ErrorC(ctx, "Error loading ACTIVE deployment version using entityService:\n %v", err)
			return nil, err
		} else {
			return deploymentVersion, nil
		}
	} else {
		deploymentVersion, err := srv.entityService.GetOrCreateDeploymentVersion(repo, versionFromRequest)
		if err != nil {
			log.ErrorC(ctx, "Error loading deployment version  using DAO:\n %v", err)
			return nil, err
		}
		if deploymentVersion.Stage == domain.ArchivedStage {
			return nil, errorcodes.NewCpError(errorcodes.OperationOnArchivedVersionError, ErrVersionArchived.Error(), err)
		}
		return deploymentVersion, nil
	}
}

func (srv *serviceImpl) applyStatefulSessionNewSpec(ctx context.Context, repo dao.Repository, statefulSession *domain.StatefulSession, newSpec *dto.StatefulSession) error {
	gateways, err := srv.applyStatefulSessionNewSpecInternal(ctx, repo, statefulSession, newSpec)
	if err != nil {
		log.ErrorC(ctx, "Stateful session configuration apply failed with error:\n %v", err)
		return err
	}
	return srv.generateEnvoyEntityVersions(ctx, repo, gateways)
}

func (srv *serviceImpl) applyStatefulSessionNewSpecInternal(ctx context.Context, repo dao.Repository, statefulSession *domain.StatefulSession, newSpec *dto.StatefulSession) ([]string, error) {
	var gatewaysToNotify []string
	if newSpec.IsDeleteRequest() {
		if statefulSession.Id == 0 {
			log.DebugC(ctx, "Stateful session config does not exist, no need to delete")
			return nil, nil
		}
		if util.SliceContains(newSpec.Gateways, statefulSession.Gateways...) {
			// delete whole stateful session configuration
			err := repo.DeleteStatefulSessionConfig(statefulSession.Id)
			if err != nil {
				log.ErrorC(ctx, "Error deleting statefulSession using DAO:\n %v", err)
			}
			log.DebugC(ctx, "Stateful session config removed successfully")
			return statefulSession.Gateways, err
		} else {
			// only need to remove some gateways from stateful session configuration
			statefulSession.Gateways = util.SubtractFromSlice(statefulSession.Gateways, newSpec.Gateways...)
			gatewaysToNotify = newSpec.Gateways
		}
	} else {
		statefulSession.Gateways = util.MergeStringSlices(statefulSession.Gateways, newSpec.Gateways)
		gatewaysToNotify = statefulSession.Gateways
		if newSpec.Cookie != nil {
			statefulSession.CookieName = newSpec.Cookie.Name
			if newSpec.Cookie.Path.Valid {
				statefulSession.CookiePath = newSpec.Cookie.Path.String
			} else {
				statefulSession.CookiePath = ""
			}
			statefulSession.CookieTtl = newSpec.Cookie.Ttl
		}
		statefulSession.Enabled = newSpec.IsEnabled()
	}

	err := srv.entityService.PutStatefulSession(repo, statefulSession)
	if err != nil {
		log.ErrorC(ctx, "Error saving stateful session using entityService:\n %v", err)
		return nil, err
	}
	log.DebugC(ctx, "Stateful session saved successfully")
	return gatewaysToNotify, nil
}

func (srv *serviceImpl) generateEnvoyEntityVersions(ctx context.Context, repo dao.Repository, gateways []string) error {
	for _, gateway := range gateways {
		version := domain.NewEnvoyConfigVersion(gateway, domain.RouteConfigurationTable)
		err := repo.SaveEnvoyConfigVersion(version)
		if err != nil {
			log.ErrorC(ctx, "Saving new envoy config version after stateful session update has failed:\n %v", err)
			return err
		}
		log.InfoC(ctx, "Saved new envoyConfigVersion for nodeGroup %s and entity %s: %+v", gateway, domain.RouteConfigurationTable, version)
	}
	return nil
}

func (srv *serviceImpl) sendChangeEvent(ctx context.Context, changes []memdb.Change) error {
	if len(changes) > 0 {
		log.DebugC(ctx, "Successfully saved stateful session configuration; sending change event")
		if err := srv.bus.Publish(bus.TopicMultipleChanges, events.NewMultipleChangeEvent(changes)); err != nil {
			log.ErrorC(ctx, "Could not send stateful session changes to event bus due to error:\n %v", err)
			return err
		}
	}
	log.InfoC(ctx, "Stateful session configuration applied successfully")
	return nil
}

func (srv *serviceImpl) GetStatefulSessionResource() configresources.Resource {
	return &statefulSessionResource{validator: dto.RoutingV3RequestValidator{}, service: srv}
}
