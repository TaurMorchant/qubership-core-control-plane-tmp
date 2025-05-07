package debug

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/composite"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"strconv"
	"strings"
)

var logger = logging.GetLogger("config-validation")

type Severity int

const (
	Critical Severity = iota + 1
	Major
	Warning
)

func (s Severity) String() string {
	return [...]string{"critical", "major", "warning"}[s-1]
}

func (s Severity) EnumIndex() int {
	return int(s)
}

type ProblemType int

const (
	VHostsConflict ProblemType = iota + 1
	ClusterDuplicate
	Bgd1Cluster
	OrphanedCluster
	PrefixSlash
	Loop
)

func (p ProblemType) String() string {
	return [...]string{"vHostsConflict", "clusterDuplicate", "bgd1Cluster", "orphanedCluster", "prefixSlashInBorderGW", "loop"}[p-1]
}

func (p ProblemType) getMessage() string {
	return [...]string{"there are two virtual services with conflicting hosts in gateway",
		"there are duplicated clusters with the same target addresses",
		"there are legacy blue-green clusters with several bg versions",
		"there are orphaned clusters that are not referenced by any route",
		"custom route with prefix '/' in public, private or internal gateway",
		"there are loop route redirects back to itself (endpoint=host)"}[p-1]
}

func (p ProblemType) EnumIndex() int {
	return int(p)
}

type Detail struct {
	Gateway         string            `json:"gateway,omitempty"`
	VirtualServices []*VirtualService `json:"virtualServices,omitempty"`
	Clusters        []*Cluster        `json:"clusters,omitempty"`
}

type VirtualService struct {
	Name  string   `json:"name"`
	Hosts []string `json:"hosts"`
}

type Cluster struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}
type Problem struct {
	ProblemType string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Details     []*Detail `json:"details"`
}

type StatusConfig struct {
	Status   string     `json:"status"`
	Problems []*Problem `json:"problems,omitempty"`
}

func ValidateConfig(dao dao.Repository, compositeService *composite.Service) (*StatusConfig, error) {
	problems := make([]*Problem, 0)
	//there are two virtual services with conflicting hosts in gateway
	problemVirtualHostsConflict, err := validateVirtualHostsConflict(dao)
	if err != nil {
		logger.Errorf("Failed to validate virtual hosts: %v", err)
		return nil, err
	}
	//there are duplicated clusters with the same target addresses (address + port)
	problemClustersEndpoints, err := validateClustersDuplicate(dao)
	if err != nil {
		logger.Errorf("Failed to validate clusters endpoints: %v", err)
		return nil, err
	}
	//there are legacy blue-green clusters with several bg versions
	problemClustersBGD1, err := validateBGD1Cluster(dao)
	if err != nil {
		logger.Errorf("Failed to validate bgd1 clusters: %v", err)
		return nil, err
	}
	//there are orphaned clusters that are not referenced by any route
	problemOrphanedCluster, err := validateOrphanedClusters(dao)
	if err != nil {
		logger.Errorf("Failed to validate orphaned clusters: %v", err)
		return nil, err
	}
	//Routes with prefix "/" in public, private and internal gateways, that do not lead to the same gateway in Composite Deployment baseline.
	problemSlashPrefix, err := validateSlashPrefix(dao, compositeService)
	if err != nil {
		logger.Errorf("Failed to validate slash prefix: %v", err)
		return nil, err
	}
	// Loop detection
	problemLoop, err := validateLoopDetection(dao)
	if err != nil {
		logger.Errorf("Failed to validate loop: %v", err)
		return nil, err
	}

	for _, problem := range []*Problem{problemVirtualHostsConflict, problemClustersEndpoints, problemClustersBGD1, problemOrphanedCluster, problemSlashPrefix, problemLoop} {
		if problem != nil {
			problems = append(problems, problem)
		}
	}

	if len(problems) == 0 {
		return &StatusConfig{Status: "ok"}, nil
	}

	return &StatusConfig{Status: "problem", Problems: problems}, nil
}

