package route

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
	"github.com/stretchr/testify/assert"
	"sync/atomic"

	//"control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	//"control-plane/services/entity"
	//"control-plane/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/registration"
	//"control-plane/services/routingmode"
	//"github.com/stretchr/testify/assert"
	"testing"
)

func TestDeleteDomains(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return regSrv.RegisterRoutes(context.Background(), dao, request)
	})
	assert.Nil(t, err)

	registrationServiceContext := regSrv.WithContext(context.Background(), daoMock)

	vHostId := request.RouteConfigurations[0].VirtualHosts[0].Routes[0].VirtualHostId

	vHDomains, err := registrationServiceContext.DeleteDomains(vHostId, "")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(vHDomains))

	vHDomains, err = registrationServiceContext.DeleteDomains(vHostId, request.RouteConfigurations[0].VirtualHosts[0].Domains[0].Domain)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vHDomains))
	assert.Equal(t, request.RouteConfigurations[0].VirtualHosts[0].Domains[0], vHDomains[0])
}

func TestDeleteRoutesByNodeGroup(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return regSrv.RegisterRoutes(context.Background(), dao, request)
	})
	assert.Nil(t, err)

	registrationServiceContext := regSrv.WithContext(context.Background(), daoMock)

	nodeGroup := request.RouteConfigurations[0].NodeGroupId
	reqNamespace := ""
	reqVersion := ""
	rawPrefixes := ""

	routes, err := registrationServiceContext.DeleteRoutes(nodeGroup, reqNamespace, reqVersion, rawPrefixes)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, request.RouteConfigurations[0].VirtualHosts[0].Routes[0].Uuid, routes[0].Uuid)
}

func TestDeleteRoutesByRawPrefixNamespaceVersion(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return regSrv.RegisterRoutes(context.Background(), dao, request)
	})
	assert.Nil(t, err)

	registrationServiceContext := regSrv.WithContext(context.Background(), daoMock)

	vHostId := request.RouteConfigurations[0].VirtualHosts[0].Routes[0].VirtualHostId
	reqNamespace := ""
	reqVersion := "unknown"
	rawPrefixes := ""

	routes, err := registrationServiceContext.DeleteRoutesByRawPrefixNamespaceVersion(vHostId, reqNamespace, reqVersion, rawPrefixes)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(routes))

	reqVersion = ""
	reqNamespace = "test"
	routes, err = registrationServiceContext.DeleteRoutesByRawPrefixNamespaceVersion(vHostId, reqNamespace, reqVersion)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(routes))

	rawPrefixes = "{/$ test}"
	routes, err = registrationServiceContext.DeleteRoutesByRawPrefixNamespaceVersion(vHostId, reqNamespace, reqVersion, rawPrefixes)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(routes))

	reqNamespace = ""
	reqVersion = ""
	rawPrefixes = ""
	routes, err = registrationServiceContext.DeleteRoutesByRawPrefixNamespaceVersion(vHostId, reqNamespace, reqVersion, rawPrefixes)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, request.RouteConfigurations[0].VirtualHosts[0].Routes[0].Uuid, routes[0].Uuid)
}

func TestDeleteRoutesByCondition(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()

	RegisterRouteCheck(t, daoMock, regSrv, "tlsConfigName", request)
}

func TestRegisterRoutesWithExistedClusterAndTls(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)
	request := getCorrectRequest()

	tlsConfigToApply := &domain.TlsConfig{
		Name:    "tlsConfigName",
		Enabled: true,
	}

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return dao.SaveTlsConfig(tlsConfigToApply)
	})
	assert.Nil(t, err)

	tlsSaved, err := daoMock.FindTlsConfigByName(tlsConfigToApply.Name)

	clusterToApply := &domain.Cluster{
		Name:    "ClusterName1",
		Version: 0,
		TLSId:   tlsSaved.Id,
	}

	_, err = daoMock.WithWTx(func(dao dao.Repository) error {
		return dao.SaveCluster(clusterToApply)
	})
	assert.Nil(t, err)

	RegisterRouteCheck(t, daoMock, regSrv, tlsConfigToApply.Name, request)
}

func TestValidateRequest_shouldReturnFalse_shenRequestIsNotValid(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	domains := []*domain.VirtualHostDomain{
		{
			Domain: "test-service:8080",
		},
	}
	request := getRequest(domains)

	valid, msg, err := Validate(context.Background(), daoMock, request)
	assert.Nil(t, err)
	assert.NotEqual(t, "", msg)
	assert.False(t, valid)

	_, err = daoMock.WithWTx(func(dao dao.Repository) error {
		return regSrv.RegisterRoutes(context.Background(), dao, request)
	})
	assert.Nil(t, err)

	request = getRequest([]*domain.VirtualHostDomain{})
	valid, msg, err = Validate(context.Background(), daoMock, request)
	assert.Nil(t, err)
	assert.NotEqual(t, "", msg)
	assert.False(t, valid)
}

