package ui

import (
	"context"
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	asrt "github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestV3Service_GetAllSimplifiedRouteConfigs(t *testing.T) {
	assert := asrt.New(t)
	ds := newMockDs()
	s := &V3Service{
		ds:         ds,
		relService: &entityServiceWrapper{},
		sorter:     routeSorterMock{},
	}

	ds.mockData["FindAllRouteConfigs"] = func(args ...interface{}) (interface{}, error) {
		return ([]*domain.RouteConfiguration)(nil), errors.New("can't find all route configs")
	}
	uiRc, err := s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.Nil(uiRc)
	assert.Error(err)

	ds.mockData["FindAllRouteConfigs"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.RouteConfiguration{}, nil
	}
	ds.mockData["FindVirtualHostsByRouteConfigurationId"] = func(args ...interface{}) (interface{}, error) {
		return ([]*domain.VirtualHost)(nil), errors.New("can't find all virtualHosts")
	}
	uiRc, err = s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.NotNil(uiRc)
	assert.Empty(uiRc)
	assert.Nil(err)

	ds.mockData["FindVirtualHostsByRouteConfigurationId"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.VirtualHost{}, nil
	}
	ds.mockData["FindAllDeploymentVersions"] = func(args ...interface{}) (interface{}, error) {
		return ([]*domain.DeploymentVersion)(nil), errors.New("can't find deployment versions")
	}
	uiRc, err = s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.NotNil(uiRc)
	assert.Empty(uiRc)
	assert.Nil(err)

	ds.mockData["FindAllDeploymentVersions"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.DeploymentVersion{}, nil
	}
	uiRc, err = s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.NotNil(uiRc)
	assert.Nil(err)
	assert.Empty(uiRc)

	ds.mockData["FindAllRouteConfigs"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.RouteConfiguration{{
			Id:          2,
			NodeGroupId: "test-node-group",
		}}, nil
	}
	ds.mockData["FindVirtualHostsByRouteConfigurationId"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.VirtualHost{{
			Id:                   5,
			Name:                 "test-virtual-host",
			RouteConfigurationId: 2,
		}}, nil
	}
	ds.mockData["FindAllDeploymentVersions"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.DeploymentVersion{{Version: "v2", Stage: "Active"}}, nil
	}
	ds.mockData["FindRoutesByVirtualHostIdAndDeploymentVersion"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.Route{}, nil
	}
	uiRc, err = s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.NotNil(uiRc)
	assert.Nil(err)
	assert.Equal([]SimplifiedRouteConfig{{
		NodeGroup:    "test-node-group",
		VirtualHosts: []VirtualHost{{Id: 5, Name: "test-virtual-host", DeploymentVersions: []DeploymentVersion{}}},
	}}, uiRc)

	ds.mockData["FindAllRouteConfigs"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.RouteConfiguration{{
			NodeGroupId: "test-node-group",
		}}, nil
	}
	ds.mockData["FindRoutesByVirtualHostIdAndDeploymentVersion"] = func(args ...interface{}) (interface{}, error) {
		return []*domain.Route{{}}, nil
	}
	uiRc, err = s.GetAllSimplifiedRouteConfigs(context.Background())
	assert.NotNil(uiRc)
	assert.Nil(err)
	assert.Equal([]SimplifiedRouteConfig{{
		NodeGroup: "test-node-group",
		VirtualHosts: []VirtualHost{{
			Id:                 5,
			Name:               "test-virtual-host",
			DeploymentVersions: []DeploymentVersion{{Name: "v2", Stage: "Active"}},
		}},
	}}, uiRc)

}

