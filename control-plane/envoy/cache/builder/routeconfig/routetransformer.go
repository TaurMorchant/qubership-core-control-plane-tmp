package routeconfig

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"

type RouteTransformer struct {
	transformationRulesChain []TransformationRule
}

func NewRouteTransformer(rules ...TransformationRule) *RouteTransformer {
	return &RouteTransformer{transformationRulesChain: rules}
}

func (rt *RouteTransformer) Transform(routeGroups []*RouteGroup) []*domain.Route {
	result := make([]*domain.Route, 0)
	for _, routeGroup := range routeGroups {
		for _, rule := range rt.transformationRulesChain {
			if rule.IsApplicable(routeGroup) {
				result = append(result, rule.Apply(routeGroup)...)
			}
		}
	}
	return result
}

type TransformationRule interface {
	IsApplicable(routeGroup *RouteGroup) bool
	Apply(routeGroup *RouteGroup) []*domain.Route
}

type GenericVersionedRouteTransformRule struct{}

func NewGenericVersionedRouteTransformRule() *GenericVersionedRouteTransformRule {
	return &GenericVersionedRouteTransformRule{}
}

func (rule *GenericVersionedRouteTransformRule) IsApplicable(routeGroup *RouteGroup) bool {
	return routeGroup.IsMultiVersioned() && routeGroup.HasActiveVersion()
}

func (rule *GenericVersionedRouteTransformRule) Apply(routeGroup *RouteGroup) []*domain.Route {
	result := make([]*domain.Route, 0)
	activeVersionRoute := routeGroup.GetActiveVersionRoute()
	result = append(result, activeVersionRoute)

	// If not active routes have differences
	for _, route := range routeGroup.GetNoActiveVersionRoutes() {
		if route.DirectResponseCode == 0 {
			if route.RouteAction() != activeVersionRoute.RouteAction() {
				newRoute := route.Clone()
				newRoute.HeaderMatchers = append(newRoute.HeaderMatchers,
					&domain.HeaderMatcher{Name: "x-version", ExactMatch: newRoute.DeploymentVersionVal.Version})
				result = append(result, newRoute)
			}
		} else {
			if route.DirectResponseCode != activeVersionRoute.DirectResponseCode {
				newRoute := route.Clone()
				newRoute.HeaderMatchers = append(newRoute.HeaderMatchers,
					&domain.HeaderMatcher{Name: "x-version", ExactMatch: newRoute.DeploymentVersionVal.Version})
				result = append(result, newRoute)
			}
		}
	}
	return result
}

type NoActiveRouteTransformer struct{}

func NewNoActiveRouteTransformer() *NoActiveRouteTransformer {
	return &NoActiveRouteTransformer{}
}

func (rule *NoActiveRouteTransformer) IsApplicable(routeGroup *RouteGroup) bool {
	return !routeGroup.HasActiveVersion()
}

func (rule *NoActiveRouteTransformer) Apply(routeGroup *RouteGroup) []*domain.Route {
	result := make([]*domain.Route, 0)
	for _, route := range routeGroup.GetNoActiveVersionRoutes() {
		route = route.Clone()
		route.HeaderMatchers = append(route.HeaderMatchers,
			&domain.HeaderMatcher{Name: "x-version", ExactMatch: route.DeploymentVersionVal.Version})
		result = append(result, route)
	}
	return result
}

type SimpleRouteTransformationRule struct{}

func NewSimpleRouteTransformationRule() *SimpleRouteTransformationRule {
	return &SimpleRouteTransformationRule{}
}

func (rule *SimpleRouteTransformationRule) IsApplicable(routeGroup *RouteGroup) bool {
	return !routeGroup.IsMultiVersioned() && routeGroup.HasActiveVersion()
}

func (rule *SimpleRouteTransformationRule) Apply(routeGroup *RouteGroup) []*domain.Route {
	return routeGroup.Routes()
}
