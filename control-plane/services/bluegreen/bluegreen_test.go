package bluegreen

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/constancy"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/loadbalance"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	routeKeyFormat             = "%s||%s||%s"
	testVirtualHostName        = "test-nodegroup"
	testRouteConfigurationName = "test-nodegroup-routes"
	testNodeGroupName          = "test-nodegroup"
	firstTestClusterName       = "test-clustername"
	secondTestClusterName      = "ext-authz"
	firstTestEndpointAddress   = "http://test-clustername"
	secondTestEndpointAddress  = "http://test-address2"
)

func Test_Promote_RemoveCandidates_PushLegacy(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)

	// execute promote for v2
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v2, 0)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	assert.Nil(t, getVersion(dVersions, "v3"))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.LegacyStage))
	v2Promoted := getVersionByStage(dVersions, "v2", domain.ActiveStage)
	assert.NotNil(t, v2Promoted)
	assert.True(t, v2Promoted.UpdatedWhen.After(v2.UpdatedWhen))
	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.Nil(t, getVersion(dVersions, "v3"))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))
	activeDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.ActiveStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, activeDVersions)
	assert.Equal(t, 1, len(activeDVersions))
	assert.Equal(t, domain.ActiveStage, activeDVersions[0].Stage)
	assert.Equal(t, v2.Version, activeDVersions[0].Version)
	legacyDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.LegacyStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, legacyDVersions)
	assert.Equal(t, 1, len(legacyDVersions))
	assert.Equal(t, domain.LegacyStage, legacyDVersions[0].Stage)
	assert.Equal(t, v1.Version, legacyDVersions[0].Version)
	candidateDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.CandidateStage)
	assert.Nil(t, err)
	assert.Empty(t, candidateDVersions)
	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 2, len(actualRoutes))

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_Promote_PushToArchive(t *testing.T) {
	memStorage, bgService := prepareTest("v2")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.LegacyStage)
	v2 := domain.NewDeploymentVersion("v2", domain.ActiveStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)

	// execute promote for v3
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v3, 1)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 3, len(dVersions))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.ArchivedStage))
	assert.NotNil(t, getVersionByStage(dVersions, "v2", domain.LegacyStage))
	v3Promoted := getVersionByStage(dVersions, "v3", domain.ActiveStage)
	assert.NotNil(t, v3Promoted)
	assert.True(t, v3Promoted.UpdatedWhen.After(v3.UpdatedWhen))
	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 3, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.ArchivedStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v3", domain.ActiveStage))
	archivedDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.ArchivedStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, archivedDVersions)
	assert.Equal(t, 1, len(archivedDVersions))
	assert.Equal(t, domain.ArchivedStage, archivedDVersions[0].Stage)
	assert.Equal(t, v1.Version, archivedDVersions[0].Version)
	legacyDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.LegacyStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, legacyDVersions)
	assert.Equal(t, 1, len(legacyDVersions))
	assert.Equal(t, domain.LegacyStage, legacyDVersions[0].Stage)
	assert.Equal(t, v2.Version, legacyDVersions[0].Version)
	activeDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.ActiveStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, activeDVersions)
	assert.Equal(t, 1, len(activeDVersions))
	assert.Equal(t, domain.ActiveStage, activeDVersions[0].Stage)
	assert.Equal(t, v3.Version, activeDVersions[0].Version)
	candidateDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.CandidateStage)
	assert.Nil(t, err)
	assert.Empty(t, candidateDVersions)
	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 6, len(actualEndpoints))

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_Promote_Rollback_V1Unique(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)

	// add unique route to v1 endpoint
	uniqueRoute := createRoute(0, "/api/v1/test-unique", "v1")
	uniqueRoute.ClusterName = firstTestClusterName
	// and emulate forbidden route creation during routes registration
	forbiddenRoute := createRoute(0, "/api/v1/test-unique", "v2")
	forbiddenRoute.Autogenerated = true
	forbiddenRoute.DirectResponseCode = uint32(404)
	forbiddenRoute.ClusterName = firstTestClusterName
	vHosts, err := memStorage.FindAllVirtualHosts()
	assert.Nil(t, err)
	for _, vHost := range vHosts {
		if vHost.Name == testVirtualHostName {
			uniqueRoute.VirtualHostId = vHost.Id
			forbiddenRoute.VirtualHostId = vHost.Id
			break
		}
	}
	_, err = memStorage.WithWTx(func(dao dao.Repository) error {
		saveRoutes(t, dao, uniqueRoute, forbiddenRoute)
		return nil
	})
	assert.Nil(t, err)

	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 5, len(actualRoutes))
	var uniqueRouteV1 *domain.Route = nil
	var uniqueRouteV2 *domain.Route = nil
	for _, route := range actualRoutes {
		if route.Prefix == "/api/v1/test-unique" {
			if route.DeploymentVersion == "v1" {
				uniqueRouteV1 = route
			} else if route.DeploymentVersion == "v2" {
				uniqueRouteV2 = route
			} else {
				t.Errorf("Route %v has unexpected deployment version! Expected versions: v1 or v2", *route)
			}
		}
	}
	assert.NotNil(t, uniqueRouteV1)
	assert.NotNil(t, uniqueRouteV2)
	assert.False(t, uniqueRouteV1.Autogenerated)
	assert.True(t, uniqueRouteV2.Autogenerated)
	assert.Equal(t, uint32(404), uniqueRouteV2.DirectResponseCode)

	// execute promote for v2
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v2, 3)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.LegacyStage))
	v2Promoted := getVersionByStage(dVersions, "v2", domain.ActiveStage)
	assert.NotNil(t, v2Promoted)
	assert.True(t, v2Promoted.UpdatedWhen.After(v2.UpdatedWhen))

	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))

	//check routes
	actualRoutes, err = memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	uniqueRouteV1 = nil
	uniqueRouteV2 = nil
	for _, route := range actualRoutes {
		if route.Prefix == "/api/v1/test-unique" {
			if route.DeploymentVersion == "v1" {
				uniqueRouteV1 = route
			} else if route.DeploymentVersion == "v2" {
				uniqueRouteV2 = route
			} else {
				t.Errorf("Route %v has unexpected deployment version! Expected versions: v1 or v2", *route)
			}
		}
	}
	assert.NotNil(t, uniqueRouteV1)
	assert.Nil(t, uniqueRouteV2)
	assert.False(t, uniqueRouteV1.Autogenerated)

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))
	for _, endpoint := range actualEndpoints {
		assert.Equal(t, 2, len(endpoint.DeploymentVersion))
		if strings.HasPrefix(endpoint.Address, firstTestEndpointAddress) {
			expectedAddr := fmt.Sprintf("%s-%s:8080", firstTestEndpointAddress, endpoint.DeploymentVersion)
			assert.Equal(t, expectedAddr, endpoint.Address)
		}
	}

	// execute rollback
	time.Sleep(10 * time.Millisecond)
	dVersions, err = bgService.Rollback(nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	v1RolledBack := getVersionByStage(dVersions, "v1", domain.ActiveStage)
	assert.NotNil(t, v1RolledBack)
	assert.True(t, v1RolledBack.UpdatedWhen.After(v1.UpdatedWhen))
	assert.NotNil(t, getVersionByStage(dVersions, "v2", domain.CandidateStage))

	//check deployment versions
	actualDVersions, err = memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.ActiveStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.CandidateStage))

	//check routes
	actualRoutes, err = memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 4, len(actualRoutes))
	uniqueRouteV1 = nil
	uniqueRouteV2 = nil
	for _, route := range actualRoutes {
		if route.Prefix == "/api/v1/test-unique" {
			if route.DeploymentVersion == "v1" {
				uniqueRouteV1 = route
			} else if route.DeploymentVersion == "v2" {
				uniqueRouteV2 = route
			} else {
				t.Errorf("Route %v has unexpected deployment version! Expected versions: v1 or v2", *route)
			}
		}
	}
	assert.NotNil(t, uniqueRouteV1)
	assert.NotNil(t, uniqueRouteV2)
	assert.False(t, uniqueRouteV1.Autogenerated)
	assert.True(t, uniqueRouteV2.Autogenerated)
	assert.Equal(t, uint32(404), uniqueRouteV2.DirectResponseCode)

	//check clusters
	actualClusters, err = memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err = memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))
	for _, endpoint := range actualEndpoints {
		assert.Equal(t, 2, len(endpoint.DeploymentVersion))
		if strings.HasPrefix(endpoint.Address, firstTestEndpointAddress) {
			expectedAddr := fmt.Sprintf("%s-%s:8080", firstTestEndpointAddress, endpoint.DeploymentVersion)
			assert.Equal(t, expectedAddr, endpoint.Address)
		}
	}

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_Promote_Rollback_V2Unique(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)

	// add unique route to v1 endpoint
	uniqueRoute := createRoute(0, "/api/v1/test-unique", "v2")
	uniqueRoute.ClusterName = firstTestClusterName
	vHosts, err := memStorage.FindAllVirtualHosts()
	assert.Nil(t, err)
	for _, vHost := range vHosts {
		if vHost.Name == testVirtualHostName {
			uniqueRoute.VirtualHostId = vHost.Id
			break
		}
	}
	_, err = memStorage.WithWTx(func(dao dao.Repository) error {
		saveRoutes(t, dao, uniqueRoute)
		return nil
	})
	assert.Nil(t, err)

	// execute promote for v2
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v2, 3)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.LegacyStage))
	v2Promoted := getVersionByStage(dVersions, "v2", domain.ActiveStage)
	assert.NotNil(t, v2Promoted)
	assert.True(t, v2Promoted.UpdatedWhen.After(v2.UpdatedWhen))

	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))

	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 4, len(actualRoutes))
	var uniqueRouteV1 *domain.Route = nil
	var uniqueRouteV2 *domain.Route = nil
	for _, route := range actualRoutes {
		if route.Prefix == "/api/v1/test-unique" {
			if route.DeploymentVersion == "v1" {
				uniqueRouteV1 = route
			} else if route.DeploymentVersion == "v2" {
				uniqueRouteV2 = route
			} else {
				t.Errorf("Route %v has unexpected deployment version! Expected versions: v1 or v2", *route)
			}
		}
	}
	assert.NotNil(t, uniqueRouteV1)
	assert.NotNil(t, uniqueRouteV2)
	assert.True(t, uniqueRouteV1.Autogenerated)
	assert.Equal(t, uint32(404), uniqueRouteV1.DirectResponseCode)
	assert.False(t, uniqueRouteV2.Autogenerated)

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))
	for _, endpoint := range actualEndpoints {
		assert.Equal(t, 2, len(endpoint.DeploymentVersion))
		if strings.HasPrefix(endpoint.Address, firstTestEndpointAddress) {
			expectedAddr := fmt.Sprintf("%s-%s:8080", firstTestEndpointAddress, endpoint.DeploymentVersion)
			assert.Equal(t, expectedAddr, endpoint.Address)
		}
	}

	// execute rollback
	time.Sleep(10 * time.Millisecond)
	dVersions, err = bgService.Rollback(nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	v1RolledBack := getVersionByStage(dVersions, "v1", domain.ActiveStage)
	assert.NotNil(t, v1RolledBack)
	assert.True(t, v1RolledBack.UpdatedWhen.After(v1.UpdatedWhen))
	assert.NotNil(t, getVersionByStage(dVersions, "v2", domain.CandidateStage))

	//check deployment versions
	actualDVersions, err = memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.ActiveStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.CandidateStage))

	//check routes
	actualRoutes, err = memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 3, len(actualRoutes))
	uniqueRouteV1 = nil
	uniqueRouteV2 = nil
	for _, route := range actualRoutes {
		if route.Prefix == "/api/v1/test-unique" {
			if route.DeploymentVersion == "v1" {
				uniqueRouteV1 = route
			} else if route.DeploymentVersion == "v2" {
				uniqueRouteV2 = route
			} else {
				t.Errorf("Route %v has unexpected deployment version! Expected versions: v1 or v2", *route)
			}
		}
	}
	assert.Nil(t, uniqueRouteV1)
	assert.NotNil(t, uniqueRouteV2)
	assert.False(t, uniqueRouteV2.Autogenerated)

	//check clusters
	actualClusters, err = memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err = memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))
	for _, endpoint := range actualEndpoints {
		assert.Equal(t, 2, len(endpoint.DeploymentVersion))
		if strings.HasPrefix(endpoint.Address, firstTestEndpointAddress) {
			expectedAddr := fmt.Sprintf("%s-%s:8080", firstTestEndpointAddress, endpoint.DeploymentVersion)
			assert.Equal(t, expectedAddr, endpoint.Address)
		}
	}

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_Promote_Rollback_NonBgService(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveNonBlueGreenService(t, memStorage)

	// execute promote for v2
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v2, 3)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.LegacyStage))
	v2Promoted := getVersionByStage(dVersions, "v2", domain.ActiveStage)
	assert.NotNil(t, v2Promoted)
	assert.True(t, v2Promoted.UpdatedWhen.After(v2.UpdatedWhen))

	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))

	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 1, len(actualRoutes))
	assert.False(t, actualRoutes[0].Autogenerated)
	assert.Equal(t, "v2", actualRoutes[0].DeploymentVersion)

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 2, len(actualEndpoints))
	assert.Equal(t, "v2", actualEndpoints[0].DeploymentVersion)
	assert.Equal(t, "v2", actualEndpoints[1].DeploymentVersion)

	// execute rollback
	time.Sleep(10 * time.Millisecond)
	dVersions, err = bgService.Rollback(nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	v1RolledBack := getVersionByStage(dVersions, "v1", domain.ActiveStage)
	assert.NotNil(t, v1RolledBack)
	assert.True(t, v1RolledBack.UpdatedWhen.After(v1.UpdatedWhen))
	assert.NotNil(t, getVersionByStage(dVersions, "v2", domain.CandidateStage))

	//check deployment versions
	actualDVersions, err = memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.ActiveStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.CandidateStage))

	//check routes
	actualRoutes, err = memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 1, len(actualRoutes))
	assert.False(t, actualRoutes[0].Autogenerated)
	assert.Equal(t, "v1", actualRoutes[0].DeploymentVersion)

	//check clusters
	actualClusters, err = memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err = memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 2, len(actualEndpoints))
	assert.Equal(t, "v1", actualEndpoints[0].DeploymentVersion)
	assert.Equal(t, "v1", actualEndpoints[1].DeploymentVersion)

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_DeleteCandidate(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.LegacyStage)
	v2 := domain.NewDeploymentVersion("v2", domain.ActiveStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)

	// create hash policy for v2 and v3
	v2Policies := []*domain.HashPolicy{{HeaderName: "BID"}}
	v3Policies := []*domain.HashPolicy{{HeaderName: "X-BID"}, {CookieName: "JSESSION"}}
	err := bgService.loadBalanceService.ApplyLoadBalanceConfig(ctx, firstTestClusterName, v2.Version, v2Policies)
	assert.Nil(t, err)
	err = bgService.loadBalanceService.ApplyLoadBalanceConfig(ctx, firstTestClusterName, v3.Version, v3Policies)
	assert.Nil(t, err)

	// delete candidate v3
	err = bgService.DeleteCandidate(nil, v3)
	assert.Nil(t, err)

	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))
	candidateDVersions, err := memStorage.FindDeploymentVersionsByStage(domain.CandidateStage)
	assert.Nil(t, err)
	assert.Empty(t, candidateDVersions)

	//check routes
	actualRoutes, err := memStorage.FindAllRoutes()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualRoutes)
	assert.Equal(t, 2, len(actualRoutes))

	//check clusters
	actualClusters, err := memStorage.FindAllClusters()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualClusters)
	assert.Equal(t, 2, len(actualClusters))

	//check endpoints
	actualEndpoints, err := memStorage.FindAllEndpoints()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualEndpoints)
	assert.Equal(t, 4, len(actualEndpoints))
	v3Endpoints, err := memStorage.FindEndpointsByDeploymentVersion(v3.Version)
	assert.Nil(t, err)
	assert.Empty(t, v3Endpoints)

	//check hash policy
	actualHashPolicy, err := memStorage.FindHashPolicyByClusterAndVersions(firstTestClusterName, v2.Version)
	assert.Nil(t, err)
	assert.NotEmpty(t, actualHashPolicy)
	assert.Equal(t, 1, len(actualHashPolicy))
	actualHashPolicy, err = memStorage.FindHashPolicyByClusterAndVersions(firstTestClusterName, v3.Version)
	assert.Nil(t, err)
	assert.Empty(t, actualHashPolicy)

	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Empty(t, actualMsVersions)
}

