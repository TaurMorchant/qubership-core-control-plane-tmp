package registration

import (
	"fmt"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/tlsmode"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"sync/atomic"
	"testing"
)

const namespace = "test-namespace"

func Test_v3RequestProcessor_processVirtualServiceRequestV3(t *testing.T) {
	tests := []struct {
		name         string
		Endpoint     string
		tlsEnabled   string
		expectedPort int32
	}{
		{
			name:         "testTlsDisabledTlsAndHttpEndpoint",
			Endpoint:     "http://test-http.service:8080",
			tlsEnabled:   "false",
			expectedPort: 8080,
		},
		{
			name:         "testTlsEnabledTlsAndHttpEndpoint",
			Endpoint:     "http://test-http.service:8080",
			tlsEnabled:   "true",
			expectedPort: 8080,
		},
		{
			name:         "testTlsEnabledTlsAndHttpsEndpoint",
			Endpoint:     "https://test-http.service:8443",
			tlsEnabled:   "true",
			expectedPort: 8443,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewV3RequestProcessor(getDao())
			virtualServicesRequestV3 := dto.VirtualService{
				RouteConfiguration: dto.RouteConfig{
					Routes: []dto.RouteV3{
						{
							Destination: dto.RouteDestination{
								Endpoint: tt.Endpoint,
								Cluster:  "test-cluster",
								CircuitBreaker: dto.CircuitBreaker{
									Threshold: dto.Threshold{
										MaxConnections: 2,
									},
								},
							},
							Rules: []dto.Rule{{
								Match: dto.RouteMatch{
									Prefix:         "/api/v1/test-https/resource",
									HeaderMatchers: []dto.HeaderMatcher{},
								},
							}},
						},
					},
				},
			}

			_ = os.Setenv("INTERNAL_TLS_ENABLED", tt.tlsEnabled)
			defer disableTls()
			configloader.Init(configloader.EnvPropertySource())
			tlsmode.SetUpTlsProperties()

			processedRequest, err := p.ProcessVirtualServiceRequestV3(virtualServicesRequestV3, "", "v1")
			assert.Nil(t, err)
			assert.NotNil(t, processedRequest)
			assert.Equal(t, fmt.Sprintf("%s%d", "test-cluster||test-cluster||", tt.expectedPort), processedRequest.Clusters[0].Name)
			assert.Equal(t, tt.expectedPort, processedRequest.Clusters[0].Endpoints[0].Port)
		})
	}
}

func disableTls() {
	os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()
}

func Test_v3GenerateDomains(t *testing.T) {
	tests := []struct {
		name   string
		port   int
		hosts  []string
		result []string
	}{
		{
			name:   "test1234Port",
			port:   1234,
			hosts:  []string{"host1"},
			result: []string{"host1:1234", "host1.namespace:1234", "host1.namespace.svc:1234", "host1.namespace.svc.cluster.local:1234"},
		},
		{
			name:   "test0Port",
			port:   0,
			hosts:  []string{"host1"},
			result: []string{"host1:8080", "host1.namespace:8080", "host1.namespace.svc:8080", "host1.namespace.svc.cluster.local:8080"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v3 := NewV3RequestProcessor(getDao())
			result := v3.GenerateDomains("namespace", tt.port, tt.hosts)
			assert.Equal(t, len(tt.result), len(result))
			for _, host := range result {
				assert.True(t, contains(tt.result, host))
			}
		})
	}
}

func contains(hosts []string, host string) bool {
	for _, currentHost := range hosts {
		if currentHost == host {
			return true
		}
	}

	return false
}

