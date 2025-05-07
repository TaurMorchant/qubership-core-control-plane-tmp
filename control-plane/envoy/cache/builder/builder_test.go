package builder

import (
	v3listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	v3runtime "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	pstruct "github.com/golang/protobuf/ptypes/struct"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/builder/common"
	mock_dao "github.com/netcracker/qubership-core-control-plane/test/mock/dao"
	mock_listener "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder/listener"
	mock_routeconfig "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder/routeconfig"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
	"testing"
)

func TestNewEnvoyConfigBuilder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	props := &common.EnvoyProxyProperties{}
	facadeListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	gatewayListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	meshVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	gatewayVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	ingressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	egressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)

	builder := NewEnvoyConfigBuilder(daoMock, props, facadeListenerBuilder, gatewayListenerBuilder, meshVhBuilder, gatewayVhBuilder, ingressVhBuilder, egressVhBuilder)

	err := builder.RegisterGateway(&domain.NodeGroup{
		GatewayType: "fake",
	})
	assert.NotNil(t, err)

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        "test-gw",
		GatewayType: "",
	})
	assert.Nil(t, err)
	assert.Nil(t, builder.listenerBuilders["test-gw"])
	assert.Nil(t, builder.virtualHostBuilders["test-gw"])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        domain.PublicGateway,
		GatewayType: domain.Egress,
	})
	assert.Nil(t, err)
	assert.Equal(t, gatewayListenerBuilder, builder.listenerBuilders[domain.PublicGateway])
	assert.Equal(t, gatewayVhBuilder, builder.virtualHostBuilders[domain.PublicGateway])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        domain.PrivateGateway,
		GatewayType: domain.Egress,
	})
	assert.Nil(t, err)
	assert.Equal(t, gatewayListenerBuilder, builder.listenerBuilders[domain.PrivateGateway])
	assert.Equal(t, gatewayVhBuilder, builder.virtualHostBuilders[domain.PrivateGateway])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        domain.InternalGateway,
		GatewayType: domain.Egress,
	})
	assert.Nil(t, err)
	assert.Equal(t, gatewayListenerBuilder, builder.listenerBuilders[domain.InternalGateway])
	assert.Equal(t, gatewayVhBuilder, builder.virtualHostBuilders[domain.InternalGateway])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        "test-gw",
		GatewayType: domain.Ingress,
	})
	assert.Nil(t, err)
	assert.Equal(t, gatewayListenerBuilder, builder.listenerBuilders["test-gw"])
	assert.Equal(t, ingressVhBuilder, builder.virtualHostBuilders["test-gw"])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        "test-gw",
		GatewayType: domain.Egress,
	})
	assert.Nil(t, err)
	assert.Equal(t, facadeListenerBuilder, builder.listenerBuilders["test-gw"])
	assert.Equal(t, egressVhBuilder, builder.virtualHostBuilders["test-gw"])

	err = builder.RegisterGateway(&domain.NodeGroup{
		Name:        "test-gw",
		GatewayType: domain.Mesh,
	})
	assert.Nil(t, err)
	assert.Equal(t, facadeListenerBuilder, builder.listenerBuilders["test-gw"])
	assert.Equal(t, meshVhBuilder, builder.virtualHostBuilders["test-gw"])
}

func TestBuildRuntime(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	props := &common.EnvoyProxyProperties{Googlere2: &common.GoogleRe2Properties{
		Maxsize:  "maxsize",
		WarnSize: "warnsize",
	}}
	facadeListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	gatewayListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	meshVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	gatewayVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	ingressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	egressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)

	builder := NewEnvoyConfigBuilder(daoMock, props, facadeListenerBuilder, gatewayListenerBuilder, meshVhBuilder, gatewayVhBuilder, ingressVhBuilder, egressVhBuilder)

	actual, err := builder.BuildRuntime("", "runtime-name")
	assert.Nil(t, err)
	assert.Equal(t, v3runtime.Runtime{
		Name: "runtime-name",
		Layer: &pstruct.Struct{
			Fields: map[string]*pstruct.Value{
				"re2.max_program_size.error_level": {
					Kind: &structpb.Value_StringValue{StringValue: props.Googlere2.Maxsize},
				},
				"re2.max_program_size.warn_level": {
					Kind: &structpb.Value_StringValue{StringValue: props.Googlere2.WarnSize},
				},
			},
		},
	}, *actual)
}