func Test_ParseMicroserviceHost(t *testing.T) {
	service := Service{}
	hostWithoutVersion := "umi"
	hostWithVersion := "umi-v2"
	expectedVersion := "v2"

	parsedHost, parsedVersion := service.parseMicroserviceHost(hostWithoutVersion)
	assert.Empty(t, parsedVersion)
	assert.Equal(t, hostWithoutVersion, parsedHost)

	parsedHost, parsedVersion = service.parseMicroserviceHost(hostWithVersion)
	assert.Equal(t, expectedVersion, parsedVersion)
	assert.Equal(t, hostWithoutVersion, parsedHost)
}

func createAndSaveDefault(t *testing.T, memStorage *dao.InMemDao) {
	_, err := memStorage.WithWTx(func(dao dao.Repository) error {
		routeConfig := createDefaultRouteConfiguration()
		assert.Nil(t, dao.SaveRouteConfig(routeConfig))
		assert.Nil(t, dao.SaveNodeGroup(createDefaultNodeGroup()))
		vH := createDefaultVirtualHost(routeConfig.Id)
		assert.Nil(t, dao.SaveVirtualHost(vH))
		createAndSaveDefaultRoutes(t, dao, vH.Id)
		createAndSaveDefaultClustersWithEndpoints(t, dao)
		return nil
	})
	assert.Empty(t, err)
}

