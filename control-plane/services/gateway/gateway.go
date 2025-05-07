package gateway

import (
	"context"
	"errors"
	goerrors "github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/util"
)

var (
	log = util.NewLoggerWrap("services/gateway")

	inputErrors = []error{ErrMultipleVirtualHosts, ErrInvalidVirtualHost, ErrInvalidEndpoint, ErrHasTlsConfig, ErrHasVirtualHost, ErrHasCluster}

	ErrMultipleVirtualHosts = errors.New("gateway: this gateway already has more than one virtual service")
	ErrInvalidVirtualHost   = errors.New("gateway: this gateway virtual service does not have host starting with '*'")
	ErrInvalidEndpoint      = errors.New("gateway: ingress gateway can only have endpoints leading to public or private gateway")
	ErrHasTlsConfig         = errors.New("gateway: cannot delete gateway declaration since it is connected to some tls config")
	ErrHasVirtualHost       = errors.New("gateway: cannot delete gateway declaration since it is connected to some virtual host")
	ErrHasCluster           = errors.New("gateway: cannot delete gateway declaration since it is connected to some cluster")
)

type service[R dto.GatewayDeclaration] struct {
	dao           dao.Dao
	entityService entity.ServiceInterface
	bus           bus.BusPublisher
}

func NewService(dao dao.Dao, entityService entity.ServiceInterface, busPublisher bus.BusPublisher) *service[dto.GatewayDeclaration] {
	return &service[dto.GatewayDeclaration]{dao: dao, entityService: entityService, bus: busPublisher}
}

func (r *service[R]) GetConfigRes() cfgres.ConfigRes[dto.GatewayDeclaration] {
	return cfgres.ConfigRes[dto.GatewayDeclaration]{
		Key: cfgres.ResourceKey{
			APIVersion: "nc.core.mesh/v3",
			Kind:       "GatewayDeclaration",
		},
		Applier: r,
	}
}

func (r *service[R]) Validate(ctx context.Context, res dto.GatewayDeclaration) (bool, string) {
	// all validations for this request must be performed inside the transaction, so they will be carried out by r.Apply func
	return true, ""
}

func (r *service[R]) IsOverriddenByCR(_ context.Context, res dto.GatewayDeclaration) bool {
	return res.Overridden
}

func IsInputError(err error) bool {
	for _, inputErr := range inputErrors {
		if inputErr == err {
			return true
		}
	}
	return false
}

func (r *service[R]) Apply(ctx context.Context, res dto.GatewayDeclaration) (any, error) {
	log.InfoC(ctx, "Request to apply declarative gateway configuration %+v", res)
	changes, err := r.dao.WithWTx(func(repo dao.Repository) error {
		existingNodegroup, err := repo.FindNodeGroupByName(res.Name)
		if err != nil {
			return log.ErrorC(ctx, err, "Could not find existing node group by name %s due to DAO error", res.Name)
		}
		if res.IsDeleteRequest() {
			return r.deleteGatewayDeclaration(ctx, repo, res, existingNodegroup)
		}
		return r.applyGatewayDeclaration(ctx, repo, res, existingNodegroup)
	})
	if err != nil {
		if goErr, ok := err.(*goerrors.Error); ok {
			if cause := goErr.Unwrap(); IsInputError(cause) {
				return nil, cause
			}
		}
		return nil, log.ErrorC(ctx, err, "Could not apply gateway declaration %+v, DAO transaction finished with error", res)
	}
	log.InfoC(ctx, "Gateway declaration %+v has been applied successfully", res)

	if len(changes) > 0 {
		if err = r.bus.Publish(bus.TopicChanges, events.NewChangeEventByNodeGroup(res.Name, changes)); err != nil {
			return nil, log.ErrorC(ctx, err, "Could not publish changes to event bus on topic %s", bus.TopicChanges)
		}
	}
	return map[string]string{"message": "gateway declaration applied successfully"}, nil
}

func (r *service[R]) GetAll(ctx context.Context) ([]dto.GatewayDeclaration, error) {
	nodeGroups, err := r.dao.FindAllNodeGroups()
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not get all node groups due to DAO error")
	}
	result := make([]dto.GatewayDeclaration, 0, len(nodeGroups))
	for _, nodeGroup := range nodeGroups {
		result = append(result, convertToDto(nodeGroup))
	}
	return result, nil
}

func convertToDto(nodeGroup *domain.NodeGroup) dto.GatewayDeclaration {
	return dto.GatewayDeclaration{
		Name:              nodeGroup.Name,
		GatewayType:       nodeGroup.GatewayType,
		AllowVirtualHosts: util.WrapValue(!nodeGroup.ForbidVirtualHosts),
	}
}

func (r *service[R]) deleteGatewayDeclaration(ctx context.Context, repo dao.Repository, req dto.GatewayDeclaration, existingNodegroup *domain.NodeGroup) error {
	if existingNodegroup == nil {
		return nil
	}
	routeConfigsToDrop, err := r.validateCanDeleteNodeGroup(ctx, repo, req)
	if err != nil {
		return err
	}
	for _, routeConfig := range routeConfigsToDrop {
		if err := repo.DeleteRouteConfigById(routeConfig.Id); err != nil {
			return log.ErrorC(ctx, err, "Could not cascade delete route config during node group %s deletion due to DAO error", req.Name)
		}
	}
	if err := repo.DeleteListenerByNodeGroupName(req.Name); err != nil {
		return log.ErrorC(ctx, err, "Could not cascade delete listener during node group %s deletion due to DAO error", req.Name)
	}
	if err := repo.DeleteNodeGroupByName(req.Name); err != nil {
		return log.ErrorC(ctx, err, "Could not delete node group %s due to DAO error", req.Name)
	}
	// need to generate at least one new envoy entity version for any entity so change event is processed properly
	return r.entityService.GenerateEnvoyEntityVersions(ctx, repo, req.Name, domain.ListenerTable)
}

