package ui

import (
	"context"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	routes_utils "github.com/netcracker/qubership-core-control-plane/util/routes"
	"strings"
)

type V3Service struct {
	ds         DataStorage
	relService RelationsService
	sorter     RouteSorter
}

func NewV3Service(rep dao.Repository, entityService *entity.Service, sorter RouteSorter) *V3Service {
	return &V3Service{ds: rep, relService: &entityServiceWrapper{rep: rep, service: entityService}, sorter: sorter}
}

type DataStorage interface {
	FindAllRouteConfigs() ([]*domain.RouteConfiguration, error)
	FindVirtualHostsByRouteConfigurationId(configId int32) ([]*domain.VirtualHost, error)
	FindAllDeploymentVersions() ([]*domain.DeploymentVersion, error)
	FindRoutesByVirtualHostIdAndDeploymentVersion(vhId int32, version string) ([]*domain.Route, error)
	FindRouteByUuid(uuid string) (*domain.Route, error)
	FindVirtualHostById(vhId int32) (*domain.VirtualHost, error)
	FindDeploymentVersion(version string) (*domain.DeploymentVersion, error)
	FindRouteConfigById(routeConfigurationId int32) (*domain.RouteConfiguration, error)
	FindClusterByName(key string) (*domain.Cluster, error)
	FindEndpointsByClusterIdAndDeploymentVersion(clusterId int32, dVersions *domain.DeploymentVersion) ([]*domain.Endpoint, error)
	FindHeaderMatcherByRouteId(routeId int32) ([]*domain.HeaderMatcher, error)
}

type RelationsService interface {
	GetClustersWithRelations() ([]*domain.Cluster, error)
}

type RouteSorter interface {
	Sort(routes []*domain.Route) []*domain.Route
}

func (s *V3Service) GetAllSimplifiedRouteConfigs(ctx context.Context) ([]SimplifiedRouteConfig, error) {
	routeConfigs, err := s.ds.FindAllRouteConfigs()
	if err != nil {
		return nil, errors.WrapPrefix(err, "finding all RouteConfigs caused error", 1)
	}
	if len(routeConfigs) == 0 {
		log.WarnC(ctx, "There is no route-configurations to return. Return empty array.")
		return []SimplifiedRouteConfig{}, nil
	}
	var virtualHosts []*domain.VirtualHost
	for _, routeConfig := range routeConfigs {
		vhs, err := s.ds.FindVirtualHostsByRouteConfigurationId(routeConfig.Id)
		if err != nil {
			return nil, errors.WrapPrefix(err, fmt.Sprintf("finding VirtualHosts for RouteConfigurationId=%d caused error", routeConfig.Id), 1)
		}
		routeConfig.VirtualHosts = vhs
		virtualHosts = append(virtualHosts, vhs...)
	}

	dVersions, err := s.ds.FindAllDeploymentVersions()
	if err != nil {
		return nil, errors.WrapPrefix(err, "finding all DeploymentVersions caused error", 1)
	}
	virtualHostsVersions := make(map[int32][]*domain.DeploymentVersion)
	for _, dVersion := range dVersions {
		for _, virtualHost := range virtualHosts {
			routes, err := s.ds.FindRoutesByVirtualHostIdAndDeploymentVersion(virtualHost.Id, dVersion.Version)
			if err != nil {
				return nil, errors.Wrap(err, 1)
			}
			if len(routes) > 0 {
				virtualHostsVersions[virtualHost.Id] = append(virtualHostsVersions[virtualHost.Id], dVersion)
			}
		}
	}
	simplRouteConfigs := adaptRouteConfigsToUI(routeConfigs, virtualHostsVersions)
	return simplRouteConfigs, nil
}

func (s *V3Service) GetRoutesPage(ctx context.Context, params SearchRoutesParameters) (PageRoutes, error) {
	if params.Size < 1 {
		return PageRoutes{}, errors.New("size of routes per page must be greater than 0")
	}
	if params.Page < 1 {
		return PageRoutes{}, errors.New("number of page must be greater than 0")
	}
	routes, err := s.ds.FindRoutesByVirtualHostIdAndDeploymentVersion(params.VirtualHostId, params.Version)
	if err != nil {
		return PageRoutes{}, errors.Wrap(err, 1)
	}
	filteredRoutes := routes
	if params.Search != "" {
		filteredRoutes = filterRoutes(routes, func(route *domain.Route) bool {
			return strings.Contains(route.Prefix, params.Search) ||
				strings.Contains(route.ClusterName, params.Search) ||
				strings.Contains(route.Regexp, params.Search)
		})
	}
	for _, route := range filteredRoutes {
		route.HeaderMatchers, err = s.ds.FindHeaderMatcherByRouteId(route.Id)
		if err != nil {
			return PageRoutes{}, errors.Wrap(err, 1)
		}
	}
	filteredRoutes = s.sorter.Sort(filteredRoutes)
	lowerBound := params.LowerBound()
	upperBound := params.UpperBound()
	if lowerBound+1 > len(filteredRoutes) {
		if params.Size > len(filteredRoutes) {
			params.Page = 1
			params.Size = len(filteredRoutes)
		} else {
			params.Page = len(filteredRoutes) / params.Size
		}
		lowerBound = params.LowerBound()
	}
	if upperBound > len(filteredRoutes) {
		upperBound = len(filteredRoutes)
	}
	rawPageRoutes := filteredRoutes[lowerBound:upperBound]
	pageRoutes := adaptRoutesToUI(rawPageRoutes)
	virtualHost, err := s.ds.FindVirtualHostById(params.VirtualHostId)
	if err != nil {
		return PageRoutes{}, errors.Wrap(err, 1)
	}
	dVersion, err := s.ds.FindDeploymentVersion(params.Version)
	if err != nil {
		return PageRoutes{}, errors.Wrap(err, 1)
	}
	routeConfig, err := s.ds.FindRouteConfigById(virtualHost.RouteConfigurationId)
	if err != nil {
		return PageRoutes{}, errors.Wrap(err, 1)
	}
	return PageRoutes{
		TotalCount:      len(filteredRoutes),
		Routes:          pageRoutes,
		NodeGroup:       routeConfig.NodeGroupId,
		VirtualHostName: virtualHost.Name,
		VersionName:     dVersion.Version,
		VersionStage:    dVersion.Stage,
	}, nil
}

