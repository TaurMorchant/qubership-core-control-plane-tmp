package composite

import (
	context "context"
	"errors"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/clustering"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	events2 "github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/route/factory"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/routingmode"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/rest"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetCompositeStructure_shouldRemoveNameSpace_whenBaselineMode(t *testing.T) {
	originalClient := rest.Client
	defer func() { rest.Client = originalClient }()
	rest.Client = &restClientMock{}

	srv, _, _, _ := createServiceWithMode(t, BaselineMode)
	namespaceName := "namespace"

	err := srv.AddCompositeNamespace(context.Background(), namespaceName)
	assert.Nil(t, err)

	structure, err := srv.GetCompositeStructure()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(structure.Satellites))
	assert.Equal(t, namespaceName, structure.Satellites[0])
}

func TestGetCompositeStructure_shouldReturnError_whenNotBaselineMode(t *testing.T) {
	srv, _, _, _ := createService(t)
	structure, err := srv.GetCompositeStructure()
	assert.NotNil(t, err)
	assert.Equal(t, ErrSatelliteMode, err)

	assert.Nil(t, structure.Satellites)
	assert.Equal(t, "", structure.Baseline)
}

func TestRemoveCompositeNamespace_shouldRemoveNameSpace_whenBaselineMode(t *testing.T) {
	originalClient := rest.Client
	defer func() { rest.Client = originalClient }()
	rest.Client = &restClientMock{}

	srv, _, daoMock, _ := createServiceWithMode(t, BaselineMode)
	namespaceName := "namespace"

	err := srv.AddCompositeNamespace(context.Background(), namespaceName)
	assert.Nil(t, err)

	composite, err := daoMock.FindCompositeSatelliteByNamespace(namespaceName)
	assert.Equal(t, namespaceName, composite.Namespace)

	err = srv.RemoveCompositeNamespace(namespaceName)
	assert.Nil(t, err)

	composites, err := daoMock.FindAllCompositeSatellites()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(composites))
}

func TestRemoveCompositeNamespace_shouldReturnError_whenEmptyNameSpace(t *testing.T) {
	srv, _, _, _ := createServiceWithMode(t, BaselineMode)
	err := srv.RemoveCompositeNamespace("")
	assert.NotNil(t, err)
	assert.Equal(t, "composite: attempt to delete empty composite satellite namespace from the composite structure", err.Error())
}

func TestRemoveCompositeNamespace_shouldReturnError_whenNotBaselineMode(t *testing.T) {
	srv, _, _, _ := createService(t)
	err := srv.RemoveCompositeNamespace("namespace")
	assert.NotNil(t, err)
	assert.Equal(t, ErrSatelliteMode, err)
}

func TestAddCompositeNamespace_shouldAddNameSpace_whenBaselineMode(t *testing.T) {
	originalClient := rest.Client
	defer func() { rest.Client = originalClient }()
	rest.Client = &restClientMock{}

	srv, _, daoMock, _ := createServiceWithMode(t, BaselineMode)
	namespaceName := "namespace"

	err := srv.AddCompositeNamespace(context.Background(), namespaceName)
	assert.Nil(t, err)

	composite, err := daoMock.FindCompositeSatelliteByNamespace(namespaceName)
	assert.Equal(t, namespaceName, composite.Namespace)

	err = srv.AddCompositeNamespace(context.Background(), namespaceName)
	assert.Nil(t, err)

	composites, err := daoMock.FindAllCompositeSatellites()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(composites))
}

func TestAddCompositeNamespace_shouldReturnError_whenEmptyNameSpace(t *testing.T) {
	srv, _, _, _ := createServiceWithMode(t, BaselineMode)
	err := srv.AddCompositeNamespace(context.Background(), "")
	assert.NotNil(t, err)
	assert.Equal(t, errorcodes.ValidationRequestError.Code, err.(*errorcodes.CpErrCodeError).ErrorCode.Code)
	assert.Equal(t, "composite: attempt to add empty composite satellite namespace to the composite structure", err.(*errorcodes.CpErrCodeError).Detail)
}

func TestAddCompositeNamespace_shouldReturnError_whenNotBaselineMode(t *testing.T) {
	srv, _, _, _ := createService(t)
	err := srv.AddCompositeNamespace(context.Background(), "namespace")
	assert.NotNil(t, err)
	assert.Equal(t, ErrSatelliteMode, err)
}