func validateVirtualHostsConflict(dao dao.Repository) (*Problem, error) {
	nodeGroups, err := dao.FindAllNodeGroups()
	if err != nil {
		logger.Errorf("Failed to load all node groups using DAO: %v", err)
		return nil, err
	}

	details := make([]*Detail, 0)

	for _, nodeGroup := range nodeGroups {
		// Get route configs from gateway(node group)
		routeConfigs, err := dao.FindRouteConfigsByNodeGroupId(nodeGroup.Name)
		if err != nil {
			logger.Errorf("Failed to find route configs by node group id using DAO: %v", err)
			return nil, err
		}

		virtualHosts := make([]*domain.VirtualHost, 0)

		for _, routeConfig := range routeConfigs {
			virtualHostsI, err := dao.FindVirtualHostsByRouteConfigurationId(routeConfig.Id)
			if err != nil {
				logger.Errorf("Failed to find virtual hosts by route configuration id using DAO: %v", err)
				return nil, err
			}
			virtualHosts = append(virtualHosts, virtualHostsI...)
		}

		virtualServicesWithStar := make([]*VirtualService, 0)
		virtualServicesWithStarMoreThanOneDomain := make([]*VirtualService, 0)
		for _, virtualHost := range virtualHosts {
			domains, err := dao.FindVirtualHostDomainByVirtualHostId(virtualHost.Id)
			if err != nil {
				logger.Errorf("Failed to find virtual host domains by virtual host id using DAO: %v", err)
				return nil, err
			}
			isThereStar, moreThanOneDomain := validateDomains(domains)
			if isThereStar && moreThanOneDomain {
				virtualService := &VirtualService{
					Name:  virtualHost.Name,
					Hosts: virtualServicesDomainsToStringArray(domains),
				}
				virtualServicesWithStarMoreThanOneDomain = append(virtualServicesWithStarMoreThanOneDomain, virtualService)
			}
			if isThereStar && !moreThanOneDomain {
				virtualService := &VirtualService{
					Name:  virtualHost.Name,
					Hosts: virtualServicesDomainsToStringArray(domains),
				}
				virtualServicesWithStar = append(virtualServicesWithStar, virtualService)
			}
		}

		if (len(virtualHosts) > 1 && len(virtualServicesWithStar) > 0) || len(virtualServicesWithStarMoreThanOneDomain) > 0 {
			toSlice := &Detail{
				Gateway:         nodeGroup.Name,
				VirtualServices: append(virtualServicesWithStar, virtualServicesWithStarMoreThanOneDomain...),
				Clusters:        nil,
			}
			details = append(details, toSlice)
		}
	}
	if len(details) > 0 {
		problem := &Problem{
			ProblemType: VHostsConflict.String(),
			Severity:    Critical.String(),
			Message:     VHostsConflict.getMessage(),
			Details:     details,
		}
		return problem, nil
	}
	return nil, nil
}

func virtualServicesDomainsToStringArray(domains []*domain.VirtualHostDomain) []string {
	stringHosts := make([]string, len(domains))
	for i, domain := range domains {
		stringHosts[i] = domain.Domain
	}
	return stringHosts
}

// return bool (there is "*" in domain), bool (there are several domains/hosts)
func validateDomains(domains []*domain.VirtualHostDomain) (bool, bool) {
	var star, several bool
	if len(domains) > 1 {
		several = true
	}
	for _, domain := range domains {
		if domain.Domain == "*" {
			star = true
		}
	}
	return star, several
}

// there are duplicated clusters with the same target addresses(address + port)
func validateClustersDuplicate(dao dao.Repository) (*Problem, error) {
	clusters, err := dao.FindAllClusters()
	if err != nil {
		logger.Errorf("Failed to load all clusters using DAO: %v", err)
		return nil, err
	}

	details := make([]*Detail, 0)
	clustersEndpoints := make(map[string][]*domain.Endpoint)

	if len(clusters) < 2 {
		return nil, nil
	}

	for index, cluster := range clusters[:len(clusters)-1] {
		for _, clusterNext := range clusters[index+1:] {
			endpoints, err := getEndpoints(clustersEndpoints, cluster.Name, dao)
			if err != nil {
				logger.Errorf("Failed to get endpoints for cluster: %v", err)
				return nil, err
			}
			for _, endpoint := range endpoints {
				endpointsNext, err := getEndpoints(clustersEndpoints, clusterNext.Name, dao)
				if err != nil {
					logger.Errorf("Failed to get endpoints for cluster: %v", err)
					return nil, err
				}
				for _, endpointClusterNext := range endpointsNext {
					if endpoint.Address == endpointClusterNext.Address && endpoint.Port == endpointClusterNext.Port {
						detail := &Detail{Clusters: []*Cluster{
							{
								Name:      cluster.Name,
								Endpoints: endpointsToStringArray(endpoints),
							},
							{
								Name:      clusterNext.Name,
								Endpoints: endpointsToStringArray(endpointsNext),
							},
						},
						}
						details = append(details, detail)
					}
				}
			}
		}
	}

	if len(details) > 0 {
		problem := &Problem{
			ProblemType: ClusterDuplicate.String(),
			Severity:    Major.String(),
			Message:     ClusterDuplicate.getMessage(),
			Details:     details,
		}
		return problem, nil
	}
	return nil, nil
}

