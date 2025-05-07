package main

import (
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	asrt "github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func Test_IT_RouteRegexp_MultipleVariables(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", true)
	defer traceSrvContainer.Purge()

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/{var1}/test-service/{var3}/common/{var2}"}}},
		},
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules:       []dto.Rule{{Match: dto.RouteMatch{Prefix: "/api/v1/{var1}/test-service/{var1}/common"}}},
		},
	)

	internalGateway.VerifyGatewayRequest(assert, http.StatusOK, "/api/v1/var1/test-service/var3/common/var2",
		"/api/v1/var1/test-service/var3/common/var2")

	// cleanup v1 routes
	internalGateway.CleanupGatewayRoutes(assert, "v1",
		"/api/v1/{var1}/test-service/{var3}/common/{var2}",
		"/api/v1/{var1}/test-service/{var1}/common")
}