func TestServiceInitSatellite_shouldCreateFallbackRoutes_whenStatusCodeCorrect(t *testing.T) {
	srv, entitySrv, daoMock, busMock := createService(t)
	configloader.InitWithSourcesArray([]*configloader.PropertySource{})
	tlsmode.SetUpTlsProperties()

	rest.Client = &restClientMock{}
	err := srv.InitSatellite(500 * time.Millisecond)
	assert.Nil(t, err)

	verifyGatewayFallbackRoute(t, domain.PublicGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.PrivateGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.InternalGateway, entitySrv, daoMock, busMock, false)
	listeners, err := daoMock.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(listeners))

	// verify operation idempotent
	err = srv.InitSatellite(500 * time.Millisecond)
	verifyGatewayFallbackRoute(t, domain.PublicGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.PrivateGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.InternalGateway, entitySrv, daoMock, busMock, false)
	assert.Nil(t, err)
	listeners, err = daoMock.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(listeners))
}

func TestServiceInitSatellite_shouldSaveFatalErrors_whenStatusCodeWrong(t *testing.T) {
	srv, _, _, _ := createService(t)
	configloader.InitWithSourcesArray([]*configloader.PropertySource{})
	clustering.CleanFatalErrors()

	rest.Client = &restClientStatusCodeWrongMock{}
	err := srv.InitSatellite(500 * time.Millisecond)
	assert.NotNil(t, err)

	fatalErrors := clustering.GetFatalErrors()
	assert.True(t, len(fatalErrors) > 0)
	assert.Equal(t, "unexpected response code when registering this namespace in composite with baseline test-baseline: 404", fatalErrors[0].Error())
}

func TestServiceInitSatellite_shouldSaveFatalErrors_whenRestFails(t *testing.T) {
	srv, _, _, _ := createService(t)
	configloader.InitWithSourcesArray([]*configloader.PropertySource{})
	clustering.CleanFatalErrors()

	rest.Client = &restClientFailsMock{}
	err := srv.InitSatellite(500 * time.Millisecond)
	assert.NotNil(t, err)

	fatalErrors := clustering.GetFatalErrors()
	assert.True(t, len(fatalErrors) > 0)
	assert.Equal(t, "failed to register this namespace in composite with baseline test-baseline: error", fatalErrors[0].Error())
}

func TestServiceModeString(t *testing.T) {
	var TestBaselineMode ServiceMode = 3
	result := TestBaselineMode.String()
	assert.Equal(t, "<Unknown composite.ServiceMode: 3>", result)

	result = SatelliteMode.String()
	assert.Equal(t, "SatelliteMode", result)

	result = BaselineMode.String()
	assert.Equal(t, "BaselineMode", result)
}

func TestInitSatellite_createFallbackRoutesToBaseline(t *testing.T) {
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()

	srv, entitySrv, daoMock, busMock := createService(t)
	err := srv.createFallbackRoutesToBaseline()

	verifyGatewayFallbackRoute(t, domain.PublicGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.PrivateGateway, entitySrv, daoMock, busMock, false)
	verifyGatewayFallbackRoute(t, domain.InternalGateway, entitySrv, daoMock, busMock, false)

	assert.Nil(t, err)
}

func TestInitSatellite_createFallbackRoutesToBaselineWithTls(t *testing.T) {
	_ = os.Setenv("INTERNAL_TLS_ENABLED", "true")
	defer os.Unsetenv("INTERNAL_TLS_ENABLED")
	configloader.Init(configloader.EnvPropertySource())
	tlsmode.SetUpTlsProperties()

	srv, entitySrv, daoMock, busMock := createService(t)
	err := srv.createFallbackRoutesToBaseline()

	verifyGatewayFallbackRoute(t, domain.PublicGateway, entitySrv, daoMock, busMock, true)
	verifyGatewayFallbackRoute(t, domain.PrivateGateway, entitySrv, daoMock, busMock, true)
	verifyGatewayFallbackRoute(t, domain.InternalGateway, entitySrv, daoMock, busMock, true)

	assert.Nil(t, err)
}

func createServiceWithMode(t *testing.T, mode ServiceMode) (*Service, *entity.Service, *dao.InMemDao, *BusMock) {
	configloader.Init(configloader.EnvPropertySource())

	busMock := &BusMock{make(map[string][]interface{})}
	entitySrv, daoMock := getDependencies(t)
	routingModeService := routingmode.NewService(daoMock, "v1")
	routeComponentsFactory := factory.NewComponentsFactory(entitySrv)
	regSrv := route.NewRegistrationService(routeComponentsFactory, entitySrv, daoMock, busMock, routingModeService)

	return NewService("test-baseline", mode, daoMock, entitySrv, regSrv, busMock), entitySrv, daoMock, busMock
}

func createService(t *testing.T) (*Service, *entity.Service, *dao.InMemDao, *BusMock) {
	return createServiceWithMode(t, SatelliteMode)
}