type routeSorter struct{}

func DefaultRouteSorter() *routeSorter {
	return &routeSorter{}
}

func (r routeSorter) Sort(routes []*domain.Route) []*domain.Route {
	return routes_utils.OrderRoutesForEnvoy(routes)
}

func (s *V3Service) GetAllClusters(ctx context.Context) ([]Cluster, error) {
	clusters, err := s.relService.GetClustersWithRelations()
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}
	versions, err := s.ds.FindAllDeploymentVersions()
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}
	uiClusters := adaptClusterToUI(clusters, versions)
	return uiClusters, nil
}

func (s *V3Service) GetRouteDetails(ctx context.Context, uuid string) (RouteDetails, error) {
	route, err := s.ds.FindRouteByUuid(uuid)
	if err != nil {
		return RouteDetails{}, errors.Wrap(err, 1)
	}
	routeDetails := RouteDetails{
		Path:                   route.Path,
		PathRewrite:            route.PathRewrite,
		Prefix:                 route.Prefix,
		PrefixRewrite:          route.PrefixRewrite,
		Regexp:                 route.Regexp,
		RegexpRewrite:          route.RegexpRewrite,
		ClusterName:            route.ClusterName,
		DirectResponse:         route.DirectResponseCode,
		HostRewrite:            route.HostRewrite,
		RequestHeadersToRemove: route.RequestHeadersToRemove,
	}
	if route.Timeout.Valid {
		value := route.Timeout.Int64
		routeDetails.Timeout = &value
	}
	if route.IdleTimeout.Valid {
		value := route.IdleTimeout.Int64
		routeDetails.IdleTimeout = &value
	}
	if route.HostAutoRewrite.Valid {
		value := route.HostAutoRewrite.Bool
		routeDetails.HostAutoRewrite = &value
	}
	for _, headerToAdd := range route.RequestHeadersToAdd {
		routeDetails.RequestHeadersToAdd = append(routeDetails.RequestHeadersToAdd, HeaderToAdd{
			Name:  headerToAdd.Name,
			Value: headerToAdd.Value,
		})
	}
	for _, headerMatcher := range route.HeaderMatchers {
		routeDetails.HeaderMatchers = append(routeDetails.HeaderMatchers, HeaderMatcher{
			Name:           headerMatcher.Name,
			ExactMatch:     headerMatcher.ExactMatch,
			SafeRegexMatch: headerMatcher.SafeRegexMatch,
			RangeMatch: RangeMatch{
				Start: headerMatcher.RangeMatch.Start,
				End:   headerMatcher.RangeMatch.End,
			},
			PresentMatch: headerMatcher.PresentMatch,
			PrefixMatch:  headerMatcher.PrefixMatch,
			SuffixMatch:  headerMatcher.SuffixMatch,
			InvertMatch:  headerMatcher.InvertMatch,
		})

	}
	if routeDetails.Endpoint, err = s.resolveEndpoint(route); err != nil {
		return RouteDetails{}, errors.Wrap(err, 1)
	}

	return routeDetails, nil
}

func (s *V3Service) resolveEndpoint(route *domain.Route) (string, error) {
	dVersion, err := s.ds.FindDeploymentVersion(route.DeploymentVersion)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}
	cluster, err := s.ds.FindClusterByName(route.ClusterName)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}
	endpoints, err := s.ds.FindEndpointsByClusterIdAndDeploymentVersion(cluster.Id, dVersion)
	if err != nil {
		return "", errors.Wrap(err, 1)
	}
	if len(endpoints) != 1 {
		return "", errors.New(fmt.Sprintf("there is the only one endpoint by version and clusterId but got '%d'", len(endpoints)))
	}
	endpoint := fmt.Sprintf("http://%s:%d", endpoints[0].Address, endpoints[0].Port)
	return endpoint, err
}

func filterRoutes(routes []*domain.Route, test func(*domain.Route) bool) []*domain.Route {
	n := 0
	for _, route := range routes {
		if test(route) {
			routes[n] = route
			n++
		}
	}
	return routes[:n]
}

type entityServiceWrapper struct {
	service *entity.Service
	rep     dao.Repository
}

func (w *entityServiceWrapper) GetClustersWithRelations() ([]*domain.Cluster, error) {
	return w.service.GetClustersWithRelations(w.rep)
}
