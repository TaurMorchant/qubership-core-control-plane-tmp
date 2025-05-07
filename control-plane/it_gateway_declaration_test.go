package main

import (
	"bytes"
	"github.com/netcracker/qubership-core-control-plane/lib"
	asrt "github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_IT_GatewayDeclaration(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	config := `apiVersion: nc.core.mesh/v3
kind: RouteConfiguration
metadata:
 name: test-routes-for-gw-declaration
 namespace: ''
spec:
 gateways:
   - integration-gateway
 virtualServices:
 - name: integration-gateway
   routeConfiguration:
     routes:
     - destination:
         cluster: private-gateway-service
         endpoint: private-gateway-service:8080
       rules:
       - match:
           prefix: /api/v1/integration-gateway
         prefixRewrite: /api/v1/control-plane/routes/clusters`
	applyConfig(assert, config)

	config = `apiVersion: nc.core.mesh/v3
kind: GatewayDeclaration
metadata:
 name: integration-gateway
 namespace: ''
spec:
 name: integration-gateway
 gatewayType: ingress`
	applyConfig(assert, config)

	_, status := internalGateway.SendGatewayRequest(assert, http.MethodDelete, "/api/v3/control-plane/gateways/specs",
		bytes.NewReader([]byte("{\"name\": \"integration-gateway\"}")))
	assert.Equal(http.StatusBadRequest, status)

	cluster, err := lib.GenericDao.FindClusterByName("private-gateway-service||private-gateway-service||8080")
	assert.Nil(err)
	deleteCluster(assert, cluster.Id)

	deleteVirtualService(assert, "integration-gateway", "integration-gateway")

	_, status = internalGateway.SendGatewayRequest(assert, http.MethodDelete, "/api/v3/control-plane/gateways/specs",
		bytes.NewReader([]byte("{\"name\": \"integration-gateway\"}")))
	assert.Equal(http.StatusOK, status)
}