func getEndpoints(clustersEndpoints map[string][]*domain.Endpoint, clusterName string, dao dao.Repository) ([]*domain.Endpoint, error) {
	endpoints, ok := clustersEndpoints[clusterName]
	if !ok {
		var err error
		endpoints, err = dao.FindEndpointsByClusterName(clusterName)
		if err != nil {
			logger.Errorf("Failed to find endpoints by cluster name using DAO: %v", err)
			return nil, err
		}
		clustersEndpoints[clusterName] = endpoints
	}
	return endpoints, nil
}

func endpointsToStringArray(endpoints []*domain.Endpoint) []string {
	endpointsString := make([]string, 0)
	for _, endpoint := range endpoints {
		endpointString := endpoint.Address + ":" + strconv.Itoa(int(endpoint.Port))
		endpointsString = append(endpointsString, endpointString)
	}
	return endpointsString
}

// there are legacy blue-green clusters with several bg versions
func validateBGD1Cluster(dao dao.Repository) (*Problem, error) {
	clusters, err := dao.FindAllClusters()
	if err != nil {
		logger.Errorf("Failed to load all clusters using DAO: %v", err)
		return nil, err
	}

	badClusters := make([]*Cluster, 0)

	for _, cluster := range clusters {
		endpoints, err := dao.FindEndpointsByClusterName(cluster.Name)
		if err != nil {
			logger.Errorf("Failed to find endpoints by cluster name using DAO: %v", err)
			return nil, err
		}
		if len(endpoints) > 1 {
			badCluster := &Cluster{
				Name:      cluster.Name,
				Endpoints: endpointsToStringArray(endpoints),
			}
			badClusters = append(badClusters, badCluster)
		}
	}

	if len(badClusters) > 0 {
		detail := &Detail{Clusters: badClusters}
		problem := &Problem{
			ProblemType: Bgd1Cluster.String(),
			Severity:    Major.String(),
			Message:     Bgd1Cluster.getMessage(),
			Details:     []*Detail{detail},
		}
		return problem, nil
	}
	return nil, nil
}

func validateOrphanedClusters(dao dao.Repository) (*Problem, error) {
	clusters, err := dao.FindAllClusters()
	if err != nil {
		logger.Errorf("Failed to load all clusters using DAO:\n %v", err)
		return nil, err
	}
	routes, err := dao.FindAllRoutes()
	if err != nil {
		logger.Errorf("Failed to load all routes using DAO: %v", err)
		return nil, err
	}

	orphanedClusters := make([]*Cluster, 0)

	for _, cluster := range clusters {
		if cluster.Name == "ext-authz" {
			continue
		}
		if isClusterOrphaned(cluster.Name, routes) {
			endpoints, err := dao.FindEndpointsByClusterName(cluster.Name)
			if err != nil {
				logger.Errorf("Failed to find endpoints by cluster name using DAO: %v", err)
				return nil, err
			}
			orphanedCluster := &Cluster{
				Name:      cluster.Name,
				Endpoints: endpointsToStringArray(endpoints),
			}
			orphanedClusters = append(orphanedClusters, orphanedCluster)
		}
	}

	if len(orphanedClusters) > 0 {
		detail := &Detail{Clusters: orphanedClusters}
		problem := &Problem{
			ProblemType: OrphanedCluster.String(),
			Severity:    Warning.String(),
			Message:     OrphanedCluster.getMessage(),
			Details:     []*Detail{detail},
		}
		return problem, nil
	}
	return nil, nil
}

func isClusterOrphaned(clusterName string, routes []*domain.Route) bool {
	for _, route := range routes {
		if clusterName == route.ClusterName {
			return false
		}
	}
	return true
}