func TestValidateRequest_shouldReturnTrue_shenRequestIsValid(t *testing.T) {
	_, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()
	valid, msg, err := Validate(context.Background(), daoMock, request)
	assert.Nil(t, err)
	assert.Equal(t, "", msg)
	assert.True(t, valid)
}

func TestRegisterRoutes(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)
	request := getCorrectRequest()

	RegisterRouteCheck(t, daoMock, regSrv, "tlsConfigName", request)
}

func TestRegisterRoutesSaveClusterWithDefaultTlSButItHasItInDatabase(t *testing.T) {
	regSrv, daoMock, _ := createServiceWithMode(t)

	request := getCorrectRequest()
	request.ClusterTlsConfig = map[string]string{
		"ClusterName1": "ClusterName1-tls",
	}

	tlsConfigToApply := &domain.TlsConfig{
		Name:    "tlsConfigName",
		Enabled: true,
	}

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return dao.SaveTlsConfig(tlsConfigToApply)
	})
	assert.Nil(t, err)

	tlsSaved, err := daoMock.FindTlsConfigByName(tlsConfigToApply.Name)

	clusterToApply := &domain.Cluster{
		Name:    "ClusterName1",
		Version: 0,
		TLSId:   tlsSaved.Id,
	}

	_, err = daoMock.WithWTx(func(dao dao.Repository) error {
		return dao.SaveCluster(clusterToApply)
	})
	assert.Nil(t, err)

	RegisterRouteCheck(t, daoMock, regSrv, tlsConfigToApply.Name, request)
}

func RegisterRouteCheck(t *testing.T, daoMock *dao.InMemDao, regSrv *RegistrationService, tlsConfigName string, request registration.ProcessedRequest) {

	_, err := daoMock.WithWTx(func(dao dao.Repository) error {
		return regSrv.RegisterRoutes(context.Background(), dao, request)
	})
	assert.Nil(t, err)

	nodeGroups, err := daoMock.FindAllNodeGroups()
	assert.Nil(t, err)
	assert.Equal(t, len(request.NodeGroups), len(nodeGroups))

	deplVersions, err := daoMock.FindAllDeploymentVersions()
	assert.Nil(t, err)
	assert.Equal(t, len(request.DeploymentVersions)-1, len(deplVersions))

	tlsConfigs, err := daoMock.FindAllTlsConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigs))
	assert.Equal(t, tlsConfigName, tlsConfigs[0].Name)

	clusters, err := daoMock.FindAllClusters()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, request.Clusters[0].Endpoints[0], clusters[0].Endpoints[0])
	assert.Equal(t, request.Clusters[0].CircuitBreaker, clusters[0].CircuitBreaker)
	assert.Equal(t, tlsConfigs[0].Id, clusters[0].TLSId)

	listeners, err := daoMock.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(listeners))
	assert.Equal(t, request.Listeners[0], *listeners[0])

	virtualHosts, err := daoMock.FindAllVirtualHosts()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(virtualHosts))
	assert.Equal(t, request.RouteConfigurations[0].VirtualHosts[0], virtualHosts[0])
}

func getCorrectRequest() registration.ProcessedRequest {
	domains := []*domain.VirtualHostDomain{
		{
			Domain: "*",
		},
	}
	return getRequest(domains)
}

func getRequest(domains []*domain.VirtualHostDomain) registration.ProcessedRequest {
	tlsConfig := &domain.TlsConfig{
		Name: "tlsConfigName",
	}
	request := registration.ProcessedRequest{
		NodeGroups: []domain.NodeGroup{
			{Name: "testNodeGroup1"},
			{Name: "testNodeGroup2"},
		},
		DeploymentVersions: []string{"v1", "v2", ""},
		Clusters: []domain.Cluster{
			{Id: 2, Name: "ClusterName1", Endpoints: []*domain.Endpoint{{Id: 3, Address: "test-service", Port: 8080, DeploymentVersion: "v1"}}, CircuitBreakerId: 4, CircuitBreaker: &domain.CircuitBreaker{Id: 4, ThresholdId: 5, Threshold: &domain.Threshold{Id: 5, MaxConnections: 2}}},
		},
		RouteConfigurations: []domain.RouteConfiguration{
			{
				VirtualHosts: []*domain.VirtualHost{
					{
						Id:      6,
						Name:    "VirtualHostName",
						Domains: domains,
						Routes: []*domain.Route{
							{
								VirtualHostId:            6,
								InitialDeploymentVersion: "v1",
								DeploymentVersion:        "v1",
								Uuid:                     "9ff49a83-5747-42b9-a566-716df613545a",
							},
						},
					},
				},
				NodeGroupId: "testNodeGroup1",
				Name:        "name",
			},
		},
		ClusterTlsConfig: map[string]string{
			"ClusterName1": tlsConfig.Name,
		},
		ClusterNodeGroups: map[string][]string{
			"ClusterName1": {"private-gateway-service"},
		},
		Listeners: []domain.Listener{
			{
				Id:                     4,
				Name:                   "private-gateway-service-listener",
				BindHost:               "0.0.0.0",
				BindPort:               "8080",
				RouteConfigurationName: "private-gateway-service-routes",
				NodeGroupId:            "private-gateway-service",
			},
		},
		GroupedRoutes: registration.NewGroupedRoutesMap(),
	}
	request.GroupedRoutes.PutRoute("", request.Clusters[0].Name, request.DeploymentVersions[0], request.RouteConfigurations[0].VirtualHosts[0].Routes[0])

	return request
}

