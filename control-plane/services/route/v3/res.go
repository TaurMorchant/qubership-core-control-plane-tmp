package v3

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	v3 "github.com/netcracker/qubership-core-control-plane/restcontrollers/v3"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/services/route"
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

type routingRequestResource struct {
	validator v3.RequestValidator
	service   v3.RouteService
}

func (c *routingRequestResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "RouteConfiguration",
	}
}

func (c *routingRequestResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.RoutingConfigRequestV3{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			req := *entity.(*dto.RoutingConfigRequestV3)
			return c.validator.Validate(req)
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			routingReq := entity.(*dto.RoutingConfigRequestV3)
			if namespaceRaw, ok := md["namespace"]; ok {
				if namespace, ok := namespaceRaw.(string); ok {
					routingReq.Namespace = strings.TrimSpace(namespace)
				} else {
					return nil, errors.New("field 'namespace' must have string type")
				}
			}
			if err := c.service.RegisterRoutingConfig(ctx, *routingReq); err != nil {
				return nil, errors.Wrap(err, "failed to register routes via v3 api")
			}
			return nil, nil
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.RoutingConfigRequestV3).Overridden
		},
	}
}

func (s *Service) GetRoutingRequestResource() cfgres.Resource {
	return &routingRequestResource{
		validator: dto.RoutingV3RequestValidator{},
		service:   s,
	}
}

type virtualServiceResource struct {
	validator v3.RequestValidator
	service   v3.RouteService
}

func (v *virtualServiceResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "VirtualService",
	}
}

func (v *virtualServiceResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.VirtualService{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			if ok, message := route.ValidateMetadataStringField(md, "name"); !ok {
				return false, message
			}
			if ok, message := route.ValidateMetadataStringField(md, "gateway"); !ok {
				return false, message
			}
			req := *entity.(*dto.VirtualService)
			return v.validator.ValidateVirtualService(req, []string{md["gateway"].(string)})
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			vs := entity.(*dto.VirtualService)
			vs.Name = md["name"].(string)
			nodeGroup := md["gateway"].(string)
			if err := v.service.CreateVirtualService(ctx, nodeGroup, *vs); err != nil {
				return nil, errors.Wrap(err, "creating virtual service caused error")
			}
			return nil, nil
		},
		IsOverriddenByCR: func(ctx context.Context, md cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.VirtualService).Overridden
		},
	}
}

func (s *Service) GetVirtualServiceResource() cfgres.Resource {
	return &virtualServiceResource{
		service:   s,
		validator: dto.RoutingV3RequestValidator{},
	}
}

type deleteRouteResource struct {
	service v3.RouteService
}

func (d deleteRouteResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "RoutesDrop",
	}
}

func (d deleteRouteResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf([]dto.RouteDeleteRequestV3{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			return true, ""
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			requests := *entity.(*[]dto.RouteDeleteRequestV3)
			if namespaceRaw, ok := md["namespace"]; ok {
				if namespace, ok := namespaceRaw.(string); ok {
					namespace = strings.TrimSpace(namespace)
					for i, request := range requests {
						request.Namespace = namespace
						requests[i] = request
					}
				} else {
					return nil, errors.New("field 'namespace' must have string type")
				}
			}
			deletedRoutes, err := d.service.DeleteRoutes(ctx, requests)
			if err != nil {
				return nil, errors.Wrap(err, "removing routes caused error")
			}
			return deletedRoutes, nil
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			requests := *entity.(*[]dto.RouteDeleteRequestV3)
			if len(requests) == 0 {
				return false
			}
			overridden := requests[0].Overridden
			for _, request := range requests {
				if overridden != request.Overridden {
					return false
				}
				overridden = request.Overridden
			}
			return overridden
		},
	}
}

func (s *Service) GetRoutesDropResource() cfgres.Resource {
	return &deleteRouteResource{
		service: s,
	}
}
