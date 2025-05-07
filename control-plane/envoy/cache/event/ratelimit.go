package event

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/action"
)

func (parser *changeEventParserImpl) processRateLimitChanges(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, changes []memdb.Change) {
	for _, change := range changes {
		if change.Deleted() {
			rateLimit := change.Before.(*domain.RateLimit)
			parser.updateRateLimit(actions, entityVersions, nodeGroup, rateLimit)
		} else {
			rateLimit := change.After.(*domain.RateLimit)
			parser.updateRateLimit(actions, entityVersions, nodeGroup, rateLimit)
		}
	}
}

func (parser *changeEventParserImpl) updateRateLimit(actions action.ActionsMap, entityVersions map[string]string, nodeGroup string, rateLimit *domain.RateLimit) {
	vHostsToUpdate, err := findVirtualHostsToUpdateByRateLimit(parser.dao, rateLimit)
	if err != nil {
		logger.Panicf("Could not find virtual hosts to update by rate limit change:\n %v", err)
	}
	for vHostId := range vHostsToUpdate {
		vHost, err := parser.dao.FindVirtualHostById(vHostId)
		if err != nil {
			logger.Panicf("Could not find virtual host by id to update with rate limit change:\n %v", err)
		}
		parser.updateVirtualHost(actions, entityVersions, nodeGroup, vHost)
	}
}

func (builder *compositeUpdateBuilder) processRateLimitChanges(changes []memdb.Change) {
	logger.Debug("Processing rate limit multiple change event")
	for _, change := range changes {
		var rateLimit *domain.RateLimit = nil
		if change.Deleted() {
			rateLimit = change.Before.(*domain.RateLimit)
		} else {
			rateLimit = change.After.(*domain.RateLimit)
		}
		builder.updateRateLimit(rateLimit)
	}
}

func (builder *compositeUpdateBuilder) updateRateLimit(rateLimit *domain.RateLimit) {
	vHostsToUpdate, err := findVirtualHostsToUpdateByRateLimit(builder.repo, rateLimit)
	if err != nil {
		logger.Panicf("Could not find virtual hosts to update by rate limit change:\n %v", err)
	}
	for vHostId := range vHostsToUpdate {
		builder.updateVirtualHost(vHostId)
	}
}

func findVirtualHostsToUpdateByRateLimit(repo dao.Repository, rateLimit *domain.RateLimit) (map[int32]bool, error) {
	virtualHostsToUpdate := make(map[int32]bool)

	allVHosts, err := repo.FindAllVirtualHosts()
	if err != nil {
		logger.Errorf("Failed to load all virtual hosts using DAO:\n %v", err)
		return nil, err
	}
	for _, vHost := range allVHosts {
		if vHost.RateLimitId == rateLimit.Name {
			virtualHostsToUpdate[vHost.Id] = true
		}
	}

	routes, err := repo.FindRoutesByRateLimit(rateLimit.Name)
	if err != nil {
		logger.Errorf("Failed to find routes by rate limit %s using DAO:\n %v", rateLimit.Name, err)
		return nil, err
	}
	for _, route := range routes {
		virtualHostsToUpdate[route.VirtualHostId] = true
	}
	return virtualHostsToUpdate, nil
}
