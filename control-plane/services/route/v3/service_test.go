package v3

import (
	"fmt"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/route"
	"github.com/netcracker/qubership-core-control-plane/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/services/route/registration"
	"github.com/netcracker/qubership-core-control-plane/services/routingmode"
	"github.com/stretchr/testify/assert"
	"sort"
	"sync/atomic"
	"testing"
)

// registration

func TestRegisterRoutesV3(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func verifyVirtualHostsWithConcreteDomain(t *testing.T, inMemDao *dao.InMemDao, expectedVirtualHostsNum int, vhosts ...*domain.VirtualHost) {
	assert.NotEmpty(t, vhosts)
	assert.Equal(t, expectedVirtualHostsNum, len(vhosts))

	for i := 1; i <= expectedVirtualHostsNum; i++ {
		expectedServiceName := "test-vs"
		if expectedVirtualHostsNum > 1 {
			expectedServiceName = fmt.Sprintf("%s%v", expectedServiceName, i)
		}
		srvIsPresent := false
		for _, vhost := range vhosts {
			if vhost.Name == expectedServiceName {
				srvIsPresent = true
				expectedDomains := []*domain.VirtualHostDomain{
					{Domain: fmt.Sprintf("www.test1-vs%d.com", i), VirtualHostId: vhost.Id},
					{Domain: fmt.Sprintf("www.test2-vs%d.com", i), VirtualHostId: vhost.Id},
				}
				actualDomains, err := inMemDao.FindVirtualHostDomainByVirtualHostId(vhost.Id)
				assert.Nil(t, err)
				assertContainsDomains(t, expectedDomains, actualDomains)
				break
			}
		}
		assert.True(t, srvIsPresent)
	}
}

func verifyVirtualHostsWithGoogleDomain(t *testing.T, inMemDao *dao.InMemDao, expectedVirtualHostsNum int, vhosts ...*domain.VirtualHost) {
	assert.NotEmpty(t, vhosts)
	assert.Equal(t, expectedVirtualHostsNum, len(vhosts))

	for i := 1; i <= expectedVirtualHostsNum; i++ {
		expectedServiceName := "test-vs"
		if expectedVirtualHostsNum > 1 {
			expectedServiceName = fmt.Sprintf("%s%v", expectedServiceName, i)
		}
		srvIsPresent := false
		for _, vhost := range vhosts {
			if vhost.Name == expectedServiceName {
				srvIsPresent = true
				expectedDomains := []*domain.VirtualHostDomain{{Domain: "www.google.com", VirtualHostId: vhost.Id}}
				actualDomains, err := inMemDao.FindVirtualHostDomainByVirtualHostId(vhost.Id)
				assert.Nil(t, err)
				assertContainsDomains(t, expectedDomains, actualDomains)
				break
			}
		}
		assert.True(t, srvIsPresent)
	}
}

func verifyVirtualHostsWithAnyDomain(t *testing.T, inMemDao *dao.InMemDao, expectedVirtualHostsNum int, vhosts ...*domain.VirtualHost) {
	assert.NotEmpty(t, vhosts)
	assert.Equal(t, expectedVirtualHostsNum, len(vhosts))

	for i := 1; i <= expectedVirtualHostsNum; i++ {
		expectedServiceName := "test-vs"
		if expectedVirtualHostsNum > 1 {
			expectedServiceName = fmt.Sprintf("%s%v", expectedServiceName, i)
		}
		srvIsPresent := false
		for _, vhost := range vhosts {
			if vhost.Name == expectedServiceName {
				srvIsPresent = true
				expectedDomains := []*domain.VirtualHostDomain{{Domain: "*", VirtualHostId: vhost.Id}}
				actualDomains, err := inMemDao.FindVirtualHostDomainByVirtualHostId(vhost.Id)
				assert.Nil(t, err)
				assertContainsDomains(t, expectedDomains, actualDomains)
				break
			}
		}
		assert.True(t, srvIsPresent)
	}
}

func assertContainsDomains(t *testing.T, expected []*domain.VirtualHostDomain, actual []*domain.VirtualHostDomain) {
	for _, expectedDomain := range expected {
		found := false
		for _, actualDomain := range actual {
			if found = actualDomain.Equals(expectedDomain); found {
				break
			}
		}
		if !found {
			t.Fatalf("Expected virtual host domain %+v not present in %v", expectedDomain, actual)
		}
	}
}

func assertNotContainsDomains(t *testing.T, absent []*domain.VirtualHostDomain, actual []*domain.VirtualHostDomain) {
	for _, absentDomain := range absent {
		found := false
		for _, actualDomain := range actual {
			if found = actualDomain.Equals(absentDomain); found {
				break
			}
		}
		if found {
			t.Fatalf("Absent virtual host domain %+v is present in %v", absentDomain, actual)
		}
	}
}

func TestRegisterRoutesV3_WithDefaultDomains(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithEmptyDomains()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithIncorrectRoute(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestgetV3ServiceWithIncorrectRoute()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithManyVirtualServicesWithSameRoutes(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithSameRoutes()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithManyVirtualServicesWithSameRoutesWithDifferentVersions(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithSameRoutesWithDifferentVersions()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	dVersion, err := inMemDao.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersion)
	assert.Equal(t, 2, len(dVersion))

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 2, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithManyVirtualServicesWithDifferentRoutes(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentRoutes()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithManyVirtualServicesWithDifferentRoutesWithHeaders(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentRoutesWithHeaders()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))
	for _, route := range routes {
		assert.Equal(t, 2, len(route.RequestHeadersToAdd))
		assert.Equal(t, 2, len(route.RequestHeadersToRemove))
	}

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithManyVirtualServicesWithDifferentClustersAndEndpoints(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpoints()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 6, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 3, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 3, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_WithDifferentClustersAndEndpointsWithRouteHeaders(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 6, len(routes))
	for _, route := range routes {
		headerMatchers, err := inMemDao.FindHeaderMatcherByRouteId(route.Id)
		assert.Nil(t, err)
		assert.NotEmpty(t, headerMatchers)
		assert.Equal(t, 1, len(headerMatchers))
	}

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 3, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 3, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 8, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestRegisterRoutesV3_RegisterRoutesConfig_MetadataNamespaceField(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	configresources.RegisterResource(v3Service.GetRoutingRequestResource())

	config := configresources.ConfigResource{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "RouteConfiguration",
		Metadata:   map[string]interface{}{},
		Spec:       createV3Request(false),
	}
	_, cpErr := configresources.HandleConfigResource(nil, config)
	assert.Nil(t, cpErr)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	config.Metadata["namespace"] = ""
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.Nil(t, cpErr)

	config.Metadata["namespace"] = "   "
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.Nil(t, cpErr)

	config.Metadata["namespace"] = 123
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["namespace"] = nil
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)
}

// GetVirtualService

func TestV3Service_GetVirtualService(t *testing.T) {
	v3Service, _ := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	vsGateway1, err := v3Service.GetVirtualService("gateway1", "test-vs")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway1.Clusters))
	assert.Equal(t, 1, len(vsGateway1.Clusters[0].Endpoints))
	assert.Equal(t, 1, len(vsGateway1.VirtualHost.Routes))
	assert.Equal(t, 0, len(vsGateway1.VirtualHost.Routes[0].RouteMatcher.HeaderMatchers))
	assert.LessOrEqual(t, 1, len(vsGateway1.VirtualHost.Domains))

	vsGateway2, err := v3Service.GetVirtualService("gateway2", "test-vs")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway2.Clusters))
	assert.Equal(t, 1, len(vsGateway2.Clusters[0].Endpoints))
	assert.Equal(t, 1, len(vsGateway2.VirtualHost.Routes))
	assert.Equal(t, 0, len(vsGateway2.VirtualHost.Routes[0].RouteMatcher.HeaderMatchers))
	assert.LessOrEqual(t, 1, len(vsGateway2.VirtualHost.Domains))
}

