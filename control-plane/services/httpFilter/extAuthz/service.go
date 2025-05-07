package extAuthz

import (
	"context"
	"errors"
	"fmt"
	goerrors "github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/route"
	"github.com/netcracker/qubership-core-control-plane/services/route/registration"
	"github.com/netcracker/qubership-core-control-plane/util"
	"strings"
)

var (
	log          = util.NewLoggerWrap("extAuthz")
	ErrNameTaken = errors.New("extAuthz: extAuthz filter with such name is used by another gateway")
)

type Service interface {
	Apply(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) error
	ValidateApply(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) (bool, string)
	ValidateDelete(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) (bool, string)
	Delete(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) error
	Get(ctx context.Context, gateway string) (*dto.ExtAuthz, error)
}

type service struct {
	dao                dao.Dao
	bus                bus.BusPublisher
	entityService      entity.ServiceInterface
	routeSrv           route.ClusterRegistrationService
	v3RequestProcessor registration.V3RequestProcessor
}

func NewService(dao dao.Dao, bus bus.BusPublisher, entityService entity.ServiceInterface, routeSrv route.ClusterRegistrationService, v3RequestProcessor registration.V3RequestProcessor) *service {
	return &service{
		dao:                dao,
		bus:                bus,
		entityService:      entityService,
		routeSrv:           routeSrv,
		v3RequestProcessor: v3RequestProcessor,
	}
}

func (s *service) ValidateApply(_ context.Context, extAuthz dto.ExtAuthz, gateways ...string) (bool, string) {
	if isValid, errMsg := s.validateGatewaysAndFilterName(extAuthz, gateways...); !isValid {
		return isValid, errMsg
	}
	if extAuthz.Destination.Cluster == "" {
		return false, "field 'destination.cluster' in 'extAuthz' filter spec must not be empty"
	}
	if extAuthz.Destination.Endpoint == "" && extAuthz.Destination.TlsEndpoint == "" {
		return false, "at least one of the fields 'destination.endpoint' and 'destination.tlsEndpoint' in 'extAuthz' filter spec must not be empty"
	}
	return true, ""
}

func (s *service) ValidateDelete(_ context.Context, extAuthz dto.ExtAuthz, gateways ...string) (bool, string) {
	return s.validateGatewaysAndFilterName(extAuthz, gateways...)
}

func (s *service) validateGatewaysAndFilterName(extAuthz dto.ExtAuthz, gateways ...string) (bool, string) {
	if len(gateways) == 0 {
		return false, "field 'gateways' must not be empty"
	}
	if len(gateways) > 1 {
		return false, "field 'gateways' must contain only one gateway in case 'extAuthz' is not null, since different gateways cannot have 'extAuthz' filters with the same names"
	}
	for _, gateway := range gateways {
		if strings.TrimSpace(gateway) == "" {
			return false, "gateway name must not be empty string"
		}
		if domain.IsOobGateway(gateway) {
			return false, "it is forbidden to configure extAuthz filter for " + gateway
		}
	}
	if extAuthz.Name == "" {
		return false, "field 'name' in 'extAuthz' filter spec must not be empty"
	}
	return true, ""
}

func (s *service) Apply(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) error {
	log.InfoC(ctx, "Request to apply ExtAuthz filter %+v for %+q", extAuthz, gateways)
	changes, err := s.dao.WithWTx(func(repo dao.Repository) error {
		for _, gateway := range gateways {
			if err := s.applyForGateway(ctx, repo, extAuthz, gateway); err != nil {
				return log.ErrorC(ctx, err, "Could not apply extAuthz filter for gateway %s", gateway)
			}
		}
		return nil
	})
	if err != nil {
		if goErr, ok := err.(*goerrors.Error); ok {
			if cause := goErr.Unwrap(); cause == ErrNameTaken {
				return cause
			}
		}
		return log.ErrorC(ctx, err, "Transaction failed during extAuthz filter apply")
	}
	if len(changes) > 0 {
		for _, gateway := range gateways {
			if err = s.bus.Publish(bus.TopicChanges, events.NewChangeEventByNodeGroup(gateway, changes)); err != nil {
				return log.ErrorC(ctx, err, "Could not publish changes to event bus on topic %s", bus.TopicChanges)
			}
		}
	}
	log.InfoC(ctx, "ExtAuthz filter %+v for %+q applied successfully", extAuthz, gateways)
	return nil
}

