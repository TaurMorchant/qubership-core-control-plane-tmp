package main

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	asrt "github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

type statusesMap struct {
	statusesByReq map[string]map[int]int
}

func (m *statusesMap) save(method, path string, status int) {
	reqKey := fmt.Sprintf("%s#%s", method, path)
	if m.statusesByReq == nil {
		m.statusesByReq = make(map[string]map[int]int)
	}
	if _, exists := m.statusesByReq[reqKey]; !exists {
		m.statusesByReq[reqKey] = make(map[int]int)
	}
	m.statusesByReq[reqKey][status]++
}

func (m *statusesMap) getOccurrences(method, path string, status int) int {
	reqKey := fmt.Sprintf("%s#%s", method, path)
	return m.statusesByReq[reqKey][status]
}

func doTestRequest(assert *asrt.Assertions, method, path string, statusesMap *statusesMap) {
	_, status := internalGateway.SendGatewayRequest(assert, method, path, nil)
	statusesMap.save(method, path, status)
}

func Test_IT_RateLimit(t *testing.T) {
	skipTestIfDockerDisabled(t)
	assert := asrt.New(t)

	traceSrvContainer := createTraceServiceContainer(TestCluster, "v1", true)
	defer traceSrvContainer.Purge()

	registerRateLimit(assert, &dto.RateLimit{Name: "getOrdersAndCustomersRateLimit", LimitRequestPerSecond: 3})
	registerRateLimit(assert, &dto.RateLimit{Name: "metricsRateLimit", LimitRequestPerSecond: 1})

	internalGateway.RegisterRoutesAndWait(
		assert,
		60*time.Second,
		"v1",
		dto.RouteV3{
			Destination: dto.RouteDestination{Cluster: TestCluster, Endpoint: TestEndpointV1},
			Rules: []dto.Rule{
				{Match: dto.RouteMatch{Prefix: "/api/orders"}, RateLimit: "getOrdersAndCustomersRateLimit"},
				{Match: dto.RouteMatch{Prefix: "/api/customers"}, RateLimit: "getOrdersAndCustomersRateLimit"},
				{Match: dto.RouteMatch{Prefix: "/api/products"}},
				{Match: dto.RouteMatch{Prefix: "/api/metrics"}, RateLimit: "metricsRateLimit"},
			},
		},
	)

	statusesMap := &statusesMap{}

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		doTestRequest(assert, http.MethodGet, "/api/orders", statusesMap)
		doTestRequest(assert, http.MethodPost, "/api/orders", statusesMap)
		doTestRequest(assert, http.MethodGet, "/api/customers", statusesMap)
		doTestRequest(assert, http.MethodPost, "/api/customers", statusesMap)
		doTestRequest(assert, http.MethodGet, "/api/products", statusesMap)
		doTestRequest(assert, http.MethodPost, "/api/products", statusesMap)
		doTestRequest(assert, http.MethodGet, "/api/metrics", statusesMap)
		doTestRequest(assert, http.MethodPost, "/api/metrics", statusesMap)

	}
	for reqKey, statuses := range statusesMap.statusesByReq {
		for status, occurrenceNum := range statuses {
			log.DebugC(ctx, "Status %d occurred %d times for %s", status, occurrenceNum, reqKey)
		}
	}
	assert.GreaterOrEqual(3,
		statusesMap.getOccurrences(http.MethodGet, "/api/metrics", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/metrics", 200))
	assert.LessOrEqual(1,
		statusesMap.getOccurrences(http.MethodGet, "/api/metrics", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/metrics", 200))
	assert.LessOrEqual(10,
		statusesMap.getOccurrences(http.MethodGet, "/api/products", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/products", 200))
	assert.Equal(0,
		statusesMap.getOccurrences(http.MethodGet, "/api/products", 429)+
			statusesMap.getOccurrences(http.MethodPost, "/api/products", 429))
	assert.GreaterOrEqual(9,
		statusesMap.getOccurrences(http.MethodGet, "/api/orders", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/orders", 200)+
			statusesMap.getOccurrences(http.MethodGet, "/api/customers", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/customers", 200))
	assert.LessOrEqual(3,
		statusesMap.getOccurrences(http.MethodGet, "/api/orders", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/orders", 200)+
			statusesMap.getOccurrences(http.MethodGet, "/api/customers", 200)+
			statusesMap.getOccurrences(http.MethodPost, "/api/customers", 200))

	// cleanup v1 routes
	internalGateway.CleanupGatewayRoutes(assert, "v1",
		"/api/orders",
		"/api/customers",
		"/api/products",
		"/api/metrics")
}