func TestV3Service_GetVirtualServiceNotFound(t *testing.T) {
	v3Service, _ := getV3Service()
	vs, err := v3Service.GetVirtualService("gateway1", "test-vs")
	assert.NotNil(t, err)
	assert.Equal(t, dto.VirtualServiceResponse{}, vs)
}

func TestV3Service_GetVirtualServiceWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers(t *testing.T) {
	v3Service, _ := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	vsGateway1VirtualService1, err := v3Service.GetVirtualService("gateway1", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway1VirtualService1.Clusters))
	assert.Equal(t, 1, len(vsGateway1VirtualService1.Clusters[0].Endpoints))
	assert.Equal(t, 1, len(vsGateway1VirtualService1.VirtualHost.Routes))
	assert.Equal(t, 1, len(vsGateway1VirtualService1.VirtualHost.Routes[0].RouteMatcher.HeaderMatchers))
	assert.Equal(t, 0, len(vsGateway1VirtualService1.VirtualHost.Routes[0].RouteMatcher.RemoveHeaders))
	assert.Equal(t, 0, len(vsGateway1VirtualService1.VirtualHost.Routes[0].RouteMatcher.AddHeaders))
	assert.Equal(t, 2, len(vsGateway1VirtualService1.VirtualHost.Domains))
	assert.Equal(t, 2, len(vsGateway1VirtualService1.VirtualHost.AddHeaders))
	assert.Equal(t, 2, len(vsGateway1VirtualService1.VirtualHost.RemoveHeaders))

	vsGateway1VirtualService2, err := v3Service.GetVirtualService("gateway1", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway1VirtualService2.Clusters))
	assert.Equal(t, 1, len(vsGateway1VirtualService2.Clusters[0].Endpoints))
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.Routes))
	for _, route := range vsGateway1VirtualService2.VirtualHost.Routes {
		assert.Equal(t, 1, len(route.RouteMatcher.HeaderMatchers))
		assert.Equal(t, 0, len(route.RouteMatcher.RemoveHeaders))
		assert.Equal(t, 0, len(route.RouteMatcher.AddHeaders))
	}
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.Domains))
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.AddHeaders))
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.RemoveHeaders))

	vsGateway2VirtualService1, err := v3Service.GetVirtualService("gateway2", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway2VirtualService1.Clusters))
	assert.Equal(t, 1, len(vsGateway2VirtualService1.Clusters[0].Endpoints))
	assert.Equal(t, 1, len(vsGateway2VirtualService1.VirtualHost.Routes))
	assert.Equal(t, 1, len(vsGateway2VirtualService1.VirtualHost.Routes[0].RouteMatcher.HeaderMatchers))
	assert.Equal(t, 0, len(vsGateway2VirtualService1.VirtualHost.Routes[0].RouteMatcher.RemoveHeaders))
	assert.Equal(t, 0, len(vsGateway2VirtualService1.VirtualHost.Routes[0].RouteMatcher.AddHeaders))
	assert.Equal(t, 2, len(vsGateway2VirtualService1.VirtualHost.Domains))
	assert.Equal(t, 2, len(vsGateway2VirtualService1.VirtualHost.AddHeaders))
	assert.Equal(t, 2, len(vsGateway2VirtualService1.VirtualHost.RemoveHeaders))

	vsGateway2VirtualService2, err := v3Service.GetVirtualService("gateway2", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway2VirtualService2.Clusters))
	assert.Equal(t, 1, len(vsGateway2VirtualService2.Clusters[0].Endpoints))
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.Routes))
	for _, route := range vsGateway2VirtualService2.VirtualHost.Routes {
		assert.Equal(t, 1, len(route.RouteMatcher.HeaderMatchers))
		assert.Equal(t, 0, len(route.RouteMatcher.RemoveHeaders))
		assert.Equal(t, 0, len(route.RouteMatcher.AddHeaders))
	}
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.Domains))
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.AddHeaders))
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.RemoveHeaders))
}

