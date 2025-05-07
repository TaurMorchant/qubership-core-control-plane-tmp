package main

import (
	"encoding/json"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/lib"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	asrt "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"
)

const TestClusterName = "test-service||test-service||8080"
const TestEndpointV2 = "test-service-v2:8080"
const TestEndpointV3 = "test-service-v3:8080"

func Test_IT_BlueGreen_DeleteCandidate(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	versions := getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	internalGateway.RegisterRoutesAndWait(assert, 60*time.Second, "v2", dto.RouteV3{
		Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
		Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v2/test-service/"}}},
	})

	versionsMap := getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)

	deletedVersion := deleteVersion(assert, "v2")
	assert.Equal("v2", deletedVersion.Version)
	assert.Equal(domain.CandidateStage, deletedVersion.Stage)

	versions = getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)
}

func Test_IT_BlueGreen_GetMsVersion(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	internalGateway.RegisterRoutesAndWait(assert, 60*time.Second, "v1", dto.RouteV3{
		Destination: dto.RouteDestination{Cluster: "single-version-srv", Endpoint: "single-version-srv-v1:8080"},
		Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/single-version-srv/"}}},
	})
	internalGateway.RegisterRoutesAndWait(assert, 60*time.Second, "v2", dto.RouteV3{
		Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
		Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v2/test-service/"}}},
	})
	version := requestMicroserviceVersion(assert, "control-plane")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "single-version-srv-v1")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v2")
	assert.Equal("v2", version)

	promoteAndWait(assert, 60*time.Second, "v2", 5)
	version = requestMicroserviceVersion(assert, "control-plane")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "single-version-srv-v1")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v2")
	assert.Equal("", version)

	internalGateway.RegisterRoutesAndWait(assert, 60*time.Second, "v3", dto.RouteV3{
		Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV3},
		Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v3/test-service/"}}},
	})
	version = requestMicroserviceVersion(assert, "control-plane")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "single-version-srv-v1")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v2")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v3")
	assert.Equal("v3", version)

	internalGateway.RegisterRoutesAndWait(assert, 60*time.Second, "v4", dto.RouteV3{
		Destination: dto.RouteDestination{Cluster: "another-service", Endpoint: "another-service-v4:8080"},
		Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v4/another-service/"}}},
	})
	version = requestMicroserviceVersion(assert, "control-plane")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "single-version-srv-v1")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v2:8080")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v3:8888")
	assert.Equal("v3", version)
	version = requestMicroserviceVersion(assert, "another-service-v4")
	assert.Equal("v4", version)

	_ = deleteVersion(assert, "v4")
	_ = deleteVersion(assert, "v3")
	rollbackAndWait(assert, 60*time.Second)
	version = requestMicroserviceVersion(assert, "control-plane")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "single-version-srv-v1")
	assert.Equal("", version)
	version = requestMicroserviceVersion(assert, "test-service-v2")
	assert.Equal("v2", version)

	// cleanup
	_ = deleteVersion(assert, "v2")
	cluster, err := lib.GenericDao.FindClusterByName("single-version-srv||single-version-srv||8080")
	assert.Nil(err)
	internalGateway.DeleteClusterAndWait(assert, 60*time.Second, cluster.Id, "single-version-srv||single-version-srv||8080")
}

func Test_IT_BlueGreen_PromoteAndRollback(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	srv1Container := createTraceServiceContainer(TestCluster, "v1", true)
	defer srv1Container.Purge()
	srv2Container := createTraceServiceContainer(TestCluster, "v2", true)
	defer srv2Container.Purge()

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/common"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/unique"}}},
		},
	)

	versions := getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v2",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/common"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v2/test-service/unique"}}},
		},
	)

	versionsMap := getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)

	endpoints, err := lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint before promote: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode := GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/common")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/unique")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v2")
	assert.Equal(http.StatusNotFound, statusCode)

	_, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v2/test-service/unique")
	assert.Equal(http.StatusNotFound, statusCode)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v1")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	_ = promoteAndWait(assert, 60*time.Second, "v2", 3)

	versionsMap = getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.LegacyStage, versionsMap["v1"].Stage)
	assert.Equal(domain.ActiveStage, versionsMap["v2"].Stage)

	endpoints, err = lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint after promote: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/common")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	_, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/unique")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v2")
	assert.Equal(http.StatusNotFound, statusCode)

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v2/test-service/unique")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v1")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	_ = rollbackAndWait(assert, 60*time.Second)

	versionsMap = getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)

	endpoints, err = lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint after rollback: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/common")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/common", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v1/test-service/unique")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v1")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v1", respFromService.ServiceName)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v1/test-service/unique", "v2")
	assert.Equal(http.StatusNotFound, statusCode)

	_, statusCode = GetFromTraceService(assert, internalGateway.Url+"/api/v2/test-service/unique")
	assert.Equal(http.StatusNotFound, statusCode)
	_, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v1")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/api/v2/test-service/unique", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	// cleanup v2
	deletedVersion := deleteVersion(assert, "v2")
	assert.Equal("v2", deletedVersion.Version)
	assert.Equal(domain.CandidateStage, deletedVersion.Stage)

	versions = getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	// cleanup v1 routes
	internalGateway.DeleteRoutesAndWait(assert, 60*time.Second, dto.RouteDeleteRequestV3{
		Gateways:       []string{"internal-gateway-service"},
		VirtualService: "internal-gateway-service",
		RouteDeleteRequest: dto.RouteDeleteRequest{
			Routes:  []dto.RouteDeleteItem{{"/api/v1/test-service/common"}, {"/api/v1/test-service/unique"}},
			Version: "v1",
		},
	})
}

