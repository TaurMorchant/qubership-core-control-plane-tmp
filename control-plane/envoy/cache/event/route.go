package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processRouteChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			route := change.Before.(*domain.Route)
			parser.updateRoute(actions, entityVersions, nodeGroup, route)
		} else {
			route := change.After.(*domain.Route)
			parser.updateRoute(actions, entityVersions, nodeGroup, route)
		}
	}
}

func (parser *changeEventParserImpl) processHeaderMatcherChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			headerMatcher := change.Before.(*domain.HeaderMatcher)
			parser.updateHeaderMatcher(actions, entityVersions, nodeGroup, headerMatcher)
		} else {
			headerMatcher := change.After.(*domain.HeaderMatcher)
			parser.updateHeaderMatcher(actions, entityVersions, nodeGroup, headerMatcher)
		}
	}
}

func (parser *changeEventParserImpl) updateRoute(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, route *domain.Route) {
	virtualHost, err := parser.dao.FindVirtualHostById(route.VirtualHostId)
	if err != nil {
		logger.Panicf("Failed to find virtual host by id using DAO: %v", err)
	}
	if virtualHost != nil {
		parser.updateVirtualHost(actions, entityVersions, nodeGroup, virtualHost)
	}
}

func (parser *changeEventParserImpl) updateHeaderMatcher(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, headerMatcher *domain.HeaderMatcher) {
	route, err := parser.dao.FindRouteById(headerMatcher.RouteId)
	if err != nil {
		logger.Panicf("Failed to find route by id using DAO: %v", err)
	}
	if route != nil {
		parser.updateRoute(actions, entityVersions, nodeGroup, route)
	}
}

func (builder *compositeUpdateBuilder) processRouteChanges(changes []memdb.Change) {
	for _, change := range changes {
		var route *domain.Route = nil
		if change.Deleted() {
			route = change.Before.(*domain.Route)
		} else {
			route = change.After.(*domain.Route)
		}
		builder.updateVirtualHost(route.VirtualHostId)
	}
}

func (builder *compositeUpdateBuilder) processHeaderMatcherChanges(changes []memdb.Change) {
	for _, change := range changes {
		var headerMatcher *domain.HeaderMatcher = nil
		if change.Deleted() {
			headerMatcher = change.Before.(*domain.HeaderMatcher)
		} else {
			headerMatcher = change.After.(*domain.HeaderMatcher)
		}
		route, err := builder.repo.FindRouteById(headerMatcher.RouteId)
		if err != nil {
			logger.Panicf("Failed to find route by id using DAO: %v", err)
		}
		if route != nil { // if nil, route is deleted and its deletion is processed by domain.Route change event
			builder.updateVirtualHost(route.VirtualHostId)
		}
	}
}