// delete

func TestV3Service_DeleteRoutes(t *testing.T) {
	v3Service, _ := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routeToDeleteRequest := []dto.RouteDeleteRequestV3{
		{
			Gateways:       []string{"gateway1"},
			VirtualService: "test-vs1",
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Namespace: "",
				Version:   "v1",
				Routes: []dto.RouteDeleteItem{
					{
						Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
					},
				},
			},
		},
	}

	deleteRoutes, err := v3Service.DeleteRoutes(nil, routeToDeleteRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(deleteRoutes))

	vsGateway1VirtualService1, err := v3Service.GetVirtualService("gateway1", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(vsGateway1VirtualService1.VirtualHost.Routes))

	vsGateway1VirtualService2, err := v3Service.GetVirtualService("gateway1", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.Routes))

	vsGateway2VirtualService1, err := v3Service.GetVirtualService("gateway2", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway2VirtualService1.VirtualHost.Routes))

	vsGateway2VirtualService2, err := v3Service.GetVirtualService("gateway2", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.Routes))
}

func TestV3Service_DeleteRoutesWhichDoesNotExist(t *testing.T) {
	v3Service, _ := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routeToDeleteRequest := []dto.RouteDeleteRequestV3{
		{
			Gateways:       []string{"gateway1"},
			VirtualService: "test-vs1",
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Namespace: "",
				Version:   "v1",
				Routes: []dto.RouteDeleteItem{
					{
						Prefix: "/api/v3/tenant-manager/tenant",
					},
				},
			},
		},
	}
	deleteRoutes, err := v3Service.DeleteRoutes(nil, routeToDeleteRequest)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(deleteRoutes))

	vsGateway1VirtualService1, err := v3Service.GetVirtualService("gateway1", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway1VirtualService1.VirtualHost.Routes))

	vsGateway1VirtualService2, err := v3Service.GetVirtualService("gateway1", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.Routes))

	vsGateway2VirtualService1, err := v3Service.GetVirtualService("gateway2", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway2VirtualService1.VirtualHost.Routes))

	vsGateway2VirtualService2, err := v3Service.GetVirtualService("gateway2", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.Routes))
}

func TestV3Service_DeleteRoutesWithEmptyRoutesInRequest(t *testing.T) {
	v3Service, _ := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	routeToDeleteRequest := []dto.RouteDeleteRequestV3{
		{
			Gateways:       []string{"gateway1"},
			VirtualService: "test-vs1",
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Namespace: "",
				Version:   "v1",
				Routes:    []dto.RouteDeleteItem{},
			},
		},
	}
	deleteRoutes, err := v3Service.DeleteRoutes(nil, routeToDeleteRequest)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(deleteRoutes))

	vsGateway1VirtualService1, err := v3Service.GetVirtualService("gateway1", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(vsGateway1VirtualService1.VirtualHost.Routes))

	vsGateway1VirtualService2, err := v3Service.GetVirtualService("gateway1", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway1VirtualService2.VirtualHost.Routes))

	vsGateway2VirtualService1, err := v3Service.GetVirtualService("gateway2", "test-vs1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vsGateway2VirtualService1.VirtualHost.Routes))

	vsGateway2VirtualService2, err := v3Service.GetVirtualService("gateway2", "test-vs2")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(vsGateway2VirtualService2.VirtualHost.Routes))
}