// Routes with prefix "/" in public, private and internal gateways, that do not lead to the same gateway in Composite Deployment baseline.
func validateSlashPrefix(dao dao.Repository, compositeService *composite.Service) (*Problem, error) {
	details := make([]*Detail, 0)

	virtualHosts, err := dao.FindAllVirtualHosts()
	if err != nil {
		logger.Errorf("Failed to find virtual hosts by route configuration id using DAO: %v", err)
		return nil, err
	}

	for _, gateway := range []string{"private-gateway-service", "public-gateway-service", "internal-gateway-service"} {
		badClustersMap := make(map[string]*Cluster, 0)

		for _, virtualHost := range virtualHosts {
			if virtualHost.Name == gateway {
				routes, err := dao.FindRoutesByVirtualHostId(virtualHost.Id)
				if err != nil {
					logger.Errorf("Failed to find routes by virtual host id using DAO: %v", err)
					return nil, err
				}
				for _, route := range routes {
					if route.Prefix == "/" {
						endpoints, err := dao.FindEndpointsByClusterName(route.ClusterName)
						if err != nil {
							logger.Errorf("Failed to find endpoints by cluster name using DAO: %v", err)
							return nil, err
						}
						if compositeService.Mode() == composite.SatelliteMode {
							compositeStructure, _ := compositeService.GetCompositeStructure()
							for _, endpoint := range endpoints {
								if strings.Contains(endpoint.Address, compositeStructure.Baseline) {
									continue
								}
							}
						}
						_, ok := badClustersMap[route.ClusterName]
						if !ok {
							cluster := &Cluster{
								Name:      route.ClusterName,
								Endpoints: endpointsToStringArray(endpoints),
							}
							badClustersMap[route.ClusterName] = cluster
						}
					}
				}
			}
		}

		if len(badClustersMap) > 0 {
			badClusters := make([]*Cluster, 0, len(badClustersMap))
			for _, badClusterFromMap := range badClustersMap {
				badClusters = append(badClusters, badClusterFromMap)
			}
			detail := &Detail{
				Gateway:  gateway,
				Clusters: badClusters,
			}
			details = append(details, detail)
		}
	}

	if len(details) > 0 {
		problem := &Problem{
			ProblemType: PrefixSlash.String(),
			Severity:    Critical.String(),
			Message:     PrefixSlash.getMessage(),
			Details:     details,
		}
		return problem, nil
	}
	return nil, nil
}

// There are loop route redirects back to itself (hostRewrite=Domain)
func validateLoopDetection(dao dao.Repository) (*Problem, error) {
	routeConfigs, err := dao.FindAllRouteConfigs()
	if err != nil {
		logger.Errorf("Failed to load all route configs using DAO: %v", err)
		return nil, err
	}

	detailsMap := make(map[string]*Detail)

	for _, routeConfig := range routeConfigs {
		virtualHosts, err := dao.FindVirtualHostsByRouteConfigurationId(routeConfig.Id)
		if err != nil {
			logger.Errorf("Failed to find virtual hosts by route configuration id using DAO: %v", err)
			return nil, err
		}
		for _, virtualHost := range virtualHosts {
			domains, err := dao.FindVirtualHostDomainByVirtualHostId(virtualHost.Id)
			if err != nil {
				logger.Errorf("Failed to find virtual host domains by virtual host id using DAO: %v", err)
				return nil, err
			}
			for _, domainI := range domains {
				routes, err := dao.FindRoutesByVirtualHostId(virtualHost.Id)
				if err != nil {
					logger.Errorf("Failed to find routes by virtual host id using DAO: %v", err)
					return nil, err
				}
				for _, route := range routes {
					endpoints, err := dao.FindEndpointsByClusterName(route.ClusterName)
					if err != nil {
						logger.Errorf("Failed to find endpoints by cluster name using DAO: %v", err)
						return nil, err
					}
					for _, endpoint := range endpoints {
						if endpoint.Address+":"+strconv.Itoa(int(endpoint.Port)) == domainI.Domain {
							key := virtualHost.Name + "+" + route.ClusterName
							_, ok := detailsMap[key]
							if !ok {
								vS := &VirtualService{
									Name:  virtualHost.Name,
									Hosts: virtualServicesDomainsToStringArray(domains),
								}
								cl := &Cluster{
									Name:      route.ClusterName,
									Endpoints: endpointsToStringArray(endpoints),
								}

								detail := &Detail{
									Gateway:         routeConfig.NodeGroupId,
									Clusters:        []*Cluster{cl},
									VirtualServices: []*VirtualService{vS},
								}
								detailsMap[key] = detail
							}
							continue
						}
					}
				}
			}

		}
	}

	if len(detailsMap) > 0 {
		details := make([]*Detail, 0, len(detailsMap))
		for _, detail := range detailsMap {
			details = append(details, detail)
		}

		problem := &Problem{
			ProblemType: Loop.String(),
			Severity:    Critical.String(),
			Message:     Loop.getMessage(),
			Details:     details,
		}
		return problem, nil
	}
	return nil, nil
}