func verifyGatewayFallbackRoute(t *testing.T, gateway string, entitySrv *entity.Service, mockDao *dao.InMemDao, busMock *BusMock, withTls bool) {
	var clusterName string
	if withTls {
		clusterName = fmt.Sprintf("%s||%s.test-baseline||8443", gateway, gateway)
	} else {
		clusterName = fmt.Sprintf("%s||%s.test-baseline||8080", gateway, gateway)
	}
	err := mockDao.WithRTx(func(dao dao.Repository) error {
		cluster, err := entitySrv.GetClusterWithRelations(dao, clusterName)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(cluster.Endpoints))
		assert.Equal(t, gateway+".test-baseline", cluster.Endpoints[0].Address)
		if withTls {
			assert.Equal(t, int32(8443), cluster.Endpoints[0].Port)
			assert.Equal(t, "https", cluster.Endpoints[0].Protocol)
		} else {
			assert.Equal(t, int32(8080), cluster.Endpoints[0].Port)
			assert.Equal(t, "http", cluster.Endpoints[0].Protocol)
		}
		routes, err := dao.FindRoutesByClusterName(clusterName)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(routes))
		assert.Equal(t, "/", routes[0].Prefix)
		assert.Equal(t, "", routes[0].PrefixRewrite)
		assert.True(t, routes[0].Fallback.Valid)
		assert.True(t, routes[0].Fallback.Bool)
		assert.Equal(t, 1, len(routes[0].RequestHeadersToAdd))
		assert.True(t, routes[0].RequestHeadersToAdd[0].Equals(
			domain.Header{
				Name:  "X-Token-Signature",
				Value: "%DYNAMIC_METADATA(envoy.filters.http.ext_authz:x.token.signature)%",
			}))
		return nil
	})
	assert.Nil(t, err)

	events := busMock.events[bus.TopicMultipleChanges]
	assert.Equal(t, 1, len(events))
	event := events[0].(*events2.MultipleChangeEvent)
	hasCluster := false
	hasEndpoint := false
	hasRoute := false
	hasTlsConfig := !withTls
	for entityName, changes := range event.Changes {
		switch entityName {
		case domain.ClusterTable:
			for _, change := range changes {
				if change.After.(*domain.Cluster).Name == clusterName {
					hasCluster = true
					break
				}
			}
			break
		case domain.TlsConfigTable:
			for _, change := range changes {
				if change.After.(*domain.TlsConfig).Name == gateway+"-tls" {
					assert.Empty(t, change.After.(*domain.TlsConfig).TrustedCA)
					assert.Empty(t, change.After.(*domain.TlsConfig).ClientCert)
					assert.Empty(t, change.After.(*domain.TlsConfig).PrivateKey)
					hasTlsConfig = true
					break
				}
			}
			break
		case domain.EndpointTable:
			for _, change := range changes {
				if change.After.(*domain.Endpoint).Address == gateway+".test-baseline" {
					if withTls {
						assert.Equal(t, int32(8443), change.After.(*domain.Endpoint).Port)
						assert.Equal(t, "https", change.After.(*domain.Endpoint).Protocol)
					} else {
						assert.Equal(t, int32(8080), change.After.(*domain.Endpoint).Port)
						assert.Equal(t, "http", change.After.(*domain.Endpoint).Protocol)
					}
					hasEndpoint = true
					break
				}
			}
			break
		case domain.RouteTable:
			for _, change := range changes {
				if change.After.(*domain.Route).ClusterName == clusterName {
					route := change.After.(*domain.Route)
					assert.Equal(t, "/", route.Prefix)
					assert.Equal(t, "", route.PrefixRewrite)
					assert.True(t, route.Fallback.Valid)
					assert.True(t, route.Fallback.Bool)
					assert.Equal(t, 1, len(route.RequestHeadersToAdd))
					assert.True(t, route.RequestHeadersToAdd[0].Equals(
						domain.Header{
							Name:  "X-Token-Signature",
							Value: "%DYNAMIC_METADATA(envoy.filters.http.ext_authz:x.token.signature)%",
						}))
					hasRoute = true
					break
				}
			}
			break
		}
	}
	assert.True(t, hasCluster)
	assert.True(t, hasEndpoint)
	assert.True(t, hasRoute)
	assert.True(t, hasTlsConfig)
}

type restClientStatusCodeWrongMock struct {
}

func (cl restClientStatusCodeWrongMock) DoRetryRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	resp.SetStatusCode(fasthttp.StatusNotFound)
	return resp, nil
}

func (cl restClientStatusCodeWrongMock) DoRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	resp.SetStatusCode(fasthttp.StatusNotFound)
	return resp, nil
}

type restClientFailsMock struct {
}

func (cl restClientFailsMock) DoRetryRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	return nil, errors.New("error")
}

func (cl restClientFailsMock) DoRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	return nil, errors.New("error")
}

type restClientMock struct {
}

func (cl restClientMock) DoRetryRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	resp.SetStatusCode(fasthttp.StatusOK)
	return resp, nil
}

func (cl restClientMock) DoRequest(context.Context, string, string, []byte, logging.Logger) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	resp.SetStatusCode(fasthttp.StatusOK)
	return resp, nil
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