func TestV3Service_DeleteVirtualService(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	err = v3Service.DeleteVirtualService(nil, "gateway1", "test-vs")
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 1, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh...)

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))

	listeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.NotEmpty(t, listeners)
	assert.Equal(t, 2, len(listeners))
}

func TestV3Service_DeleteVirtualServiceWithManyClusters(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpoints()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)

	err = v3Service.DeleteVirtualService(nil, "gateway1", "test-vs1")
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 5, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 3, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 3, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 3, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 6, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))

	listeners, err := inMemDao.FindAllListeners()
	assert.Nil(t, err)
	assert.NotEmpty(t, listeners)
	assert.Equal(t, 2, len(listeners))
}

func TestV3Service_DeleteVirtualServiceNotFound(t *testing.T) {
	v3Service, _ := getV3Service()
	err := v3Service.DeleteVirtualService(nil, "gateway1", "test-vs")
	assert.NotNil(t, err)
}

func TestV3Service_DeleteDomains(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithHosts("www.google.com", "www.test1.com")
	err := v3Service.RegisterRoutingConfig(nil, v3request)

	vHosts, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Len(t, vHosts, 2)

	actualDomains, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 4)
	for _, vHost := range vHosts {
		assertContainsDomains(t, vHost.Domains, actualDomains)
	}

	deletedDomains, err := v3Service.DeleteVirtualServiceDomains(nil,
		[]string{
			"www.google.com",
			"www.test1.com",
		},
		"gateway1",
		"test-vs")
	assert.Nil(t, err)
	assert.Len(t, deletedDomains, 2)

	actualDomains, err = inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 2)
	assertNotContainsDomains(t, vHosts[0].Domains, actualDomains)
	assertContainsDomains(t, vHosts[1].Domains, actualDomains)
}

func TestV3Service_DeleteNotExistingDomains(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithHosts("www.google.com", "www.test1.com")
	err := v3Service.RegisterRoutingConfig(nil, v3request)

	vHosts, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Len(t, vHosts, 2)

	actualDomains, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 4)
	for _, vHost := range vHosts {
		assertContainsDomains(t, vHost.Domains, actualDomains)
	}

	deletedDomains, err := v3Service.DeleteVirtualServiceDomains(nil,
		[]string{
			"www.test-absent-1.com",
			"www.test-absent-2.com",
		},
		"gateway1",
		"test-vs")
	assert.Nil(t, err)
	assert.Len(t, deletedDomains, 0)

	actualDomains, err = inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 4)
	assertContainsDomains(t, vHosts[0].Domains, actualDomains)
	assertContainsDomains(t, vHosts[1].Domains, actualDomains)
}

func TestV3Service_DeleteDomainsWithEmptyDomainsInRequest(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithHosts("www.google.com", "www.test1.com")
	err := v3Service.RegisterRoutingConfig(nil, v3request)

	vHosts, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Len(t, vHosts, 2)

	actualDomains, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 4)
	for _, vHost := range vHosts {
		assertContainsDomains(t, vHost.Domains, actualDomains)
	}

	deletedDomains, err := v3Service.DeleteVirtualServiceDomains(nil,
		[]string{},
		"gateway1",
		"test-vs")
	assert.Nil(t, err)
	assert.Len(t, deletedDomains, 0)

	actualDomains, err = inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.Len(t, actualDomains, 4)
	assertContainsDomains(t, vHosts[0].Domains, actualDomains)
	assertContainsDomains(t, vHosts[1].Domains, actualDomains)
}

func TestV3Service_RoutesDrop_MetadataNamespaceField(t *testing.T) {
	v3Service, _ := getV3Service()
	configresources.RegisterResource(v3Service.GetRoutesDropResource())

	v3request := createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers()
	v3request.Namespace = "test-namespace"
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)

	config := configresources.ConfigResource{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "RoutesDrop",
		Metadata: map[string]interface{}{
			"namespace": "test-namespace",
		},
		Spec: []dto.RouteDeleteRequestV3{
			{
				Gateways:       []string{"gateway1"},
				VirtualService: "test-vs1",
				RouteDeleteRequest: dto.RouteDeleteRequest{
					Version: "v1",
					Routes: []dto.RouteDeleteItem{
						{
							Prefix: "/api/v3/tenant-manager/tenant",
						},
					},
				},
			},
		},
	}
	_, err = configresources.HandleConfigResource(nil, config)
	assert.Nil(t, err)

	config.Metadata["namespace"] = ""
	_, err = configresources.HandleConfigResource(nil, config)
	assert.Nil(t, err)

	config.Metadata["namespace"] = "   "
	_, err = configresources.HandleConfigResource(nil, config)
	assert.Nil(t, err)

	config.Metadata["namespace"] = 123
	_, err = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, err)

	config.Metadata["namespace"] = nil
	_, err = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, err)
}

// update

