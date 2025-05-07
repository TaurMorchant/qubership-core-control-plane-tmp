package routeconfig

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/routekey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	routes_utils "github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/routes"
	"strings"
)

//go:generate mockgen -source=routepreparer.go -destination=../../../../test/mock/envoy/cache/builder/routeconfig/stub_routepreparer.go -package=mock_routeconfig
type RoutePreparer interface {
	Prepare(routes []*domain.Route) []*domain.Route
}

// RouteStatefulSessionPreparer resolves which stateful session setting should be applied for each route.
type RouteStatefulSessionPreparer struct {
	dao           dao.Repository
	entityService *entity.Service
}

func NewRouteStatefulSessionPreparer(dao dao.Repository, entityService *entity.Service) *RouteStatefulSessionPreparer {
	return &RouteStatefulSessionPreparer{dao: dao, entityService: entityService}
}

func (p *RouteStatefulSessionPreparer) resolveStatefulSessionConfig(route *domain.Route) *domain.StatefulSession {
	clusterFamilyName := clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(route.ClusterName)
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(route.ClusterName)
	nodeGroup, err := p.entityService.FindRouteNodeGroup(p.dao, route)
	if err != nil {
		logger.Panicf("Failed to find route nodeGroup using entity.Service:\n %v", err)
	}

	sessions, err := p.dao.FindStatefulSessionConfigsByClusterAndVersion(clusterFamilyName, namespace, route.DeploymentVersionVal)
	if err != nil {
		logger.Panicf("Loading stateful sessions by cluster from DAO in RouteStatefulSessionPreparer failed:\n %v", err)
	}

	var perClusterConfig *domain.StatefulSession
	for _, session := range sessions {
		if !util.SliceContains(session.Gateways, nodeGroup.Name) {
			continue
		}

		// this stateful session might be per-route, per-endpoint or per-cluster configuration, need to check
		if anotherRoute, err := p.dao.FindRouteByStatefulSession(session.Id); err != nil {
			logger.Panicf("Loading route by stateful session id from DAO in RouteStatefulSessionPreparer failed:\n %v", err)
		} else if anotherRoute != nil { // configuration bound to some another route, skipping it
			continue
		}
		if endpoint, err := p.dao.FindEndpointByStatefulSession(session.Id); err != nil {
			logger.Panicf("Loading endpoint by stateful session id from DAO in RouteStatefulSessionPreparer failed:\n %v", err)
		} else if endpoint == nil { // this is per-cluster configuration
			perClusterConfig = session.Clone()
		} else { // this is per-endpoint configuration
			return session
		}
	}
	return perClusterConfig
}

func (p *RouteStatefulSessionPreparer) Prepare(routes []*domain.Route) []*domain.Route {
	result := make([]*domain.Route, 0, len(routes))
	for _, route := range routes {
		if route.StatefulSession == nil {
			effectiveStatefulSession := p.resolveStatefulSessionConfig(route)
			if effectiveStatefulSession != nil {
				logger.Debugf("Resolved effective stateful session for route %v [%v]: %+v", route.Prefix, route.DeploymentVersion, effectiveStatefulSession)
				routeClone := route.Clone()
				routeClone.StatefulSession = effectiveStatefulSession
				result = append(result, routeClone)
			} else {
				result = append(result, route)
			}
		} else {
			// route-level stateful session config has the highest priority, so just add route without changes
			result = append(result, route)
		}
	}
	return result
}

type ExtraSlashRoutePreparer struct {
	beforePreparer RoutePreparer
}

func NewExtraSlashRoutePreparer(beforePreparer RoutePreparer) *ExtraSlashRoutePreparer {
	return &ExtraSlashRoutePreparer{beforePreparer: beforePreparer}
}

// checkSlash checks that prefix and prefixRewrite both end with "/", or both end without it.
// If prefix ends with "/", but prefixRewrite ends without "/", method adds "to end".
// Exception when prefix ends with "/" and prefixRewrite="/", nothing will be changed
// Examples :
// 1) prefix "/api/v1/ext-frontend-api/customers/import/some-operation/" prefixRewrite "/api/v1/some-operation" -> prefixRewrite "/api/v1/some-operation/"
// 2) prefix "/api/v1/ext-frontend-api/customers/import/some-operation" prefixRewrite "/api/v1/some-operation/" -> prefixRewrite "/api/v1/some-operation"
// 3) prefix "/api/v1/ext-frontend-api" prefixRewrite "/" -> prefixRewrite "/"
func (preparer *ExtraSlashRoutePreparer) checkSlash(sourceRoute *domain.Route) *domain.Route {
	// TODO: do we really need to clone route? Can we affect InMemDB or something by changing it here?
	r := sourceRoute.Clone()
	if r.PrefixRewrite != "" {
		if strings.HasSuffix(r.Prefix, "/") && !strings.HasSuffix(r.PrefixRewrite, "/") {
			r.PrefixRewrite += "/"
		} else if !strings.HasSuffix(r.Prefix, "/") && strings.HasSuffix(r.PrefixRewrite, "/") {
			if r.PrefixRewrite != "/" {
				r.PrefixRewrite = r.PrefixRewrite[:len(r.PrefixRewrite)-1]
			}
		}
	}
	return r
}

