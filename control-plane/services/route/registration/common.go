package registration

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"os"
	"strconv"
)

var (
	logger logging.Logger
)

func init() {
	logger = logging.GetLogger("common-route-builder")
}

type ProcessedRequest struct {
	NodeGroups          []domain.NodeGroup
	Listeners           []domain.Listener
	Clusters            []domain.Cluster
	ClusterNodeGroups   map[string][]string // <clusterName>: []<nodeGroupIds>
	RouteConfigurations []domain.RouteConfiguration
	GroupedRoutes       *groupedRoutesMap
	DeploymentVersions  []string
	ClusterTlsConfig    map[string]string
	RouteRateLimit      map[string]string
}

func (pr ProcessedRequest) String() string {
	return fmt.Sprintf("ProcessedRequest{nodeGroups=%v,listeners=%v,clusters=%v,clusterNodeGroups=%v,routeConfigurations=%v,versions=%v",
		pr.NodeGroups, pr.Listeners, pr.Clusters, pr.ClusterNodeGroups, pr.RouteConfigurations, pr.DeploymentVersions)
}

// groupedRoutesMap holds routes mapped by namespace -> clusterKey -> deploymentVersion,
// which is very useful for routes auto generation.
type groupedRoutesMap struct {
	// storage holds routes mapped by namespace -> clusterKey -> deploymentVersion
	storage map[string]map[string]map[string][]*domain.Route
}

func ResolveVersions(dao dao.Repository, clusterName, requestVersion, activeVersion string) (string, string, error) {
	if requestVersion == "" {
		requestVersion = activeVersion
	}

	if requestVersion != "" && activeVersion != "" && requestVersion != activeVersion {
		requestVersionNum, _ := util.GetVersionNumber(requestVersion)
		activeVersionNum, _ := util.GetVersionNumber(activeVersion)
		if requestVersionNum > activeVersionNum {
			return requestVersion, requestVersion, nil
		}
	}

	foundEndpoints, err := dao.FindEndpointsByClusterName(clusterName)
	if err != nil {
		logger.Errorf("Failed to find by cluster name %v using DAO: %v", clusterName, err)
		return "", "", err
	}
	logger.Infof("Found '%d' endpoints for cluster '%s'", len(foundEndpoints), clusterName)

	for _, existingEndpoint := range foundEndpoints {
		if requestVersion == existingEndpoint.InitialDeploymentVersion {
			return existingEndpoint.InitialDeploymentVersion, existingEndpoint.DeploymentVersion, nil
		}
	}

	for _, existingEndpoint := range foundEndpoints {
		if requestVersion == existingEndpoint.DeploymentVersion {
			return existingEndpoint.InitialDeploymentVersion, existingEndpoint.DeploymentVersion, nil
		}
	}

	return requestVersion, requestVersion, nil
}

func NewGroupedRoutesMap() *groupedRoutesMap {
	return &groupedRoutesMap{storage: make(map[string]map[string]map[string][]*domain.Route)}
}