func TestV3Service_UpdateVirtualServiceWithEmptyVirtualService(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	err = v3Service.UpdateVirtualService(nil, "gateway1", "test-vs", dto.VirtualService{})
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 0, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestV3Service_UpdateVirtualService(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	vs := dto.VirtualService{
		Hosts: []string{"www.google.com"},
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "Token",
				Value: "12345",
			},
		},
		RemoveHeaders: []string{"Authorization"},
	}
	err = v3Service.UpdateVirtualService(nil, "gateway1", "test-vs", vs)
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	sort.Slice(vh, func(i, j int) bool {
		return len(vh[i].RequestHeadersToAdd) < len(vh[j].RequestHeadersToAdd)
	})
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestV3Service_CreateWithDefaultVirtaulHostDomainsUpdateVirtualService(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithEmptyDomains()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	vs := dto.VirtualService{
		Hosts: []string{"www.google.com"},
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "Token",
				Value: "12345",
			},
		},
		RemoveHeaders: []string{"Authorization"},
	}
	err = v3Service.UpdateVirtualService(nil, "gateway1", "test-vs", vs)
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	sort.Slice(vh, func(i, j int) bool {
		return len(vh[i].RequestHeadersToAdd) < len(vh[j].RequestHeadersToAdd)
	})
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithAnyDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestV3Service_UpdateVirtualServiceWithAllValues(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3Request(false)
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	v := int64(100)
	idle := int64(200)
	vs := dto.VirtualService{
		Hosts: []string{"www.google.com"},
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "Token",
				Value: "12345",
			},
		},
		RemoveHeaders: []string{"Authorization"},
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							RemoveHeaders: []string{"Authorization"},
							AddHeaders: []dto.HeaderDefinition{
								{
									Name:  "Token",
									Value: "12345",
								},
							},
							Timeout:       &v,
							IdleTimeout:   &idle,
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
								HeaderMatchers: []dto.HeaderMatcher{
									{
										Name:       "header1",
										ExactMatch: "value1",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	err = v3Service.UpdateVirtualService(nil, "gateway1", "test-vs", vs)
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 3, len(routes))
	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i].RequestHeadersToAdd) < len(routes[j].RequestHeadersToAdd)
	})
	assert.Equal(t, 0, len(routes[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(routes[0].RequestHeadersToRemove))
	assert.Equal(t, 0, len(routes[0].HeaderMatchers))
	assert.Equal(t, 1, len(routes[1].RequestHeadersToAdd))
	assert.Equal(t, 1, len(routes[1].RequestHeadersToRemove))
	assert.Equal(t, 1, len(routes[1].HeaderMatchers))
	assert.Equal(t, int64(100), routes[1].Timeout.Int64)
	assert.Equal(t, int64(200), routes[1].IdleTimeout.Int64)

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	sort.Slice(vh, func(i, j int) bool {
		return len(vh[i].RequestHeadersToAdd) < len(vh[j].RequestHeadersToAdd)
	})
	assert.Equal(t, 2, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))
	assert.Equal(t, 1, len(vh[1].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[0])
	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh[1])

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

func TestV3Service_UpdateVirtualServiceWithUpdateAllRoutes(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createV3RequestWithManyVirtualServicesWithDifferentRoutesWithHeaders()
	err := v3Service.RegisterRoutingConfig(nil, v3request)
	assert.Nil(t, err)
	vs := dto.VirtualService{
		Hosts: []string{"www.google.com"},
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "Token1",
				Value: "123456",
			},
		},
		RemoveHeaders: []string{"Authorization1"},
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							RemoveHeaders: []string{"Authorization1"},
							AddHeaders: []dto.HeaderDefinition{
								{
									Name:  "Token1",
									Value: "123456",
								},
							},
						},
					},
				},
			},
		},
	}
	err = v3Service.UpdateVirtualService(nil, "gateway1", "test-vs1", vs)
	assert.Nil(t, err)

	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 4, len(routes))
	for index, _ := range routes {
		assert.Equal(t, 2, len(routes[index].RequestHeadersToAdd))
		assert.Equal(t, 2, len(routes[index].RequestHeadersToRemove))
	}

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 2, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 4, len(vhs))
	sort.Slice(vhs, func(i, j int) bool {
		return len(vhs[i].RequestHeadersToAdd) > len(vhs[j].RequestHeadersToAdd)
	})
	for index, _ := range vhs {
		if index == 0 {
			assert.Equal(t, 3, len(vhs[index].RequestHeadersToAdd))
			assert.Equal(t, 3, len(vhs[index].RequestHeadersToRemove))
		} else {
			assert.Equal(t, 2, len(vhs[index].RequestHeadersToAdd))
			assert.Equal(t, 2, len(vhs[index].RequestHeadersToRemove))
		}

	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 9, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 2, len(routeConfig))
}

//create virtual service