func (preparer *ExtraSlashRoutePreparer) createRoutePair(r *domain.Route) *domain.Route {
	prefixRewrite := r.PrefixRewrite
	var prefix string
	if strings.HasSuffix(r.Prefix, "/") {
		prefix = r.Prefix[:len(r.Prefix)-1]
		if r.PrefixRewrite != "" {
			prefixRewrite = r.PrefixRewrite[:len(r.PrefixRewrite)-1]
		}
	} else {
		prefix = r.Prefix + "/"
		if r.PrefixRewrite != "" && r.PrefixRewrite != "/" {
			prefixRewrite = r.PrefixRewrite + "/"
		}
	}
	anotherRoute := r.Clone()
	anotherRoute.Prefix = prefix
	anotherRoute.PrefixRewrite = prefixRewrite
	return anotherRoute
}

func (preparer *ExtraSlashRoutePreparer) Prepare(routes []*domain.Route) []*domain.Route {
	result := preparer.beforePreparer.Prepare(routes)
	routePairs := make([]*domain.Route, 0, len(result))
	for _, route := range result {
		if route.Prefix == "" && route.Regexp != "" {
			continue
		}
		route = preparer.checkSlash(route)
		if route.Prefix != "/" {
			routePairs = append(routePairs, preparer.createRoutePair(route))
		}
	}
	return append(result, routePairs...)
}

type SortRoutePreparer struct {
	beforePreparer RoutePreparer
}

func NewSortRoutePreparer(beforePreparer RoutePreparer) *SortRoutePreparer {
	return &SortRoutePreparer{beforePreparer: beforePreparer}
}

func (preparer *SortRoutePreparer) Prepare(routes []*domain.Route) []*domain.Route {
	allRoutes := preparer.beforePreparer.Prepare(routes)
	return routes_utils.OrderRoutesForEnvoy(allRoutes)
}

type RouteMultiVersionPreparer struct {
	beforePreparer   RoutePreparer
	dao              dao.Repository
	routeTransformer *RouteTransformer
}

func NewRouteMultiVersionPreparer(beforePreparer RoutePreparer, dao dao.Repository, routeTransformer *RouteTransformer) *RouteMultiVersionPreparer {
	return &RouteMultiVersionPreparer{beforePreparer: beforePreparer, dao: dao, routeTransformer: routeTransformer}
}

func (preparer *RouteMultiVersionPreparer) Prepare(routes []*domain.Route) []*domain.Route {
	routes = preparer.beforePreparer.Prepare(routes)
	allVersions, err := preparer.dao.FindAllDeploymentVersions()
	if err != nil {
		panic(err)
	}
	activeAndCandidateVersions := make([]string, 0)
	for _, ver := range allVersions {
		if ver.Stage == "ACTIVE" || ver.Stage == "CANDIDATE" {
			activeAndCandidateVersions = append(activeAndCandidateVersions, ver.Version)
		}
	}

	clustersHashPolicy := make(map[string][]*domain.HashPolicy, 0)
	multiVersionRoutes := make(map[string]*RouteGroup, 0)
	for _, route := range routes {
		if route.ClusterName != "" {
			policies, found := clustersHashPolicy[route.ClusterName]
			if !found {
				policies = preparer.collectAllVersionsHashPolicy(route.ClusterName, activeAndCandidateVersions, route.Id)
				clustersHashPolicy[route.ClusterName] = policies
			}
			route.HashPolicies = policies
		}

		matcherKey := routekey.GenerateKey(*route)
		routeGroup, found := multiVersionRoutes[matcherKey]
		if !found {
			routeGroup = NewRouteGroup(allVersions)
			multiVersionRoutes[matcherKey] = routeGroup
		}
		routeGroup.AddRoute(route)
	}
	routeGroups := make([]*RouteGroup, len(multiVersionRoutes))
	idx := 0
	for _, group := range multiVersionRoutes {
		routeGroups[idx] = group
		idx++
	}
	return preparer.routeTransformer.Transform(routeGroups)
}

func (preparer *RouteMultiVersionPreparer) collectAllVersionsHashPolicy(clusterName string, activeAndCandidateVersions []string, routeId int32) []*domain.HashPolicy {
	allHashPolicies := make([]*domain.HashPolicy, 0)
	hashPolicies, err := preparer.dao.FindHashPolicyByClusterAndVersions(clusterName, activeAndCandidateVersions...)
	routeHashPolicies, err := preparer.dao.FindHashPolicyByRouteId(routeId)
	if routeHashPolicies != nil && len(routeHashPolicies) > 0 {
		// routeHashPolicies are the policies added to the route directly
		hashPolicies = append(hashPolicies, routeHashPolicies...)
	}
	if err != nil {
		panic(err)
	}
	// Add each hash policy that is not already in allHashPolicies
	for _, hashPolicy := range hashPolicies {
		needToAdd := true
		for _, containedHashPolicy := range allHashPolicies {
			if hashPolicy.Equals(containedHashPolicy) {
				needToAdd = false
				break
			}
		}
		if needToAdd {
			allHashPolicies = append(allHashPolicies, hashPolicy)
		}
	}
	return allHashPolicies
}