// validateCanDeleteNodeGroup performs validation that node group has no related entities that could block deletion.
// Returns collection of RouteConfigurations to be dropped cascade during node group deletion;
// and error in case there are some related entities or some issue with DAO occurred.
func (r *service[R]) validateCanDeleteNodeGroup(ctx context.Context, repo dao.Repository, req dto.GatewayDeclaration) ([]*domain.RouteConfiguration, error) {
	// validate tls configs
	if relatedEntities, err := repo.FindAllTlsConfigsByNodeGroup(req.Name); err != nil {
		return nil, log.ErrorC(ctx, err, "Could not find tls configs by node group %s due to DAO error", req.Name)
	} else if len(relatedEntities) > 0 {
		return nil, ErrHasTlsConfig
	}
	// validate clusters
	if relatedEntities, err := repo.FindClusterByNodeGroup(&domain.NodeGroup{Name: req.Name}); err != nil {
		return nil, log.ErrorC(ctx, err, "Could not find clusters by node group %s due to DAO error", req.Name)
	} else if len(relatedEntities) > 0 {
		return nil, ErrHasCluster
	}
	// validate route configs
	routeConfigs, err := repo.FindRouteConfigsByNodeGroupId(req.Name)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not find route configs by node group %s due to DAO error", req.Name)
	}
	for _, routeConfig := range routeConfigs {
		vHosts, err := r.entityService.FindVirtualHostsByRouteConfig(repo, routeConfig.Id)
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Could not find virtual hosts by route config %s (id=%v) due to DAO error", routeConfig.Name, routeConfig.Id, req.Name)
		}
		if len(vHosts) > 0 {
			return nil, ErrHasVirtualHost
		}
	}
	return routeConfigs, nil
}

func (r *service[R]) applyGatewayDeclaration(ctx context.Context, repo dao.Repository, req dto.GatewayDeclaration, existingNodegroup *domain.NodeGroup) error {
	nodeGroup := &domain.NodeGroup{
		Name:        req.Name,
		GatewayType: req.GatewayType,
	}
	if req.AllowVirtualHosts != nil && !*req.AllowVirtualHosts {
		nodeGroup.ForbidVirtualHosts = true
	}

	if existingNodegroup == nil || existingNodegroup.GatewayType != nodeGroup.GatewayType ||
		existingNodegroup.ForbidVirtualHosts != nodeGroup.ForbidVirtualHosts {
		if err := r.validateCanApplyGatewayDeclaration(ctx, repo, req); err != nil {
			log.InfoC(ctx, "Attempt to register gateway declaration %+v rejected:\n %v", req, err)
			return err
		}

		if err := repo.SaveNodeGroup(nodeGroup); err != nil {
			return log.ErrorC(ctx, err, "Could not save node group %+v due to DAO error", *nodeGroup)
		}
		return r.generateEnvoyEntityVersions(ctx, repo, req.Name)
	}
	return nil
}

func (r *service[R]) generateEnvoyEntityVersions(ctx context.Context, repo dao.Repository, gateway string) error {
	// update only route configs and listeners - only those entities depend on gateway type when building envoy cache
	return r.entityService.GenerateEnvoyEntityVersions(ctx, repo, gateway, domain.RouteConfigurationTable, domain.ListenerTable)
}

func (r *service[R]) validateCanApplyGatewayDeclaration(ctx context.Context, repo dao.Repository, req dto.GatewayDeclaration) error {
	// cases when gateway declaration can conflict with existing configuration:
	// case 1: AllowVirtualHosts: false, but there are some virtualHosts registered
	// case 2: gateway type ingress, but there are some routes leading not to PGW
	virtualHostsForbidden := req.AllowVirtualHosts != nil && !*req.AllowVirtualHosts
	if !virtualHostsForbidden && req.GatewayType != domain.Ingress {
		// nothing to validate, there can be no conflicts
		return nil
	}

	virtualHosts, err := r.entityService.FindVirtualHostsByNodeGroup(repo, req.Name)
	if err != nil {
		return log.ErrorC(ctx, err, "Could not find virtual hosts for node group %s using DAO", req.Name)
	}

	// validate case 1: no conflicting virtualHosts
	if virtualHostsForbidden {
		if len(virtualHosts) > 1 {
			return ErrMultipleVirtualHosts
		}
		if len(virtualHosts) == 1 && !virtualHosts[0].HasGenericDomain() {
			return ErrInvalidVirtualHost
		}
	}

	// validate case 2: no conflicting routes
	if req.GatewayType == domain.Ingress {
		for _, virtualHost := range virtualHosts {
			for _, route := range virtualHost.Routes {
				if route.DirectResponseCode != 0 {
					continue
				}
				if err = isAllowedCluster(ctx, repo, route.ClusterName); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func isAllowedCluster(ctx context.Context, repo dao.Repository, clusterName string) error {
	endpoints, err := repo.FindEndpointsByClusterName(clusterName)
	if err != nil {
		return log.ErrorC(ctx, err, "Could not load endpoints by cluster name %s due to DAO error", clusterName)
	}
	for _, endpoint := range endpoints {
		if endpoint.Address != domain.PublicGateway && endpoint.Address != domain.PrivateGateway {
			return ErrInvalidEndpoint
		}
	}
	return nil
}