func TestV3Service_CreateVirtualService(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createVirtualService(false)
	err := v3Service.CreateVirtualService(nil, "test-node-group", v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 1, len(nodeGroups))

	vh, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vh)
	assert.Equal(t, 1, len(vh))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToAdd))
	assert.Equal(t, 0, len(vh[0].RequestHeadersToRemove))

	verifyVirtualHostsWithGoogleDomain(t, inMemDao, 1, vh...)

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 1, len(routeConfig))
}

func TestV3Service_CreateVirtualServiceWithManyVirtualServicesWithDifferentRoutesWithHeaders(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createVirtualServiceWithManyVirtualServicesWithDifferentRoutesWithHeaders()
	err := v3Service.CreateVirtualService(nil, "test-node-group", v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))
	for _, route := range routes {
		assert.Equal(t, 2, len(route.RequestHeadersToAdd))
		assert.Equal(t, 2, len(route.RequestHeadersToRemove))
	}

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 1, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 1, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 2, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 1, len(routeConfig))
}

func TestV3Service_CreateVirtualServiceTwice(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createVirtualService(false)
	err := v3Service.CreateVirtualService(nil, "test-node-group", v3request)
	assert.Nil(t, err)
	v3request = createVirtualServiceWithManyVirtualServicesWithDifferentRoutesWithHeaders()
	err = v3Service.CreateVirtualService(nil, "test-node-group", v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 2, len(routes))
	for _, route := range routes {
		assert.Equal(t, 2, len(route.RequestHeadersToAdd))
		assert.Equal(t, 2, len(route.RequestHeadersToRemove))
	}

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 1, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 1, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 1, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 1, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assertContainsDomains(t,
		[]*domain.VirtualHostDomain{
			{Domain: "www.google.com", VirtualHostId: vhs[0].Id},
			{Domain: "www.test1.com", VirtualHostId: vhs[0].Id},
			{Domain: "www.test2.com", VirtualHostId: vhs[0].Id}},
		vhd)

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 1, len(routeConfig))
}

func TestV3Service_CreateVirtualServiceWithManyVirtualServicesWithDifferentClustersAndEndpoints(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	v3request := createVirtualServiceWithManyVirtualServicesWithDifferentClustersAndEndpoints()
	err := v3Service.CreateVirtualService(nil, "test-node-group", v3request)
	assert.Nil(t, err)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 3, len(routes))

	endpoints, err := inMemDao.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, 3, len(endpoints))

	clusters, err := inMemDao.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, clusters)
	assert.Equal(t, 3, len(clusters))

	nodeGroups, err := inMemDao.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.NotEmpty(t, nodeGroups)
	assert.Equal(t, 1, len(nodeGroups))

	vhs, err := inMemDao.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhs)
	assert.Equal(t, 1, len(vhs))
	for _, vh := range vhs {
		assert.Equal(t, 2, len(vh.RequestHeadersToAdd))
		assert.Equal(t, 2, len(vh.RequestHeadersToRemove))
	}

	vhd, err := inMemDao.FindAllVirtualHostsDomain()
	assert.Nil(t, err)
	assert.NotEmpty(t, vhd)
	assert.Equal(t, 2, len(vhd))

	routeConfig, err := inMemDao.FindAllRouteConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, routeConfig)
	assert.Equal(t, 1, len(routeConfig))
}

func TestV3Service_RegisterVirtualServiceConfig_MetadataGatewayField(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	configresources.RegisterResource(v3Service.GetVirtualServiceResource())

	config := configresources.ConfigResource{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "VirtualService",
		Metadata: map[string]interface{}{
			"name":    "test-vs",
			"gateway": "gateway1",
		},
		Spec: createVirtualService(false),
	}
	_, cpErr := configresources.HandleConfigResource(nil, config)
	assert.Nil(t, cpErr)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))

	config.Metadata["gateway"] = ""
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["gateway"] = "   "
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["gateway"] = 123
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["gateway"] = nil
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)
}

func TestV3Service_RegisterVirtualServiceConfig_MetadataNameField(t *testing.T) {
	v3Service, inMemDao := getV3Service()
	configresources.RegisterResource(v3Service.GetVirtualServiceResource())

	config := configresources.ConfigResource{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "VirtualService",
		Metadata: map[string]interface{}{
			"name":    "test-vs",
			"gateway": "gateway1",
		},
		Spec: createVirtualService(false),
	}
	_, cpErr := configresources.HandleConfigResource(nil, config)
	assert.Nil(t, cpErr)
	routes, err := inMemDao.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, routes)
	assert.Equal(t, 1, len(routes))

	config.Metadata["name"] = ""
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["name"] = "   "
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["name"] = 123
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)

	config.Metadata["name"] = nil
	_, cpErr = configresources.HandleConfigResource(nil, config)
	assert.NotNil(t, cpErr)
}

func TestService_SingleOverriddenWithTrueValueForRoutingConfigRequestV3(t *testing.T) {
	srv, _ := getV3Service()
	spec := createV3Request(true)
	isOverridden := srv.GetRoutingRequestResource().GetDefinition().IsOverriddenByCR(nil, nil, &spec)
	assert.True(t, isOverridden)
}