func createAndSaveNonBlueGreenService(t *testing.T, memStorage *dao.InMemDao) {
	_, err := memStorage.WithWTx(func(dao dao.Repository) error {
		routeConfig := createDefaultRouteConfiguration()
		assert.Nil(t, dao.SaveRouteConfig(routeConfig))
		assert.Nil(t, dao.SaveNodeGroup(createDefaultNodeGroup()))
		vH := createDefaultVirtualHost(routeConfig.Id)
		assert.Nil(t, dao.SaveVirtualHost(vH))
		route := createRoute(vH.Id, "/api/v1/test", "v1")
		route.ClusterName = firstTestClusterName
		assert.Nil(t, dao.SaveRoute(route))
		cluster := createDefaultCluster(firstTestClusterName)
		assert.Nil(t, dao.SaveCluster(cluster))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, firstTestEndpointAddress, "v1")))
		cluster = createDefaultCluster(secondTestClusterName)
		assert.Nil(t, dao.SaveCluster(cluster))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, secondTestEndpointAddress, "v1")))
		return nil
	})
	assert.Empty(t, err)
}

func createAndSaveDefaultClustersWithEndpoints(t *testing.T, dao dao.Repository) {
	createAndSaveDefaultClusterWithEndpoints(t, dao, firstTestClusterName, firstTestEndpointAddress, true)
	createAndSaveDefaultClusterWithEndpoints(t, dao, secondTestClusterName, secondTestEndpointAddress, false)
}

