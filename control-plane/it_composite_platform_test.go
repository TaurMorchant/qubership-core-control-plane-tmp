package main

import (
	"encoding/json"
	"github.com/netcracker/qubership-core-control-plane/composite"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	asrt "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func Test_IT_CompositePlatformApi(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	tenantManager.clear()
	defer tenantManager.clear()

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{},
	})

	resp, err := sendCloudAdminRequest(assert, http.MethodPost, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns1", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{"test-ns1"},
	})
	tenantManager.verifyRequests(t, http.MethodPost, "/api/v4/tenant-manager/tenants/default/sync", 1)

	resp, err = sendCloudAdminRequest(assert, http.MethodPost, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns1", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{"test-ns1"},
	})
	tenantManager.verifyRequests(t, http.MethodPost, "/api/v4/tenant-manager/tenants/default/sync", 2)

	resp, err = sendCloudAdminRequest(assert, http.MethodPost, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns2", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{"test-ns1", "test-ns2"},
	})
	tenantManager.verifyRequests(t, http.MethodPost, "/api/v4/tenant-manager/tenants/default/sync", 3)

	resp, err = sendCloudAdminRequest(assert, http.MethodDelete, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns1", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{"test-ns2"},
	})

	resp, err = sendCloudAdminRequest(assert, http.MethodDelete, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns2", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{},
	})

	resp, err = sendCloudAdminRequest(assert, http.MethodDelete, "http://localhost:8080/api/v3/composite-platform/namespaces/test-ns2", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{},
	})
}

func Test_IT_CompositePlatformBaselineMode(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	verifyCompositeStructure(assert, composite.Structure{
		Baseline:   "local",
		Satellites: []string{},
	})

	traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", true)
	defer traceSrvContainer.Purge()

	const routePath = "/api/v1/composite-test-service/baseline-route"
	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules: []dto.Rule{
				{Match: dto.RouteMatch{Prefix: routePath}},
			},
		},
	)

	headers := map[string][]string{"X-Token-Signature": {"test-signature"}}
	resp, statusCode := GetFromTraceServiceWithHeaders(assert, internalGateway.Url+routePath, headers)
	assert.Equal(http.StatusOK, statusCode)
	if resp == nil {
		log.InfoC(ctx, "Didn't receive TraceResponse; status code: %d", statusCode)
	} else {
		log.InfoC(ctx, "Trace service response: %v", resp)
		assert.Equal(routePath, resp.Path)
		// verify request header X-Token-Signature removed in baseline
		assert.Equal("", resp.Headers.Get("X-Token-Signature"))
	}

	// cleanup v1 routes
	internalGateway.CleanupGatewayRoutes(assert, "v1",
		"/api/v1/composite-test-service/baseline-route")
}

func verifyCompositeStructure(assert *asrt.Assertions, expected composite.Structure) {
	resp, err := sendCloudAdminRequest(assert, http.MethodGet, "http://localhost:8080/api/v3/composite-platform/namespaces", nil)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	var compositeStructure composite.Structure
	err = json.Unmarshal(body, &compositeStructure)
	assert.Nil(err)
	assert.Equal(expected, compositeStructure)
}