func TestService_SingleOverriddenWithTrueValueForVirtualService(t *testing.T) {
	srv, _ := getV3Service()
	spec := createVirtualService(true)
	isOverridden := srv.GetVirtualServiceResource().GetDefinition().IsOverriddenByCR(nil, nil, &spec)
	assert.True(t, isOverridden)
}

func TestService_SingleOverriddenWithTrueValueForRouteDeleteRequestV3(t *testing.T) {

	srv, _ := getV3Service()
	specs := []dto.RouteDeleteRequestV3{
		{
			Gateways:       []string{"gateway1"},
			VirtualService: "test-vs1",
			Overridden:     true,
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Version: "v1",
				Routes: []dto.RouteDeleteItem{
					{
						Prefix: "/api/v3/tenant-manager/tenant",
					},
				},
			},
		},
	}
	isOverridden := srv.GetRoutesDropResource().GetDefinition().IsOverriddenByCR(nil, nil, &specs)
	assert.True(t, isOverridden)
}

func TestService_DifferentOverriddenValuesForRouteDeleteRequestV3(t *testing.T) {

	srv, _ := getV3Service()
	specs := []dto.RouteDeleteRequestV3{
		{
			Gateways:       []string{"gateway1"},
			VirtualService: "test-vs1",
			Overridden:     true,
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Version: "v1",
				Routes: []dto.RouteDeleteItem{
					{
						Prefix: "/api/v3/tenant-manager/tenant",
					},
				},
			},
		},
		{
			Gateways:       []string{"gateway2"},
			VirtualService: "test-vs1",
			Overridden:     false,
			RouteDeleteRequest: dto.RouteDeleteRequest{
				Version: "v1",
				Routes: []dto.RouteDeleteItem{
					{
						Prefix: "/api/v3/tenant-manager/tenant",
					},
				},
			},
		},
	}
	isOverridden := srv.GetRoutesDropResource().GetDefinition().IsOverriddenByCR(nil, nil, &specs)
	assert.False(t, isOverridden)
}

//-----PRIVATE

func createV3Request(overridden bool) dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			createVirtualService(false),
		},
		Overridden: overridden,
	}
}

func createV3RequestWithHosts(hosts ...string) dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			createVirtualServiceWithHosts(hosts...),
		},
	}
}

func createVirtualServiceWithManyVirtualServicesWithDifferentClustersAndEndpoints() dto.VirtualService {
	return dto.VirtualService{
		Name: "test-vs",
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "simple",
				Value: "simple-val",
			},
			{
				Name:  "simple2",
				Value: "simple-va2",
			},
		},
		Hosts:         []string{"www.test1.com", "www.test2.com"},
		RemoveHeaders: []string{"Authorization", "simple3"},
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
							},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster",
						Endpoint: "http://tenant-manager-v1:8080",
					},
				},
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes/{routeId}",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes/{routeId}",
							},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster-2",
						Endpoint: "http://tenant-manager2-v1:8080",
					},
				},
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/services",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/services",
							},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster-3",
						Endpoint: "http://tenant-manager3-v1:8080",
					},
				},
			},
		},
	}
}

func createVirtualServiceWithManyVirtualServicesWithDifferentRoutesWithHeaders() dto.VirtualService {
	return dto.VirtualService{
		Name: "test-vs",
		AddHeaders: []dto.HeaderDefinition{
			{
				Name:  "simple",
				Value: "simple-val",
			},
			{
				Name:  "simple2",
				Value: "simple-va2",
			},
		},
		Hosts:         []string{"www.test1.com", "www.test2.com"},
		RemoveHeaders: []string{"Authorization", "simple3"},
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
							},
							AddHeaders: []dto.HeaderDefinition{
								{
									Name:  "simple",
									Value: "simple-val",
								},
								{
									Name:  "simple2",
									Value: "simple-va2",
								},
							},
							RemoveHeaders: []string{"Authorization", "simple3"},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster",
						Endpoint: "http://tenant-manager-v1:8080",
					},
				},
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/tenants",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/tenants",
							},
							AddHeaders: []dto.HeaderDefinition{
								{
									Name:  "simple4",
									Value: "simple-val4",
								},
								{
									Name:  "simple5",
									Value: "simple-va5",
								},
							},
							RemoveHeaders: []string{"Authorization", "simple6"},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster",
						Endpoint: "http://tenant-manager-v1:8080",
					},
				},
			},
		},
	}
}

func createVirtualService(overridden bool) dto.VirtualService {
	return dto.VirtualService{
		Name:          "test-vs",
		AddHeaders:    []dto.HeaderDefinition{},
		Hosts:         []string{"www.google.com"},
		RemoveHeaders: []string{},
		Overridden:    overridden,
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
							},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster",
						Endpoint: "http://tenant-manager-v1:8080",
					},
				},
			},
		},
	}
}

