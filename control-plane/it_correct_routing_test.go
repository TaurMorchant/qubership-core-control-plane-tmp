package main

import (
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	asrt "github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_IT_RoutingTableOrderIsCorrect_AndEnvoyForwardCorrectly(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", false)
	defer traceSrvContainer.Purge()

	allowed := true
	internalGateway.RegisterRoutesV1AndWait(assert, time.Second*60, RouteEntityRequestV1{
		MicroserviceUrl: "test-service:8080",
		Routes:          &routesToRegister,
		Allowed:         &allowed,
	})

	for _, route := range routesToRegister {
		log.InfoC(ctx, "Verifying request to %v", route.From)
		if strings.Contains(route.From, "{") && strings.Contains(route.From, "}") {
			withoutVar := strings.ReplaceAll(strings.ReplaceAll(route.From, "{", ""), "}", "")
			internalGateway.VerifyGatewayRequest(assert, http.StatusOK, withoutVar, withoutVar)
		} else {
			internalGateway.VerifyGatewayRequest(assert, http.StatusOK, route.From, route.From)
		}
	}

	// cleanup v1 routes
	internalGateway.CleanupGatewayRoutes(assert, "v1", constructPrefixesFromRoutes(routesToRegister)...)
}

func Test_IT_EnvoyForwardsCorrectly_PSUPCLFRM1352(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", false)
	defer traceSrvContainer.Purge()

	lessRoutePrefix := "/api/v1/tenants/6b741a74-5d2e-4dfe-90ea-e44f38f2bf10/shopping-frontend/"
	moreRoutePrefix := "/api/v1/tenants/6b741a74-5d2e-4dfe-90ea-e44f38f2bf10/shopping-frontend"
	routes := routesToRegisterPSUPCLFRM1352
	routes = append(routes, dto.RouteEntry{From: lessRoutePrefix, To: "/"})

	expectedRegistrationResult := routes
	expectedRegistrationResult = append(expectedRegistrationResult, dto.RouteEntry{From: moreRoutePrefix, To: "/"})
	allowed := true
	req := RouteEntityRequestV1{
		MicroserviceUrl: "test-service:8080",
		Routes:          &routes,
		Allowed:         &allowed,
	}
	internalGateway.RegisterRoutesV1AndWait(assert, time.Second*60, req)

	for _, route := range expectedRegistrationResult {
		actualRewrite := route.From
		if len(route.To) > 0 {
			actualRewrite = route.To
		}
		log.InfoC(ctx, "Verifying request to %v", route.From)
		internalGateway.VerifyGatewayRequest(assert, http.StatusOK, actualRewrite, route.From)
	}
	internalGateway.CleanupGatewayRoutes(assert, "v1", constructPrefixesFromRoutes(routes)...)
}

var (
	routesToRegisterPSUPCLFRM1352 = []dto.RouteEntry{
		{
			From: "/api/v1/paas-mediation/namespaces/cloudbss311-platform-core-support-dev3/configmaps/bg-version/",
			To:   "/api/v1/namespaces/cloudbss311-platform-core-support-dev3/configmaps/bg-version/",
		},
		{
			From: "/api/v1/paas-mediation/namespaces/cloudbss311-platform-core-support-dev3/configmaps/bg-version",
			To:   "/api/v1/namespaces/cloudbss311-platform-core-support-dev3/configmaps/bg-version",
		},
		{
			From: "/api/v3/tenant-manager/suspend/deactivate-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v4/tenant-manager/suspend/deactivate-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v3/tenant-manager/suspend/deactivate-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v4/tenant-manager/suspend/deactivate-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v3/tenant-manager/activate/create-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v4/tenant-manager/activate/create-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v4/tenant-manager/resume/restore-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v3/tenant-manager/activate/create-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v3/tenant-manager/resume/restore-os-tenant-alias-routes/rollback/{var}",
		},
		{
			From: "/api/v4/tenant-manager/activate/create-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v3/tenant-manager/resume/restore-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v4/tenant-manager/resume/restore-os-tenant-alias-routes/perform/{var}",
		},
		{
			From: "/api/v4/tenant-manager/activate/finalize-tenant-activation/rollback/{var}",
		},
	}

	routesToRegister = []dto.RouteEntry{
		{
			From: "/api/v1/ext-frontend-api/customers/import/catalog/random",
		},
		{
			From: "/api/v1/ext-frontend-api/customers/import/{customer_id}/name",
		},
		{
			From: "/api/v1/ext-frontend-api/customers/import/{customer_id}",
		},
		{
			From: "/api/v1/ext-frontend-api/customers/{customer_id}",
		},
		{
			From: "/api/v1/ext-frontend-api/customers/import",
		},
		{
			From: "/api/v1/ext-frontend-api/{resource_id}",
		},
		{
			From: "/api/v1/ext-frontend-api/{resource_id}/version",
		},
		{
			From: "/api/v1/ext-frontend-api/{resource_id}/version/{var}",
		},
		{
			From: "/api/v1/ext-frontend-api/customers",
		},
		{
			From: "/api/v1/ext-frontend-api/method",
		},
		{
			From: "/api/v1/ext-frontend-api",
		},
		{
			From: "/api/v1/ext-frontend-api/",
		},
	}
)