func Test_validateWithDomainProducer(t *testing.T) {
	testData := initTestData()
	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := validateWithDomainProducer(tt.args.ctx, tt.args.existsDomainsProducer, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWithDomainProducer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantValid {
				t.Errorf("validateWithDomainProducer() got = %v, want %v", got, tt.wantValid)
			}
			if got1 != tt.wantMsg {
				t.Errorf("validateWithDomainProducer() got1 = %v, want %v", got1, tt.wantMsg)
			}
		})
	}
}

type RegistrationTestArgs struct {
	ctx                   context.Context
	existsDomainsProducer func() ([]*domain.VirtualHostDomain, error)
	request               registration.ProcessedRequest
}

type RegistrationTestStruct struct {
	name      string
	args      RegistrationTestArgs
	wantValid bool
	wantMsg   string
	wantErr   bool
}

func initTestData() []*RegistrationTestStruct {
	args1 := RegistrationTestArgs{
		ctx:                   context.Background(),
		existsDomainsProducer: func() ([]*domain.VirtualHostDomain, error) { return []*domain.VirtualHostDomain{}, nil },
		request: registration.ProcessedRequest{
			Clusters: []domain.Cluster{
				{
					Endpoints:      []*domain.Endpoint{{Address: "test-service", Port: 8080}},
					CircuitBreaker: &domain.CircuitBreaker{Id: 4, ThresholdId: 5, Threshold: &domain.Threshold{Id: 5, MaxConnections: 2}},
				},
			},
			RouteConfigurations: []domain.RouteConfiguration{
				{
					VirtualHosts: []*domain.VirtualHost{
						{Domains: []*domain.VirtualHostDomain{{Domain: "test-service:8080"}}},
					},
				},
			},
		},
	}

	args2 := RegistrationTestArgs{
		ctx: context.Background(),
		existsDomainsProducer: func() ([]*domain.VirtualHostDomain, error) {
			return []*domain.VirtualHostDomain{{Domain: "test-service:8080"}}, nil
		},
		request: registration.ProcessedRequest{
			Clusters: []domain.Cluster{
				{
					Endpoints:      []*domain.Endpoint{{Address: "test-service", Port: 8080}},
					CircuitBreaker: &domain.CircuitBreaker{Id: 4, ThresholdId: 5, Threshold: &domain.Threshold{Id: 5, MaxConnections: 2}},
				},
			},
			RouteConfigurations: []domain.RouteConfiguration{
				{
					VirtualHosts: []*domain.VirtualHost{
						{Domains: []*domain.VirtualHostDomain{{Domain: "*"}}},
					},
				},
			},
		},
	}

	tests := []*RegistrationTestStruct{
		{
			name:      "Wrong request",
			args:      args1,
			wantValid: false,
			wantMsg:   "Found loop in request data. Virtual host handle requests with Host: test-service:8080 and has route with destination: test-service:8080 at the same time",
			wantErr:   false,
		},
		{
			name:      "Wrong request conflicts with in-memory-data",
			args:      args2,
			wantValid: false,
			wantMsg:   "Found loop in configuration data. Virtual host handle requests with Host: test-service:8080 and has route with destination: test-service:8080 at the same time",
			wantErr:   false,
		},
	}

	return tests
}

func createServiceWithMode(t *testing.T) (*RegistrationService, *dao.InMemDao, *BusMock) {
	busMock := &BusMock{make(map[string][]interface{})}
	entitySrv, daoMock := getDependencies(t)
	routingModeService := routingmode.NewService(daoMock, "v1")
	routeComponentsFactory := factory.NewComponentsFactory(entitySrv)
	regSrv := NewRegistrationService(routeComponentsFactory, entitySrv, daoMock, busMock, routingModeService)

	return regSrv, daoMock, busMock
}

func getDependencies(t *testing.T) (*entity.Service, *dao.InMemDao) {
	mockDao := dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
	v1 := &domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}
	_, err := mockDao.WithWTx(func(dao dao.Repository) error {
		return dao.SaveDeploymentVersion(v1)
	})
	assert.Nil(t, err)
	entityService := entity.NewService("v1")
	return entityService, mockDao
}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

type BusMock struct {
	events map[string][]interface{}
}

func (m *BusMock) Publish(topic string, data interface{}) error {
	if messages, ok := m.events[topic]; ok {
		messages = append(messages, data)
	} else {
		m.events[topic] = []interface{}{data}
	}
	return nil
}

func (m *BusMock) Shutdown() {}