func createVirtualServiceWithHosts(hosts ...string) dto.VirtualService {
	return dto.VirtualService{
		Name:          "test-vs",
		AddHeaders:    []dto.HeaderDefinition{},
		Hosts:         hosts,
		RemoveHeaders: []string{},
		RouteConfiguration: dto.RouteConfig{
			Routes: []dto.RouteV3{
				{
					Rules: []dto.Rule{
						{
							PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
							Match: dto.RouteMatch{
								Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
							},
						},
					},
					Destination: dto.RouteDestination{
						Cluster:  "test-cluster",
						Endpoint: "http://tenant-manager-v1:8080",
					},
				},
			},
		},
	}
}

func createV3RequestWithEmptyDomains() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name:          "test-vs",
				AddHeaders:    []dto.HeaderDefinition{},
				Hosts:         []string{},
				RemoveHeaders: []string{},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithSameRoutes() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestgetV3ServiceWithIncorrectRoute() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name:          "test-vs",
				AddHeaders:    []dto.HeaderDefinition{},
				Hosts:         []string{},
				RemoveHeaders: []string{},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
								{
									PrefixRewrite: "",
									Match: dto.RouteMatch{
										Prefix: "",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithSameRoutesWithDifferentVersions() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Version: "v1",
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Version: "v2",
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithDifferentRoutes() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/tenants",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/tenants",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithDifferentRoutesWithHeaders() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
									AddHeaders: []dto.HeaderDefinition{
										{
											Name:  "simple",
											Value: "simple-val",
										},
										{
											Name:  "simple2",
											Value: "simple-va2",
										},
									},
									RemoveHeaders: []string{"Authorization", "simple3"},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/tenants",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/tenants",
									},
									AddHeaders: []dto.HeaderDefinition{
										{
											Name:  "simple4",
											Value: "simple-val4",
										},
										{
											Name:  "simple5",
											Value: "simple-va5",
										},
									},
									RemoveHeaders: []string{"Authorization", "simple6"},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpoints() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes/{routeId}",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes/{routeId}",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster-2",
								Endpoint: "http://tenant-manager2-v1:8080",
							},
						},
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/services",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/services",
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster-3",
								Endpoint: "http://tenant-manager3-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func createV3RequestWithManyVirtualServicesWithDifferentClustersAndEndpointsWithHeaderMatchers() dto.RoutingConfigRequestV3 {
	return dto.RoutingConfigRequestV3{
		Gateways: []string{"gateway1", "gateway2"},
		VirtualServices: []dto.VirtualService{
			{
				Name: "test-vs1",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs1.com", "www.test2-vs1.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
										HeaderMatchers: []dto.HeaderMatcher{
											{
												Name:       "some-header",
												ExactMatch: "some-value",
											},
										},
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster",
								Endpoint: "http://tenant-manager-v1:8080",
							},
						},
					},
				},
			},
			{
				Name: "test-vs2",
				AddHeaders: []dto.HeaderDefinition{
					{
						Name:  "simple",
						Value: "simple-val",
					},
					{
						Name:  "simple2",
						Value: "simple-va2",
					},
				},
				Hosts:         []string{"www.test1-vs2.com", "www.test2-vs2.com"},
				RemoveHeaders: []string{"Authorization", "simple3"},
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/routes",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/routes",
										HeaderMatchers: []dto.HeaderMatcher{
											{
												Name:       "some-header",
												ExactMatch: "some-value",
											},
										},
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster-2",
								Endpoint: "http://tenant-manager2-v1:8080",
							},
						},
						{
							Rules: []dto.Rule{
								{
									PrefixRewrite: "/api/v3/tenant/{tenantId}/services",
									Match: dto.RouteMatch{
										Prefix: "/api/v3/tenant-manager/tenant/{tenantId}/services",
										HeaderMatchers: []dto.HeaderMatcher{
											{
												Name:       "some-header",
												ExactMatch: "some-value",
											},
										},
									},
								},
							},
							Destination: dto.RouteDestination{
								Cluster:  "test-cluster-3",
								Endpoint: "http://tenant-manager3-v1:8080",
							},
						},
					},
				},
			},
		},
	}
}

func getV3Service() (*Service, *dao.InMemDao) {
	entityService := entity.NewService("v1")
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, []func([]memdb.Change) error{flushChanges})
	routeModeService := routingmode.NewService(genericDao, "v1")
	routeComponentsFactory := factory.NewComponentsFactory(entityService)
	registrationService := route.NewRegistrationService(routeComponentsFactory, entityService, genericDao, eventBus, routeModeService)
	v3RequestProcessor := registration.NewV3RequestProcessor(genericDao)
	v3Service := NewV3Service(genericDao, eventBus, routeModeService, registrationService, entityService, v3RequestProcessor)
	_, _ = genericDao.WithWTx(func(dao dao.Repository) error {
		_ = dao.SaveDeploymentVersion(&domain.DeploymentVersion{
			Version: "v1",
			Stage:   domain.ActiveStage,
		})
		return nil
	})
	return v3Service, genericDao
}

func flushChanges(changes []memdb.Change) error {
	return nil
}

type idGeneratorMock struct {
	seq int32
}

func (generator *idGeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&generator.seq, 1))
	}
	return nil
}