func createAndSaveDefaultClusterWithEndpoints(t *testing.T, dao dao.Repository, clusterName, endpointAddress string, endpointsHaveVersion bool) {
	cluster := createDefaultCluster(clusterName)
	assert.Nil(t, dao.SaveCluster(cluster))
	if endpointsHaveVersion {
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress+"-v1:8080", "v1")))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress+"-v2:8080", "v2")))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress+"-v3:8080", "v3")))
	} else {
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress, "v1")))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress, "v2")))
		assert.Nil(t, dao.SaveEndpoint(createDefaultEndpoint(cluster.Id, endpointAddress, "v3")))
	}
}

func createAndSaveDefaultRoutes(t *testing.T, dao dao.Repository, id int32) {
	routeV1 := createRoute(id, "/api/v1/test", "v1")
	routeV2 := createRoute(id, "/api/v1/test", "v2")
	routeV3 := createRoute(id, "/api/v1/test", "v3")
	saveRoutes(t, dao, routeV1, routeV2, routeV3)
}

func createRoute(vHostId int32, prefix string, dV string) *domain.Route {
	return &domain.Route{
		RouteKey:                 fmt.Sprintf(routeKeyFormat, "", prefix, dV),
		Prefix:                   prefix,
		DeploymentVersion:        dV,
		InitialDeploymentVersion: dV,
		Uuid:                     uuid.New().String(),
		VirtualHostId:            vHostId,
	}
}