func (s *service) applyForGateway(ctx context.Context, repo dao.Repository, extAuthz dto.ExtAuthz, gateway string) error {
	existingFilter, err := repo.FindExtAuthzFilterByName(extAuthz.Name)
	if err != nil {
		return log.ErrorC(ctx, err, "Could not get existing extAuthz filter by name %s using DAO", extAuthz.Name)
	}
	if existingFilter != nil && existingFilter.NodeGroup != gateway {
		return ErrNameTaken
	}

	activeVersion, err := s.entityService.GetActiveDeploymentVersion(repo)
	if err != nil {
		return log.ErrorC(ctx, err, "Could not get ACTIVE version using DAO")
	}

	nodeGroup := domain.NodeGroup{Name: gateway}
	if _, err := s.entityService.CreateOrUpdateNodeGroup(repo, nodeGroup); err != nil {
		return log.ErrorC(ctx, err, "Failed to load or save nodeGroup %v to in-memory storage", nodeGroup)
	}

	cluster := s.v3RequestProcessor.ProcessDestination(extAuthz.Destination, "", activeVersion.Version, false)
	tlsConfigName := ""
	if cluster.TLS != nil {
		tlsConfigName = cluster.TLS.Name
	}
	if err = s.routeSrv.SaveCluster(ctx, repo, *cluster, tlsConfigName, gateway); err != nil {
		return log.ErrorC(ctx, err, "Could not save cluster %s using route v3 registration service", cluster.Name)
	}

	extAuthzFilter := newExtAuthzFilter(gateway, cluster.Name, extAuthz)
	if err = repo.SaveExtAuthzFilter(extAuthzFilter); err != nil {
		return log.ErrorC(ctx, err, "Could not save extAuthz filter %+v using DAO", *extAuthzFilter)
	}

	if err = s.entityService.GenerateEnvoyEntityVersions(ctx, repo, gateway, domain.RouteConfigurationTable, domain.ListenerTable, domain.ClusterTable); err != nil {
		return log.ErrorC(ctx, err, "Could not generate new envoy entities versions after extAuthz filter apply for %s", gateway)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, extAuthz dto.ExtAuthz, gateways ...string) error {
	log.InfoC(ctx, "Request to apply ExtAuthz filter %+v for %+q", extAuthz, gateways)
	changes, err := s.dao.WithWTx(func(repo dao.Repository) error {
		for _, gateway := range gateways {
			if err := s.deleteForGateway(ctx, repo, extAuthz, gateway); err != nil {
				return log.ErrorC(ctx, err, "Could not delete extAuthz filter for gateway %s", gateway)
			}
		}
		return nil
	})
	if err != nil {
		if goErr, ok := err.(*goerrors.Error); ok {
			if cause := goErr.Unwrap(); cause == ErrNameTaken {
				return cause
			}
		}
		return log.ErrorC(ctx, err, "Transaction failed during extAuthz filter deletion")
	}
	if len(changes) > 0 {
		for _, gateway := range gateways {
			if err = s.bus.Publish(bus.TopicChanges, events.NewChangeEventByNodeGroup(gateway, changes)); err != nil {
				return log.ErrorC(ctx, err, "Could not publish changes to event bus on topic %s", bus.TopicChanges)
			}
		}
	}
	log.InfoC(ctx, "ExtAuthz filter %s for %+q deleted successfully", extAuthz.Name, gateways)
	return nil
}

func (s *service) deleteForGateway(ctx context.Context, repo dao.Repository, extAuthz dto.ExtAuthz, gateway string) error {
	existingFilter, err := repo.FindExtAuthzFilterByName(extAuthz.Name)
	if err != nil {
		return log.ErrorC(ctx, err, "Could not get existing extAuthz filter by name %s using DAO", extAuthz.Name)
	}
	if existingFilter == nil {
		log.InfoC(ctx, "No extAuthz filter to delete for name %s", extAuthz.Name)
		return nil
	}
	if existingFilter.NodeGroup != gateway {
		return ErrNameTaken
	}

	if err = repo.DeleteExtAuthzFilter(extAuthz.Name); err != nil {
		return log.ErrorC(ctx, err, "Could not delete extAuthz filter %s using DAO", extAuthz.Name)
	}

	if err = s.entityService.GenerateEnvoyEntityVersions(ctx, repo, gateway, domain.RouteConfigurationTable, domain.ListenerTable); err != nil {
		return log.ErrorC(ctx, err, "Could not generate new envoy entities versions after extAuthz filter apply for %s", gateway)
	}
	return nil
}

func (s *service) Get(ctx context.Context, gateway string) (*dto.ExtAuthz, error) {
	extAuthz, err := s.dao.WithRTxVal(func(repo dao.Repository) (interface{}, error) {
		extAuthz, err := repo.FindExtAuthzFilterByNodeGroup(gateway)
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Could not load extAuthz filter by node group %s from DAO", gateway)
		}
		if extAuthz == nil {
			return nil, nil
		}
		extAuthzDto := dto.ExtAuthz{
			Name:              extAuthz.Name,
			Destination:       dto.RouteDestination{},
			ContextExtensions: extAuthz.ContextExtensions,
		}
		if extAuthz.Timeout != 0 {
			extAuthzDto.Timeout = util.WrapValue(extAuthz.Timeout)
		}
		extAuthzDto.Destination.Cluster = clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(extAuthz.ClusterName)

		activeVersion, err := s.entityService.GetActiveDeploymentVersion(repo)
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Could not load active version from storage while getting ExtAuthz filter for %s", gateway)
		}
		if activeVersion == nil {
			return nil, log.ErrorC(ctx, errors.New("extAuthz: active version is missing in control-plane storage"), "Error while getting extAuthz filter for gateway %s", gateway)
		}

		extAuthzDto.Destination.Cluster = clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(extAuthz.ClusterName)

		endpoints, err := repo.FindEndpointsByClusterName(extAuthz.ClusterName)
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Error while getting endpoints related to extAuthz filter for gateway %s", gateway)
		}
		endpointAddr := ""
		for _, endpoint := range endpoints {
			if endpoint.DeploymentVersion == activeVersion.Version {
				endpointAddr = fmt.Sprintf("%s://%s:%d", endpoint.Protocol, endpoint.Address, endpoint.Port)
				break
			}
		}
		extAuthzDto.Destination.Endpoint = endpointAddr
		return extAuthzDto, nil
	})
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Read transaction failed while getting extAuthz filter for gateway %s", gateway)
	}

	if extAuthz != nil {
		return extAuthz.(*dto.ExtAuthz), nil
	}
	return nil, nil
}

func newExtAuthzFilter(gateway, clusterKey string, request dto.ExtAuthz) *domain.ExtAuthzFilter {
	extAuthzFilter := &domain.ExtAuthzFilter{
		Name:              request.Name,
		ClusterName:       clusterKey,
		ContextExtensions: request.ContextExtensions,
		NodeGroup:         gateway,
	}
	if request.Timeout != nil {
		extAuthzFilter.Timeout = *request.Timeout
	}
	return extAuthzFilter
}
