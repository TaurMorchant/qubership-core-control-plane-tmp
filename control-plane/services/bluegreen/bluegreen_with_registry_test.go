package bluegreen

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func Test_Promote_RemoveCandidates_PushLegacy_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)
	createDefaultMsVersions(t, memStorage)

	// execute promote for v2
	time.Sleep(10 * time.Millisecond)
	dVersions, err := bgService.Promote(nil, v2, 0)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	assert.NotNil(t, getVersionByStage(dVersions, "v1", domain.LegacyStage))
	v2Promoted := getVersionByStage(dVersions, "v2", domain.ActiveStage)
	assert.NotNil(t, v2Promoted)
	assert.True(t, v2Promoted.UpdatedWhen.After(v2.UpdatedWhen))
	assert.Nil(t, getVersion(dVersions, "v3"))

	//check deployment versions
	actualDVersions, err := memStorage.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualDVersions)
	assert.Equal(t, 2, len(actualDVersions))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v1", domain.LegacyStage))
	assert.NotNil(t, getVersionByStage(actualDVersions, "v2", domain.ActiveStage))
	assert.Nil(t, getVersion(actualDVersions, "v3"))
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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))
}

func Test_Promote_PushToArchive_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v2")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.LegacyStage)
	v2 := domain.NewDeploymentVersion("v2", domain.ActiveStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)
	createDefaultMsVersions(t, memStorage)

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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))
	assert.True(t, hasMsVersion(actualMsVersions, "v3", "v3"))
}

func Test_Promote_Rollback_V1Unique_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)
	createDefaultMsVersions(t, memStorage)

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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))

	// execute rollback
	time.Sleep(10 * time.Millisecond)
	dVersions, err = bgService.Rollback(nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, dVersions)
	assert.Equal(t, 2, len(dVersions))
	v1RolledBack := getVersionByStage(dVersions, "v1", domain.ActiveStage)
	assert.NotNil(t, v1RolledBack)
	assert.NotNil(t, getVersionByStage(dVersions, "v2", domain.CandidateStage))
	assert.True(t, v1RolledBack.UpdatedWhen.After(v1.UpdatedWhen))

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

	//check microservice versions
	actualMsVersions, err = memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))
}

func Test_Promote_Rollback_V2Unique_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)
	createDefaultMsVersions(t, memStorage)

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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))

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

	//check microservice versions
	actualMsVersions, err = memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))
}

func Test_Promote_Rollback_NonBgService_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	v2 := domain.NewDeploymentVersion("v2", domain.CandidateStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveNonBlueGreenService(t, memStorage)
	createDefaultMsVersion(t, memStorage, "v1")

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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v1"))

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

	//check microservice versions
	actualMsVersions, err = memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
}

func Test_DeleteCandidate_Reg(t *testing.T) {
	memStorage, bgService := prepareTest("v1")
	// create and save three deployment versions
	v1 := domain.NewDeploymentVersion("v1", domain.LegacyStage)
	v2 := domain.NewDeploymentVersion("v2", domain.ActiveStage)
	v3 := domain.NewDeploymentVersion("v3", domain.CandidateStage)
	saveDeploymentVersions(t, memStorage, v1, v2, v3)

	// create default route configuration
	createAndSaveDefault(t, memStorage)
	createDefaultMsVersions(t, memStorage)

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

	//check microservice versions
	actualMsVersions, err := memStorage.FindAllMicroserviceVersions()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualMsVersions))
	assert.True(t, hasMsVersion(actualMsVersions, "v1", "v1"))
	assert.True(t, hasMsVersion(actualMsVersions, "v2", "v2"))
}

func createDefaultMsVersions(t *testing.T, memStorage *dao.InMemDao) {
	_, _ = memStorage.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     firstTestClusterName,
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}))
		assert.Nil(t, dao.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     firstTestClusterName,
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v2",
		}))
		assert.Nil(t, dao.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     firstTestClusterName,
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v3",
			InitialDeploymentVersion: "v3",
		}))
		return nil
	})
}

func createDefaultMsVersion(t *testing.T, memStorage *dao.InMemDao, version string) {
	_, _ = memStorage.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, dao.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     firstTestClusterName,
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        version,
			InitialDeploymentVersion: version,
		}))
		return nil
	})
}

func hasMsVersion(slice []*domain.MicroserviceVersion, version, initialVersion string) bool {
	for _, msVersion := range slice {
		if msVersion.DeploymentVersion == version && msVersion.InitialDeploymentVersion == initialVersion {
			return true
		}
	}
	return false
}