func TestV3Service_MakeSearchRoutesPage(t *testing.T) {
	type fields struct {
		ds         *mockDs
		relService RelationsService
		sorter     RouteSorter
	}
	type args struct {
		ctx    context.Context
		params SearchRoutesParameters
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PageRoutes
		wantErr bool
	}{
		{name: "Test zero data", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          0,
					Page:          0,
					Search:        "",
				},
			},
			want:    PageRoutes{},
			wantErr: true},
		{name: "Test one route per page", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          1,
					Page:          1,
					Search:        "",
				},
			},
			want: PageRoutes{NodeGroup: "private-node-group", VirtualHostName: "virtual-host", VersionName: "v1", VersionStage: "ACTIVE",
				TotalCount: 10, Routes: []Route{{Allowed: true}}},
			wantErr: false},
		{name: "Test one route per page next page", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          1,
					Page:          2,
					Search:        "",
				},
			},
			want: PageRoutes{NodeGroup: "private-node-group", VirtualHostName: "virtual-host", VersionName: "v1", VersionStage: "ACTIVE",
				TotalCount: 10, Routes: []Route{{Allowed: false}}},
			wantErr: false},
		{name: "Test last route", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          1,
					Page:          10,
					Search:        "",
				},
			},
			want: PageRoutes{NodeGroup: "private-node-group", VirtualHostName: "virtual-host", VersionName: "v1", VersionStage: "ACTIVE",
				TotalCount: 10, Routes: []Route{{Match: "/lastRoute", Allowed: true}}},
			wantErr: false},
		{name: "Test too large size and page", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          11,
					Page:          20,
					Search:        "",
				},
			},
			want: PageRoutes{NodeGroup: "private-node-group", VirtualHostName: "virtual-host", VersionName: "v1", VersionStage: "ACTIVE",
				TotalCount: 10, Routes: []Route{
					{Allowed: true}, {}, {Allowed: true}, {Allowed: true}, {Allowed: true}, {Allowed: true}, {Allowed: true},
					{Allowed: true}, {Allowed: true}, {Match: "/lastRoute", Allowed: true}}},
			wantErr: false},
		{name: "Test last items", fields: fields{ds: newMockDs(), relService: &entityServiceWrapper{}, sorter: routeSorterMock{}},
			args: args{
				ctx: context.Background(),
				params: SearchRoutesParameters{
					VirtualHostId: 1,
					Version:       "v1",
					Size:          8,
					Page:          2,
					Search:        "",
				},
			},
			want: PageRoutes{NodeGroup: "private-node-group", VirtualHostName: "virtual-host", VersionName: "v1", VersionStage: "ACTIVE",
				TotalCount: 10, Routes: []Route{{Allowed: true}, {Match: "/lastRoute", Allowed: true}}},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &V3Service{
				ds:         tt.fields.ds,
				relService: tt.fields.relService,
				sorter:     tt.fields.sorter,
			}
			tt.fields.ds.mockData["FindRoutesByVirtualHostIdAndDeploymentVersion"] = func(args ...interface{}) (interface{}, error) {
				vHostId := args[0].(int32)
				version := args[1].(string)
				if vHostId == 1 && version == "v1" {
					return []*domain.Route{{}, {DirectResponseCode: 404}, {}, {}, {}, {}, {}, {}, {}, {Prefix: "/lastRoute"}}, nil
				}
				return []*domain.Route{}, nil
			}
			tt.fields.ds.mockData["FindVirtualHostById"] = func(args ...interface{}) (interface{}, error) {
				vhId := args[0].(int32)
				return &domain.VirtualHost{Id: vhId, Name: "virtual-host"}, nil
			}
			tt.fields.ds.mockData["FindDeploymentVersion"] = func(args ...interface{}) (interface{}, error) {
				dVersion := args[0].(string)
				return &domain.DeploymentVersion{Version: dVersion, Stage: "ACTIVE"}, nil
			}
			tt.fields.ds.mockData["FindRouteConfigById"] = func(args ...interface{}) (interface{}, error) {
				routeConfigId := args[0].(int32)
				return &domain.RouteConfiguration{Id: routeConfigId, NodeGroupId: "private-node-group"}, nil
			}
			got, err := s.GetRoutesPage(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoutesPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRoutesPage() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}

type mockDs struct {
	mockData map[string]func(args ...interface{}) (interface{}, error)
}

type routeSorterMock struct{}

func (r routeSorterMock) Sort(routes []*domain.Route) []*domain.Route {
	return routes
}

func newMockDs() *mockDs {
	return &mockDs{mockData: map[string]func(args ...interface{}) (interface{}, error){
		"FindAllRouteConfigs": func(args ...interface{}) (interface{}, error) {
			return []*domain.RouteConfiguration{}, nil
		},
		"FindVirtualHostsByRouteConfigurationId": func(args ...interface{}) (interface{}, error) {
			return []*domain.VirtualHost{}, nil
		},
		"FindAllDeploymentVersions": func(args ...interface{}) (interface{}, error) {
			return []*domain.DeploymentVersion{}, nil
		},
		"FindRoutesByVirtualHostIdAndDeploymentVersion": func(args ...interface{}) (interface{}, error) {
			return []*domain.Route{}, nil
		},
		"FindVirtualHostById": func(args ...interface{}) (interface{}, error) {
			return new(domain.VirtualHost), nil
		},
		"FindDeploymentVersion": func(args ...interface{}) (interface{}, error) {
			return new(domain.DeploymentVersion), nil
		},
		"FindRouteConfigById": func(args ...interface{}) (interface{}, error) {
			return new(domain.RouteConfiguration), nil
		},
	}}
}

func (m *mockDs) FindAllRouteConfigs() ([]*domain.RouteConfiguration, error) {
	stubFunc := m.mockData["FindAllRouteConfigs"]
	result, err := stubFunc(nil)
	return result.([]*domain.RouteConfiguration), err
}

func (m *mockDs) FindVirtualHostsByRouteConfigurationId(configId int32) ([]*domain.VirtualHost, error) {
	stubFunc := m.mockData["FindVirtualHostsByRouteConfigurationId"]
	result, err := stubFunc(configId)
	return result.([]*domain.VirtualHost), err
}

func (m *mockDs) FindAllDeploymentVersions() ([]*domain.DeploymentVersion, error) {
	stubFunc := m.mockData["FindAllDeploymentVersions"]
	result, err := stubFunc(nil)
	return result.([]*domain.DeploymentVersion), err
}

func (m *mockDs) FindRoutesByVirtualHostIdAndDeploymentVersion(vhId int32, version string) ([]*domain.Route, error) {
	stubFunc := m.mockData["FindRoutesByVirtualHostIdAndDeploymentVersion"]
	result, err := stubFunc(vhId, version)
	return result.([]*domain.Route), err
}

func (m *mockDs) FindVirtualHostById(vhId int32) (*domain.VirtualHost, error) {
	stubFunc := m.mockData["FindVirtualHostById"]
	result, err := stubFunc(vhId)
	return result.(*domain.VirtualHost), err
}

func (m *mockDs) FindDeploymentVersion(version string) (*domain.DeploymentVersion, error) {
	stubFunc := m.mockData["FindDeploymentVersion"]
	result, err := stubFunc(version)
	return result.(*domain.DeploymentVersion), err
}

func (m *mockDs) FindRouteConfigById(routeConfigurationId int32) (*domain.RouteConfiguration, error) {
	stubFunc := m.mockData["FindRouteConfigById"]
	result, err := stubFunc(routeConfigurationId)
	return result.(*domain.RouteConfiguration), err
}

func (m *mockDs) FindRouteByUuid(uuid string) (*domain.Route, error) {
	panic("implement me")
}

func (m *mockDs) FindClusterByName(key string) (*domain.Cluster, error) {
	panic("implement me")
}

func (m *mockDs) FindEndpointsByClusterIdAndDeploymentVersion(clusterId int32, dVersions *domain.DeploymentVersion) ([]*domain.Endpoint, error) {
	panic("implement me")
}

func (m *mockDs) FindHeaderMatcherByRouteId(routeId int32) ([]*domain.HeaderMatcher, error) {
	return nil, nil
}