func Test_IT_BlueGreen_Err511(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/err511"}}},
		},
	)

	versions := getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v2",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/err511"}}},
		},
	)

	versionsMap := getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)

	deadline := time.Now().Add(8 * time.Second)
	testRequestThroughGwInBackground(assert, deadline)

	for time.Now().Before(deadline) {
		time.Sleep(2 * time.Second)

		_ = promoteAndWait(assert, 60*time.Second, "v2", 3)

		versionsMap = getVersionsMap(assert)
		assert.Equal(2, len(versionsMap))
		assert.Equal(domain.LegacyStage, versionsMap["v1"].Stage)
		assert.Equal(domain.ActiveStage, versionsMap["v2"].Stage)

		time.Sleep(2 * time.Second)

		_ = rollbackAndWait(assert, 60*time.Second)

		versionsMap = getVersionsMap(assert)
		assert.Equal(2, len(versionsMap))
		assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)
		assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)
	}

	// cleanup v2
	deletedVersion := deleteVersion(assert, "v2")
	assert.Equal("v2", deletedVersion.Version)
	assert.Equal(domain.CandidateStage, deletedVersion.Stage)

	versions = getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	// cleanup v1 routes
	internalGateway.DeleteRoutesAndWait(assert, 60*time.Second, dto.RouteDeleteRequestV3{
		Gateways:       []string{"internal-gateway-service"},
		VirtualService: "internal-gateway-service",
		RouteDeleteRequest: dto.RouteDeleteRequest{
			Routes:  []dto.RouteDeleteItem{{"/api/v1/test-service/err511"}},
			Version: "v1",
		},
	})
}

func testRequestThroughGwInBackground(assert *asrt.Assertions, deadline time.Time) {
	for t := 1; t < 20; t++ {
		go func() {
			for time.Now().Before(deadline) {
				_, code := internalGateway.SendGatewayRequest(assert, http.MethodGet, "/api/v3/control-plane/versions/registry", nil, nil)
				assert.Equal(200, code)
			}
		}()
	}
}

func Test_IT_BlueGreen_PromoteAndRollbackWithRootSlash(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	srv1Container := createTraceServiceContainer(TestCluster, "v1", true)
	defer srv1Container.Purge()
	srv2Container := createTraceServiceContainer(TestCluster, "v2", true)
	defer srv2Container.Purge()

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/common"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/unique"}}},
		},
	)

	versions := getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v2",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/test-service/common"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v2/test-service/unique"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV2},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/"}}},
		},
	)

	versionsMap := getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)

	endpoints, err := lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint before promote: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode := GetFromTraceService(assert, internalGateway.Url+"/")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	_ = promoteAndWait(assert, 60*time.Second, "v2", 3)

	versionsMap = getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.LegacyStage, versionsMap["v1"].Stage)
	assert.Equal(domain.ActiveStage, versionsMap["v2"].Stage)

	endpoints, err = lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint after promote: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/", "v1")
	assert.Equal(http.StatusNotFound, statusCode)

	_ = rollbackAndWait(assert, 60*time.Second)

	versionsMap = getVersionsMap(assert)
	assert.Equal(2, len(versionsMap))
	assert.Equal(domain.CandidateStage, versionsMap["v2"].Stage)
	assert.Equal(domain.ActiveStage, versionsMap["v1"].Stage)

	endpoints, err = lib.GenericDao.FindEndpointsByClusterName(TestClusterName)
	assert.Nil(err)
	assert.Equal(2, len(endpoints))
	for _, endpoint := range endpoints {
		log.InfoC(ctx, "Endpoint after rollback: %v", *endpoint)
		assert.Equal(endpoint.Address, "test-service-"+endpoint.DeploymentVersion)
	}

	respFromService, statusCode = GetFromTraceService(assert, internalGateway.Url+"/")
	assert.Equal(http.StatusNotFound, statusCode)
	respFromService, statusCode = GetFromTraceServiceWithVersion(assert, internalGateway.Url+"/", "v2")
	assert.Equal(http.StatusOK, statusCode)
	assert.Equal("test-service-v2", respFromService.ServiceName)

	// cleanup v2
	deletedVersion := deleteVersion(assert, "v2")
	assert.Equal("v2", deletedVersion.Version)
	assert.Equal(domain.CandidateStage, deletedVersion.Stage)

	versions = getVersions(assert)
	assert.Equal(1, len(versions))
	assert.Equal("v1", versions[0].Version)
	assert.Equal(domain.ActiveStage, versions[0].Stage)

	// cleanup v1 routes
	internalGateway.DeleteRoutesAndWait(assert, 60*time.Second, dto.RouteDeleteRequestV3{
		Gateways:       []string{"internal-gateway-service"},
		VirtualService: "internal-gateway-service",
		RouteDeleteRequest: dto.RouteDeleteRequest{
			Routes:  []dto.RouteDeleteItem{{"/api/v1/test-service/common"}, {"/api/v1/test-service/unique"}},
			Version: "v1",
		},
	})
}

