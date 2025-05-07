package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
)

func (parser *changeEventParserImpl) processVirtualHostChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			virtualHost := change.Before.(*domain.VirtualHost)
			parser.updateVirtualHost(actions, entityVersions, nodeGroup, virtualHost)
		} else {
			virtualHost := change.After.(*domain.VirtualHost)
			parser.updateVirtualHost(actions, entityVersions, nodeGroup, virtualHost)
		}
	}
}

func (parser *changeEventParserImpl) processVirtualHostDomainChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		var virtualHostDomain *domain.VirtualHostDomain
		if change.Deleted() {
			virtualHostDomain = change.Before.(*domain.VirtualHostDomain)
		} else {
			virtualHostDomain = change.After.(*domain.VirtualHostDomain)
		}
		virtualHost, err := parser.dao.FindVirtualHostById(virtualHostDomain.VirtualHostId)
		if err != nil {
			logger.Panicf("Failed to find virtual host by id using DAO: %v", err)
		}
		if virtualHost != nil {
			parser.updateVirtualHost(actions, entityVersions, nodeGroup, virtualHost)
		}
	}
}

func (parser *changeEventParserImpl) updateVirtualHost(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, virtualHost *domain.VirtualHost) {
	routeConfig, err := parser.dao.FindRouteConfigById(virtualHost.RouteConfigurationId)
	if err != nil {
		logger.Panicf("Failed to find route configuration by id using DAO: %v", err)
	}
	if routeConfig != nil {
		parser.updateRouteConfig(actions, entityVersions, nodeGroup, routeConfig)
	}
}

func (builder *compositeUpdateBuilder) processVirtualHostChanges(changes []memdb.Change) {
	for _, change := range changes {
		var virtualHost *domain.VirtualHost = nil
		if change.Deleted() {
			virtualHost = change.Before.(*domain.VirtualHost)
		} else {
			virtualHost = change.After.(*domain.VirtualHost)
		}
		builder.updateRouteConfig(virtualHost.RouteConfigurationId)
	}
}

func (builder *compositeUpdateBuilder) updateVirtualHost(virtualHostId int32) {
	logger.Debugf("Updating virtual host %d", virtualHostId)
	vHost, err := builder.repo.FindVirtualHostById(virtualHostId)
	if err != nil {
		logger.Panicf("Failed to find virtual host by id using DAO: %v", err)
	}
	if vHost != nil { // if nil, this vHost is being deleted by domain.VirtualHost change event
		builder.updateRouteConfig(vHost.RouteConfigurationId)
	}
}

func (builder *compositeUpdateBuilder) processVirtualHostDomainChanges(changes []memdb.Change) {
	for _, change := range changes {
		var virtualHostDomain *domain.VirtualHostDomain = nil
		if change.Deleted() {
			virtualHostDomain = change.Before.(*domain.VirtualHostDomain)
		} else {
			virtualHostDomain = change.After.(*domain.VirtualHostDomain)
		}
		builder.updateVirtualHost(virtualHostDomain.VirtualHostId)
	}
}