func createDefaultCluster(clusterName string) *domain.Cluster {
	return domain.NewCluster(clusterName, false)
}

func createDefaultRouteConfiguration() *domain.RouteConfiguration {
	return &domain.RouteConfiguration{
		Name:        testRouteConfigurationName,
		NodeGroupId: testNodeGroupName,
	}
}

func createDefaultNodeGroup() *domain.NodeGroup {
	return &domain.NodeGroup{
		Name: testNodeGroupName,
	}
}

func createDefaultEndpoint(clusterId int32, endpointAddress, dVersion string) *domain.Endpoint {
	return domain.NewEndpoint(endpointAddress, 8080, dVersion, dVersion, clusterId)
}

func createDefaultVirtualHost(routeConfigId int32) *domain.VirtualHost {
	return &domain.VirtualHost{
		Id:   routeConfigId,
		Name: testVirtualHostName,
		Domains: []*domain.VirtualHostDomain{
			{
				Domain: "*",
			},
		},
	}
}

func saveDeploymentVersions(t *testing.T, storage *dao.InMemDao, dVs ...*domain.DeploymentVersion) {
	_, err := storage.WithWTx(func(dao dao.Repository) error {
		for _, dV := range dVs {
			assert.Nil(t, dao.SaveDeploymentVersion(dV))
		}
		return nil
	})
	assert.Nil(t, err)
}