func getVersionsMap(assert *asrt.Assertions) map[string]*domain.DeploymentVersion {
	versions := getVersions(assert)
	versionsMap := make(map[string]*domain.DeploymentVersion, len(versions))
	for _, version := range versions {
		versionsMap[version.Version] = version
	}
	return versionsMap
}

func getVersions(assert *asrt.Assertions) []*domain.DeploymentVersion {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v2/control-plane/versions", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	var versions []*domain.DeploymentVersion
	err = json.Unmarshal(bodyBytes, &versions)
	assert.Nil(err)
	return versions
}

func promoteAndWait(assert *asrt.Assertions, timeout time.Duration, version string, historySize int) []*domain.DeploymentVersion {
	versionBeforeOperation := strconv.FormatInt(time.Now().UnixNano(), 10)
	routeConfig := internalGateway.GetEnvoyRouteConfig(assert)
	if routeConfig != nil && routeConfig.VersionInfo != "" {
		versionBeforeOperation = routeConfig.VersionInfo
	}

	result := promote(assert, version, historySize)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		routeConfig := internalGateway.GetEnvoyRouteConfig(assert)
		if routeConfig != nil && routeConfig.VersionInfo != "" && versionBeforeOperation != routeConfig.VersionInfo {
			return result
		}
	}
	assert.Fail("RouteConfig was not updated in envoy before timeout exceeded")
	return nil
}

func promote(assert *asrt.Assertions, version string, historySize int) []*domain.DeploymentVersion {
	url := "http://localhost:8080/api/v2/control-plane/promote/" + version
	if historySize >= 0 {
		url = fmt.Sprintf("%s?archiveSize=%d", url, historySize)
	}
	req, _ := http.NewRequest(http.MethodPost, url, nil)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	var versions []*domain.DeploymentVersion
	err = json.Unmarshal(bodyBytes, &versions)
	assert.Nil(err)
	return versions
}

func rollbackAndWait(assert *asrt.Assertions, timeout time.Duration) []*domain.DeploymentVersion {
	versionBeforeOperation := strconv.FormatInt(time.Now().UnixNano(), 10)
	routeConfig := internalGateway.GetEnvoyRouteConfig(assert)
	if routeConfig != nil && routeConfig.VersionInfo != "" {
		versionBeforeOperation = routeConfig.VersionInfo
	}

	result := rollback(assert)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
		routeConfig := internalGateway.GetEnvoyRouteConfig(assert)
		if routeConfig != nil && routeConfig.VersionInfo != "" && versionBeforeOperation != routeConfig.VersionInfo {
			return result
		}
	}
	assert.Fail("RouteConfig was not updated in envoy before timeout exceeded")
	return nil
}

func rollback(assert *asrt.Assertions) []*domain.DeploymentVersion {
	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v2/control-plane/rollback", nil)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	var versions []*domain.DeploymentVersion
	err = json.Unmarshal(bodyBytes, &versions)
	assert.Nil(err)
	return versions
}

func deleteVersion(assert *asrt.Assertions, version string) *domain.DeploymentVersion {
	req, _ := http.NewRequest(http.MethodDelete, "http://localhost:8080/api/v2/control-plane/versions/"+version, nil)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	var dVersion domain.DeploymentVersion
	err = json.Unmarshal(bodyBytes, &dVersion)
	assert.Nil(err)
	return &dVersion
}

func requestMicroserviceVersion(assert *asrt.Assertions, microservice string) string {
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/api/v3/versions/microservices/"+microservice, nil)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)

	var dVersion map[string]string
	err = json.Unmarshal(bodyBytes, &dVersion)
	assert.Nil(err)
	return dVersion["version"]
}
