package ratelimit

import (
	"context"
	"errors"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var (
	logger logging.Logger
)

func init() {
	logger = logging.GetLogger("services/ratelimit")
}

type Service struct {
	dao           dao.Dao
	bus           bus.BusPublisher
	entityService *entity.Service
}

func NewService(dao dao.Dao, bus bus.BusPublisher, entityService *entity.Service) *Service {
	return &Service{dao: dao, bus: bus, entityService: entityService}
}

func (s *Service) GetRateLimits(ctx context.Context) ([]dto.RateLimit, error) {
	logger.DebugC(ctx, "Getting all rate limit configurations")
	rateLimits, err := s.dao.FindAllRateLimits()
	if err != nil {
		logger.ErrorC(ctx, "Failed to load all rate limits using DAO:\n %v", err)
		return nil, err
	}
	result := make([]dto.RateLimit, 0, len(rateLimits))
	for _, rateLimit := range rateLimits {
		result = append(result, dto.ConvertRateLimitToDTO(rateLimit))
	}
	return result, nil
}

func (s *Service) SaveRateLimit(ctx context.Context, rateLimit *dto.RateLimit) error {
	logger.InfoC(ctx, "Applying rate limit configuration: %+v", *rateLimit)
	rateLimitToSave := dto.ConvertRateLimitToDomain(rateLimit)
	if rateLimitToSave.LimitRequestsPerSecond <= 0 {
		return s.deleteRateLimit(ctx, rateLimitToSave)
	}
	return s.saveRateLimit(ctx, rateLimitToSave)
}

func (s *Service) DeleteRateLimit(ctx context.Context, rateLimit *dto.RateLimit) error {
	logger.InfoC(ctx, "Applying rate limit configuration: %+v", *rateLimit)
	rateLimitToDelete := dto.ConvertRateLimitToDomain(rateLimit)
	return s.deleteRateLimit(ctx, rateLimitToDelete)
}

func (s *Service) saveRateLimit(ctx context.Context, rateLimit *domain.RateLimit) error {
	logger.InfoC(ctx, "Saving rate limit configuration %+v", *rateLimit)
	return s.applyRateLimit(ctx, rateLimit, func(repo dao.Repository) error {
		if err := repo.SaveRateLimit(rateLimit); err != nil {
			logger.ErrorC(ctx, "Could not save rate limit configuration using DAO:\n %v", err)
			return err
		}
		return s.updateEnvoyConfigVersions(ctx, repo, rateLimit.Name)
	})
}

func (s *Service) deleteRateLimit(ctx context.Context, rateLimit *domain.RateLimit) error {
	logger.InfoC(ctx, "Deleting rate limit configuration %+v", *rateLimit)
	return s.applyRateLimit(ctx, rateLimit, func(repo dao.Repository) error {
		if err := repo.DeleteRateLimitByNameAndPriority(rateLimit.Name, rateLimit.Priority); err != nil {
			logger.ErrorC(ctx, "Could not delete rate limit configuration using DAO:\n %v", err)
			return err
		}
		return s.updateEnvoyConfigVersions(ctx, repo, rateLimit.Name)
	})
}

func (s *Service) applyRateLimit(ctx context.Context, rateLimit *domain.RateLimit, applyFunc func(repo dao.Repository) error) error {
	changes, err := s.dao.WithWTx(applyFunc)
	if err == nil {
		if len(changes) > 1 { // there is at least one route config using this rate limit
			event := events.NewMultipleChangeEvent(changes)
			if err := s.bus.Publish(bus.TopicMultipleChanges, event); err != nil {
				logger.ErrorC(ctx, "Failed to send event on 'multiple-change' topic during rate limit apply:\n %v", err)
				return err
			}
		}
		logger.InfoC(ctx, "Rate limit configuration %+v was applied successfully", *rateLimit)
		return nil
	} else {
		logger.ErrorC(ctx, "WTx failed during rate limit apply:\n %v", err)
		return err
	}
}

func (s *Service) updateEnvoyConfigVersions(ctx context.Context, repo dao.Repository, rateLimitName string) error {
	logger.DebugC(ctx, "Updating node groups with rate limit configuration %s", rateLimitName)
	routes, err := repo.FindRoutesByRateLimit(rateLimitName)
	if err != nil {
		logger.ErrorC(ctx, "Failed to load routes connected to rate limit config %s using DAO:\n %v", rateLimitName, err)
		return err
	}
	vHostsToUpdate := make(map[int32]bool)

	for _, route := range routes {
		vHostsToUpdate[route.VirtualHostId] = true
	}

	allVHosts, err := repo.FindAllVirtualHosts()
	if err != nil {
		logger.ErrorC(ctx, "Failed to load all virtual hosts using DAO:\n %v", err)
		return err
	}
	for _, vHost := range allVHosts {
		if vHost.RateLimitId == rateLimitName {
			vHostsToUpdate[vHost.Id] = true
		}
	}

	for vHostId := range vHostsToUpdate {
		routeConfig, err := s.entityService.FindRouteConfigurationByVirtualHostId(repo, vHostId)
		if err != nil {
			logger.ErrorC(ctx, "Failed to load route config connected to rate limit config %s using DAO:\n %v", rateLimitName, err)
			return err
		}
		if err := repo.SaveEnvoyConfigVersion(domain.NewEnvoyConfigVersion(routeConfig.NodeGroupId, domain.RouteConfigurationTable)); err != nil {
			logger.ErrorC(ctx, "Failed to update route config connected to rate limit config %s using DAO:\n %v", rateLimitName, err)
			return err
		}
	}
	return nil
}

func (s *Service) ValidateRequest(requestBody dto.RateLimit) error {
	if isValid, msg := s.GetRateLimitResource().Validate(requestBody); !isValid {
		return errors.New(msg)
	}
	return nil
}

func (s *Service) GetRateLimitResource() *rateLimitResource {
	return &rateLimitResource{service: s}
}
