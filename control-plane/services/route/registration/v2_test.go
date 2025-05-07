package registration

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"reflect"
	"testing"
)

func Test_v2RequestProcessor_ProcessRequestV2(t *testing.T) {
	t.SkipNow() // TODO to implement this test you need write Equals for ProcessedRequest or get rid of pointer entities
	type args struct {
		ctx           context.Context
		nodeGroupName string
		regRequests   []dto.RouteRegistrationRequest
		activeVersion string
	}
	tests := []struct {
		name string
		args args
		want ProcessedRequest
	}{
		{
			name: "One route",
			args: args{
				ctx:           context.Background(),
				nodeGroupName: "private-gateway-service",
				regRequests: []dto.RouteRegistrationRequest{
					{
						Cluster: "tenant-manager",
						Routes: []dto.RouteItem{
							{
								Prefix: "/api/v1/tenant-manager/info",
							},
						},
						Endpoint:  "http://tenant-manager-v1:8080",
						Allowed:   newTrue(),
						Namespace: "",
						Version:   "v1",
					},
				},
				activeVersion: "v1",
			},
			want: ProcessedRequest{
				NodeGroups: []domain.NodeGroup{{Name: "private-gateway-service"}},
				Listeners: []domain.Listener{
					{
						Name:                   "private-gateway-service-listener",
						BindHost:               "0.0.0.0",
						BindPort:               "8080",
						RouteConfigurationName: "private-gateway-service-routes",
						NodeGroupId:            "private-gateway-service",
					},
				},
				Clusters: []domain.Cluster{
					{
						Name:          "tenant-manager||tenant-manager||8080",
						LbPolicy:      "STRICT_DNS",
						DiscoveryType: "LEAST_REQUEST",
						HttpVersion:   wrapInt32(1),
						Endpoints: []*domain.Endpoint{
							{
								Address:                  "tenant-manager",
								Port:                     8080,
								DeploymentVersion:        "v1",
								InitialDeploymentVersion: "v1",
								HashPolicies:             nil,
							},
						},
					},
				},
				ClusterNodeGroups: map[string][]string{
					"tenant-manager||tenant-manager||8080": {"private-gateway-service"},
				},
				RouteConfigurations: []domain.RouteConfiguration{
					{
						Name:        "private-gateway-service-routes",
						NodeGroupId: "private-gateway-service",
						VirtualHosts: []*domain.VirtualHost{
							{
								Name: "private-gateway-service-routes",
								Routes: []*domain.Route{
									{
										Prefix:            "/api/v1/tenant-manager/info",
										ClusterName:       "tenant-manager||tenant-manager||8080",
										DeploymentVersion: "v1",
									},
								},
								Domains: []*domain.VirtualHostDomain{{Domain: "*"}},
							},
						},
					},
				},
				GroupedRoutes:      nil,
				DeploymentVersions: []string{"v1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := V2RequestProcessor{}
			if got, _ := p.ProcessRequestV2(tt.args.ctx, tt.args.nodeGroupName, tt.args.regRequests, tt.args.activeVersion); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessRequestV2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

func wrapInt32(val int32) *int32 {
	return &val
}