func saveRoutes(t *testing.T, dao dao.Repository, routes ...*domain.Route) {
	for _, route := range routes {
		assert.Nil(t, dao.SaveRoute(route))
	}
}

func getInMemRepo() *dao.InMemDao {
	return dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, []func([]memdb.Change) error{flushChanges})
}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

func flushChanges(changes []memdb.Change) error {
	flusher := &constancy.Flusher{BatchTm: &batchTransactionManagerMock{}}
	return flusher.Flush(changes)
}

type batchTransactionManagerMock struct{}

func (tm *batchTransactionManagerMock) WithTxBatch(_ func(tx constancy.BatchStorage) error) error {
	return nil
}

func (tm *batchTransactionManagerMock) IsCurrentPodDefinedAsMaster() (bool, error) {
	return true, nil
}

func prepareTest(defaultDeploymentVersion string) (*dao.InMemDao, *Service) {
	memStorage := getInMemRepo()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(memStorage, internalBus, internalBus, nil, nil)
	entityService := entity.NewService(defaultDeploymentVersion)
	loadBalanceService := loadbalance.NewLoadBalanceService(memStorage, entityService, eventBus)
	versionRegistry := NewVersionsRegistry(memStorage, entityService, eventBus)
	bgService := NewService(entityService, loadBalanceService, memStorage, eventBus, versionRegistry)
	return memStorage, bgService
}

func getVersionByStage(versions []*domain.DeploymentVersion, version string, stage string) *domain.DeploymentVersion {
	dVer := getVersion(versions, version)
	if dVer != nil && dVer.Stage == stage {
		return dVer
	}
	return nil
}

func getVersion(versions []*domain.DeploymentVersion, version string) *domain.DeploymentVersion {
	for _, dVer := range versions {
		if dVer.Version == version {
			return dVer
		}
	}
	return nil
}