func TestBuildListener(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	props := &common.EnvoyProxyProperties{}
	facadeListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	gatewayListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	meshVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	gatewayVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	ingressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	egressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)

	builder := NewEnvoyConfigBuilder(daoMock, props, facadeListenerBuilder, gatewayListenerBuilder, meshVhBuilder, gatewayVhBuilder, ingressVhBuilder, egressVhBuilder)

	listener := &domain.Listener{
		Id:                     1,
		Name:                   "test-listener",
		BindHost:               "0.0.0.0",
		BindPort:               "8080",
		RouteConfigurationName: "test-routes",
		Version:                1,
		NodeGroupId:            "test-gw",
	}

	expectedListener := &v3listener.Listener{Name: "test-listener"}

	daoMock.EXPECT().WithRTx(gomock.Any()).Return(errors.New("expected err in test"))
	actual, err := builder.BuildListener(listener, "ns-mapping", true)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	daoMock.EXPECT().WithRTx(gomock.Any()).Return(nil)
	facadeListenerBuilder.EXPECT().BuildListener(listener, "ns-mapping", true).Return(nil, errors.New("expected err in test"))
	actual, err = builder.BuildListener(listener, "ns-mapping", true)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	daoMock.EXPECT().WithRTx(gomock.Any()).Return(nil)
	facadeListenerBuilder.EXPECT().BuildListener(listener, "ns-mapping", true).Return(expectedListener, nil)
	actual, err = builder.BuildListener(listener, "ns-mapping", true)
	assert.Nil(t, err)
	assert.Equal(t, expectedListener.Name, actual.Name)

	err = builder.RegisterGateway(&domain.NodeGroup{Name: "test-gw", GatewayType: domain.Ingress})
	assert.Nil(t, err)

	daoMock.EXPECT().WithRTx(gomock.Any()).Return(nil)
	gatewayListenerBuilder.EXPECT().BuildListener(listener, "ns-mapping", true).Return(nil, errors.New("expected err in test"))
	actual, err = builder.BuildListener(listener, "ns-mapping", true)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	daoMock.EXPECT().WithRTx(gomock.Any()).Return(nil)
	gatewayListenerBuilder.EXPECT().BuildListener(listener, "ns-mapping", true).Return(expectedListener, nil)
	actual, err = builder.BuildListener(listener, "ns-mapping", true)
	assert.Nil(t, err)
	assert.Equal(t, expectedListener.Name, actual.Name)
}

func TestBuildRouteConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMock := mock_dao.NewMockDao(ctrl)
	props := &common.EnvoyProxyProperties{}
	facadeListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	gatewayListenerBuilder := mock_listener.NewMockListenerBuilder(ctrl)
	meshVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	gatewayVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	ingressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)
	egressVhBuilder := mock_routeconfig.NewMockVirtualHostBuilder(ctrl)

	builder := NewEnvoyConfigBuilder(daoMock, props, facadeListenerBuilder, gatewayListenerBuilder, meshVhBuilder, gatewayVhBuilder, ingressVhBuilder, egressVhBuilder)

	routeConfig := &domain.RouteConfiguration{Id: 1, Name: "test-routes", NodeGroupId: "test-gw"}

	daoMock.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(1)).Return(nil, errors.New("expected test err"))
	actual, err := builder.BuildRouteConfig(routeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	virtualHosts := []*domain.VirtualHost{{Id: 1, Name: "testVHost"}}
	routeConfig.VirtualHosts = virtualHosts
	daoMock.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(1)).Return(virtualHosts, nil)
	meshVhBuilder.EXPECT().BuildVirtualHosts(routeConfig).Return(nil, errors.New("expected test err"))
	routeConfig.VirtualHosts = nil
	actual, err = builder.BuildRouteConfig(routeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	daoMock.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(1)).Return(virtualHosts, nil)
	meshVhBuilder.EXPECT().BuildVirtualHosts(routeConfig).Return([]*routev3.VirtualHost{{Name: "testVHost"}}, nil)
	routeConfig.VirtualHosts = nil
	actual, err = builder.BuildRouteConfig(routeConfig)
	assert.Nil(t, err)
	assert.Equal(t, "test-routes", actual.Name)
	assert.Equal(t, 1, len(actual.VirtualHosts))
	assert.Equal(t, "testVHost", actual.VirtualHosts[0].Name)

	err = builder.RegisterGateway(&domain.NodeGroup{Name: "test-gw", GatewayType: domain.Ingress})
	assert.Nil(t, err)

	routeConfig.VirtualHosts = virtualHosts
	daoMock.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(1)).Return(virtualHosts, nil)
	ingressVhBuilder.EXPECT().BuildVirtualHosts(routeConfig).Return(nil, errors.New("expected test err"))
	routeConfig.VirtualHosts = nil
	actual, err = builder.BuildRouteConfig(routeConfig)
	assert.NotNil(t, err)
	assert.Nil(t, actual)

	daoMock.EXPECT().FindVirtualHostsByRouteConfigurationId(int32(1)).Return(virtualHosts, nil)
	ingressVhBuilder.EXPECT().BuildVirtualHosts(routeConfig).Return([]*routev3.VirtualHost{{Name: "testVHost"}}, nil)
	routeConfig.VirtualHosts = nil
	actual, err = builder.BuildRouteConfig(routeConfig)
	assert.Nil(t, err)
	assert.Equal(t, "test-routes", actual.Name)
	assert.Equal(t, 1, len(actual.VirtualHosts))
	assert.Equal(t, "testVHost", actual.VirtualHosts[0].Name)
}
