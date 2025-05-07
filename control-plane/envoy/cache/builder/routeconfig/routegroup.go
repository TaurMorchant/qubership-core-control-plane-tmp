package routeconfig

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"

type RouteGroup struct {
	allVersions    []*domain.DeploymentVersion
	routesVersions []*domain.DeploymentVersion
	routes         []*domain.Route
}

func NewRouteGroup(allVersions []*domain.DeploymentVersion) *RouteGroup {
	return &RouteGroup{allVersions: allVersions, routes: make([]*domain.Route, 0), routesVersions: make([]*domain.DeploymentVersion, 0)}
}

func (rg *RouteGroup) Routes() []*domain.Route {
	return rg.routes
}

func (rg *RouteGroup) AddRoute(route *domain.Route) {
	rg.routes = append(rg.routes, route)
	rg.routesVersions = append(rg.routesVersions, route.DeploymentVersionVal)
}

func (rg *RouteGroup) IsMultiVersioned() bool {
	return rg.routesVersions != nil && len(rg.routesVersions) > 1
}

func (rg *RouteGroup) GetActiveVersionRoute() *domain.Route {
	for _, route := range rg.routes {
		if route.DeploymentVersionVal.Stage == "ACTIVE" {
			return route
		}
	}
	return nil
}

func (rg *RouteGroup) GetNoActiveVersionRoutes() []*domain.Route {
	result := make([]*domain.Route, 0, len(rg.routes))
	for _, route := range rg.routes {
		if route.DeploymentVersionVal.Stage != "ACTIVE" {
			result = append(result, route)
		}
	}
	return result
}

func (rg *RouteGroup) HasActiveVersion() bool {
	for _, route := range rg.routes {
		if route.DeploymentVersionVal.Stage == "ACTIVE" {
			return true
		}
	}
	return false
}