func (m *groupedRoutesMap) ForEachGroup(function func(namespace, clusterKey, deploymentVersion string, routes []*domain.Route) error) error {
	for namespace, namespaceRoutes := range m.storage {
		for clusterKey, clusterRoutes := range namespaceRoutes {
			for deploymentVersion, routes := range clusterRoutes {
				if err := function(namespace, clusterKey, deploymentVersion, routes); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *groupedRoutesMap) PutRoute(namespace, clusterKey, deploymentVersion string, route *domain.Route) {
	if namespace == "" {
		namespace = msaddr.DefaultNamespace
	}
	if deploymentVersion == "" {
		deploymentVersion = "default"
	}
	if _, exists := m.storage[namespace]; !exists {
		m.storage[namespace] = make(map[string]map[string][]*domain.Route)
	}
	if _, exists := m.storage[namespace][clusterKey]; !exists {
		m.storage[namespace][clusterKey] = make(map[string][]*domain.Route)
	}
	if _, exists := m.storage[namespace][clusterKey][deploymentVersion]; !exists {
		m.storage[namespace][clusterKey][deploymentVersion] = make([]*domain.Route, 0)
	}
	m.storage[namespace][clusterKey][deploymentVersion] = append(m.storage[namespace][clusterKey][deploymentVersion], route)
}

type CommonEntityBuilder struct {
	NodeGroupName string
}

func NewCommonEntityBuilder(nodeGroupName string) CommonEntityBuilder {
	return CommonEntityBuilder{NodeGroupName: nodeGroupName}
}

func (d CommonEntityBuilder) CreateNodeGroup() domain.NodeGroup {
	return domain.NodeGroup{
		Name: d.NodeGroupName,
	}
}

func (d CommonEntityBuilder) CreateListener() domain.Listener {
	return d.createListenerV2(false, util.DefaultProtocol, util.DefaultPort)
}

func (d CommonEntityBuilder) CreateListenerWithCustomPort(port int, tlsSupported bool) domain.Listener {
	if port == 0 {
		port = util.DefaultPort
	}
	proto := d.getProto(tlsSupported)
	return d.createListenerV2(false, proto, port)
}

func (d CommonEntityBuilder) getProto(tlsSupported bool) string {
	if tlsSupported && tlsmode.GetMode() == tlsmode.Preferred {
		return util.TlsProtocol
	}

	return util.DefaultProtocol
}

func (d CommonEntityBuilder) createListenerV2(loopback bool, proto string, port int) domain.Listener {
	ipVersion := os.Getenv("IP_STACK")
	if ipVersion == "" {
		ipVersion = "v4"
	}
	var binder string
	if ipVersion == "v4" {
		if loopback {
			binder = "127.0.0.1"
		} else {
			binder = "0.0.0.0"
		}
	} else {
		if loopback {
			binder = "::1"
		} else {
			binder = "::"
		}
	}

	name := d.NodeGroupName + "-listener"
	if port != util.DefaultPort && port != util.TlsPort {
		name += "-" + strconv.Itoa(port)
	}

	withTls := false
	if proto == util.TlsProtocol {
		withTls = true
	}

	return domain.Listener{
		Name:                   name,
		BindHost:               binder,
		BindPort:               strconv.Itoa(port),
		RouteConfigurationName: d.NodeGroupName + "-routes",
		Version:                1,
		NodeGroupId:            d.NodeGroupName,
		WithTls:                withTls,
	}
}

func (d CommonEntityBuilder) CreateRouteConfiguration() domain.RouteConfiguration {
	return domain.RouteConfiguration{
		Name:         d.NodeGroupName + "-routes",
		NodeGroupId:  d.NodeGroupName,
		Version:      1,
		VirtualHosts: make([]*domain.VirtualHost, 0),
	}
}

func (d CommonEntityBuilder) CreateVirtualHost() domain.VirtualHost {
	return domain.VirtualHost{Name: d.NodeGroupName, Version: 1, Routes: make([]*domain.Route, 0)}
}

func (d CommonEntityBuilder) CreateVirtualHostDomain() domain.VirtualHostDomain {
	return domain.VirtualHostDomain{Domain: "*", Version: 1}
}

func resolveEndpoint(cluster *domain.Cluster, msAddress *msaddr.MicroserviceAddress, deploymentVersion string) domain.Endpoint {
	endpointAddress := msAddress.GetNamespacedMicroserviceHost()
	endpointPort := msAddress.GetPort()
	for _, existingEndpoint := range cluster.Endpoints {
		if existingEndpoint.Address == endpointAddress && existingEndpoint.Port == endpointPort &&
			existingEndpoint.DeploymentVersion == deploymentVersion {
			return *existingEndpoint
		}
	}
	endpoint := domain.Endpoint{
		Address:                  msAddress.GetNamespacedMicroserviceHost(),
		Port:                     msAddress.GetPort(),
		Protocol:                 msAddress.GetProto(),
		DeploymentVersion:        deploymentVersion,
		InitialDeploymentVersion: deploymentVersion,
	}
	cluster.Endpoints = append(cluster.Endpoints, &endpoint)
	return endpoint
}