func Test_v3RequestProcessor_processGateway_TLS_endpoints(t *testing.T) {
	tests := []struct {
		name                   string
		endpoint               string
		routeHeaderMatchers    []dto.HeaderMatcher
		want                   bool
		expectedHeaderMatchers []*domain.HeaderMatcher
	}{
		{
			name:                   "http scheme",
			endpoint:               "http://test-http.service",
			routeHeaderMatchers:    []dto.HeaderMatcher{},
			want:                   false,
			expectedHeaderMatchers: []*domain.HeaderMatcher{{Name: "namespace", ExactMatch: namespace, Version: 1}},
		},
		{
			name:                   "80 port",
			endpoint:               "test-http.service:80",
			routeHeaderMatchers:    nil,
			want:                   false,
			expectedHeaderMatchers: []*domain.HeaderMatcher{{Name: "namespace", ExactMatch: namespace, Version: 1}},
		},
		{
			name:                   "https scheme",
			endpoint:               "https://test-https.service",
			routeHeaderMatchers:    nil,
			want:                   true,
			expectedHeaderMatchers: []*domain.HeaderMatcher{{Name: "namespace", ExactMatch: namespace, Version: 1}},
		},
		{
			name:                   "443 port",
			endpoint:               "test-https.service:443",
			routeHeaderMatchers:    []dto.HeaderMatcher{{Name: "My-Header", ExactMatch: "val"}},
			want:                   true,
			expectedHeaderMatchers: []*domain.HeaderMatcher{{Name: "My-Header", ExactMatch: "val", Version: 1}, {Name: "namespace", ExactMatch: namespace, Version: 1}},
		},
		{
			name:                   "already has scheme matcher",
			endpoint:               "test-https.service:443",
			routeHeaderMatchers:    []dto.HeaderMatcher{{Name: ":scheme", ExactMatch: "https"}},
			want:                   true,
			expectedHeaderMatchers: []*domain.HeaderMatcher{{Name: ":scheme", ExactMatch: "https", Version: 1}, {Name: "namespace", ExactMatch: namespace, Version: 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewV3RequestProcessor(getDao())
			virtualServices := []dto.VirtualService{{RouteConfiguration: dto.RouteConfig{
				Routes: []dto.RouteV3{
					{
						Destination: dto.RouteDestination{
							Endpoint: tt.endpoint,
							Cluster:  "test-https",
						},
						Rules: []dto.Rule{{
							Match: dto.RouteMatch{
								Prefix:         "/api/v1/test-https/resource",
								HeaderMatchers: tt.routeHeaderMatchers,
							},
						}},
					},
				}}},
			}
			processedRequest, _ := p.processGateway("test", namespace, "1", false, 8080, virtualServices...)
			clusterTlsConfig := processedRequest.ClusterTlsConfig
			msAddress := msaddr.NewMicroserviceAddress(tt.endpoint, namespace)
			_, got := clusterTlsConfig[clusterkey.DefaultClusterKeyGenerator.GenerateKey("test-https", msAddress)]
			if got != tt.want {
				t.Errorf("processGateway() = %v, want %v", got, tt.want)
			}
			routeVerified := false
			err := processedRequest.GroupedRoutes.ForEachGroup(func(namespace, clusterKey, deploymentVersion string, routes []*domain.Route) error {
				assert.Equal(t, 1, len(routes))
				headerMatchers := routes[0].HeaderMatchers
				assert.True(t, headerMatchersEqual(tt.expectedHeaderMatchers, headerMatchers))
				routeVerified = true
				return nil
			})
			assert.Nil(t, err)
			assert.True(t, routeVerified)
		})
	}
}

func headerMatchersEqual(expected, actual []*domain.HeaderMatcher) bool {
	if len(expected) != len(actual) {
		return false
	}
	for _, matcher := range actual {
		matcherFound := false
		for _, another := range expected {
			if matcher.Equals(another) {
				matcherFound = true
				break
			}
		}
		if !matcherFound {
			return false
		}
	}
	for _, matcher := range expected {
		matcherFound := false
		for _, another := range actual {
			if matcher.Equals(another) {
				matcherFound = true
				break
			}
		}
		if !matcherFound {
			return false
		}
	}
	return true
}

func getDao() dao.Dao {
	callback := func([]memdb.Change) error {
		return nil
	}
	return dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, []func([]memdb.Change) error{callback})
}

type idGeneratorMock struct {
	counter int32
}

func (g *idGeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}
