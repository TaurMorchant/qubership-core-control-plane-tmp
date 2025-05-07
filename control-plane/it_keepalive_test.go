package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_IT_KeepAlive(t *testing.T) {
	skipTestIfDockerDisabled(t)
	asrt := assert.New(t)

	config := `apiVersion: nc.core.mesh/v3
kind: RouteConfiguration
metadata:
 name: tcp-server-routes
 namespace: ''
spec:
 gateways:
   - internal-gateway-service
 virtualServices:
 - name: internal-gateway-service
   routeConfiguration:
     routes:
     - destination:
         cluster: tcp-keepalive-test
         endpoint: tcp-keepalive-test:8080
         tcpKeepalive:
           probes: 4
           time: 22
           interval: 12
       rules:
       - match:
           prefix: /tcp-keepalive-test`

	internalGateway.ApplyConfigAndWait(asrt, 60*time.Second, config)

	cluster := getClusterFromDumpWithRetry(asrt, "tcp-keepalive-test||tcp-keepalive-test||8080", 60*time.Second)
	asrt.NotNil(cluster)

	tcpKeepalive := cluster.UpstreamConnectionOptions.TcpKeepalive
	asrt.NotNil(tcpKeepalive)

	asrt.Equal(4, *tcpKeepalive.KeepaliveProbes)
	asrt.Equal(22, *tcpKeepalive.KeepaliveTime)
	asrt.Equal(12, *tcpKeepalive.KeepaliveInterval)

	json := `{
  "clusterKey": "tcp-keepalive-test||tcp-keepalive-test||8080",
  "tcpKeepalive": {
    "probes": 3,
    "time": 30,
    "interval": 10
  }
}  `
	b := strings.NewReader(json)

	req, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/api/v3/clusters/tcp-keepalive", b)

	resp, err := http.DefaultClient.Do(req)
	asrt.Nil(err)
	asrt.Equal(http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	time.Sleep(8 * time.Second)

	cluster = getClusterFromDumpWithRetry(asrt, "tcp-keepalive-test||tcp-keepalive-test||8080", 60*time.Second)
	asrt.NotNil(cluster)

	tcpKeepalive = cluster.UpstreamConnectionOptions.TcpKeepalive
	asrt.NotNil(tcpKeepalive)

	asrt.Equal(3, *tcpKeepalive.KeepaliveProbes)
	asrt.Equal(30, *tcpKeepalive.KeepaliveTime)
	asrt.Equal(10, *tcpKeepalive.KeepaliveInterval)
}

func getClusterFromDumpWithRetry(asrt *assert.Assertions, clusterKey string, timeout time.Duration) *EnvoyCluster {
	deadline := time.Now().Add(timeout)

	var cluster *EnvoyCluster

	for cluster == nil && time.Now().Before(deadline) {
		configDump := internalGateway.getEnvoyConfig(asrt)
		cluster = configDump.FindClusterByName(clusterKey)
	}
	return cluster
}
