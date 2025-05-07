package dr

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	mock_clustering "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/clustering"
	mock_constancy "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/constancy"
	mock_dao "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/dao"
	mock_db "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/db"
	mock_bus "github.com/netcracker/qubership-core-control-plane/control-plane/v2/test/mock/event/bus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestProcessNotification_postProcessChanges_DeploymentVersionTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, _, _, _, _, _, _ := getService(ctrl)

	changes := []memdb.Change{
		{
			Before: &domain.DeploymentVersion{
				Version: "v1",
			},
		},
		{
			After: &domain.DeploymentVersion{
				Version: "v2",
			},
		},
	}
	drService.postProcessChanges(domain.DeploymentVersionTable, changes)
}

func TestProcessNotification_VirtualHostDomainTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, _, _, _ := getService(ctrl)

	entity := &domain.VirtualHostDomain{
		Domain:        "domain",
		Version:       1,
		VirtualHostId: int32(1),
	}
	mockDao.EXPECT().SaveEntity("virtual_host_domains", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"virtual_host_domains\", \"virtualhostid\": 1, \"domain\": \"domain\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().FindVirtualHostDomainByVirtualHostId(entity.VirtualHostId).Return([]*domain.VirtualHostDomain{entity}, nil)
	mockDao.EXPECT().DeleteVirtualHostsDomain(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"virtual_host_domains\", \"virtualhostid\": 1, \"domain\": \"domain\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_VirtualHostTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.VirtualHost{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindVirtualHostById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("virtual_hosts", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"virtual_hosts\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().FindVirtualHostById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().DeleteVirtualHost(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"virtual_hosts\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_WasmFilterTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.WasmFilter{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindWasmFilterById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("wasm_filters", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"wasm_filters\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteWasmFilterById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"wasm_filters\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_TlsConfigTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.TlsConfig{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindTlsConfigById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("tls_configs", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"tls_configs\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteTlsConfigById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"tls_configs\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_RouteConfigurationTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.RouteConfiguration{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindRouteConfigById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("route_configurations", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"route_configurations\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteRouteConfigById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"route_configurations\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_RetryPolicyTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.RetryPolicy{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindRetryPolicyById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("retry_policy", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"retry_policy\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteRetryPolicyById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"retry_policy\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_NodeGroupTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.NodeGroup{
		Name: "name",
	}

	mockStorage.EXPECT().FindNodeGroupByName("name").Return(entity, nil)
	mockDao.EXPECT().SaveEntity("node_groups", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"node_groups\", \"name\": \"name\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteNodeGroupByName(entity.Name)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"node_groups\", \"name\": \"name\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_ListenersWasmFilterTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, _, _, _ := getService(ctrl)

	entity := &domain.ListenersWasmFilter{
		ListenerId:   int32(1),
		WasmFilterId: int32(1),
	}
	mockDao.EXPECT().SaveEntity("listeners_wasm_filters", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"listeners_wasm_filters\", \"listener_id\": 1, \"wasm_filter_id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteListenerWasmFilter(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"listeners_wasm_filters\", \"listener_id\": 1, \"wasm_filter_id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_ListenerTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.Listener{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindListenerById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("listeners", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"listeners\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteListenerById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"listeners\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_HealthCheckTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.HealthCheck{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindHealthCheckById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("health_check", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"health_check\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteHealthCheckById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"health_check\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_HeaderMatcherTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.HeaderMatcher{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindHeaderMatcherById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("header_matchers", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"header_matchers\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteHeaderMatcherById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"header_matchers\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_HashPolicyTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.HashPolicy{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindHashPolicyById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("hash_policy", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"hash_policy\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteHashPolicyById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"hash_policy\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_EndpointTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.Endpoint{
		Id: int32(1),
	}
	mockStorage.EXPECT().FindEndpointById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("endpoints", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"endpoints\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteEndpoint(entity)
	mockDao.EXPECT().FindEndpointById(entity.Id).Return(entity, nil)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"endpoints\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_CompositeSatelliteTable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.CompositeSatellite{
		Namespace: "namespace",
	}
	mockStorage.EXPECT().FindCompositeSatelliteByNamespace(entity.Namespace).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("composite_satellites", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"composite_satellites\", \"namespace\": \"namespace\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteCompositeSatellite(entity.Namespace)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"composite_satellites\", \"namespace\": \"namespace\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_ClustersNodeGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.ClustersNodeGroup{
		ClustersId:     int32(1),
		NodegroupsName: "name",
	}
	mockStorage.EXPECT().FindClustersNodeGroupByIdAndNodeGroup(entity.ClustersId, entity.NodegroupsName).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("clusters_node_groups", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"clusters_node_groups\", \"clusters_id\": 1, \"nodegroups_name\": \"name\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteClustersNodeGroup(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"clusters_node_groups\", \"clusters_id\": 1, \"nodegroups_name\": \"name\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_Cluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.Cluster{
		Name: "clusterName",
	}
	mockStorage.EXPECT().FindClusterByName(entity.Name).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("clusters", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"clusters\", \"name\": \"clusterName\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteClusterByName(entity.Name)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"clusters\", \"name\": \"clusterName\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_CircuitBreaker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.CircuitBreaker{
		Id: 1,
	}
	mockStorage.EXPECT().FindCircuitBreakerById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("circuit_breakers", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"circuit_breakers\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteCircuitBreakerById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"circuit_breakers\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_Threshold(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.Threshold{
		Id: 1,
	}
	mockStorage.EXPECT().FindThresholdById(entity.Id).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("thresholds", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"thresholds\", \"id\": 1}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteThresholdById(entity.Id)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"thresholds\", \"id\": 1}"
	drService.processNotification(payload)
}

func TestProcessNotification_DeploymentVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.DeploymentVersion{
		Version: "v1",
	}
	mockStorage.EXPECT().FindDeploymentVersionByName("v1").Return(entity, nil)
	mockDao.EXPECT().SaveEntity("deployment_versions", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"deployment_versions\", \"version\": \"v1\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteDeploymentVersion(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"deployment_versions\", \"version\": \"v1\"}"
	drService.processNotification(payload)
}

func TestProcessNotification_EnvoyConfigVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, mockBus, _, _, _, _ := getService(ctrl)

	entity := &domain.EnvoyConfigVersion{
		NodeGroup:  "public-gateway-service",
		EntityType: "listeners",
		Version:    int64(1646309974795615488),
	}
	mockDao.EXPECT().SaveEntity(domain.EnvoyConfigVersionTable, entity)
	mockBus.EXPECT().Publish(bus.TopicPartialReapply, &events.PartialReloadEvent{EnvoyVersions: []*domain.EnvoyConfigVersion{entity}}).Return(nil)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"envoy_config_version\", \"node_group\": \"public-gateway-service\", \"entity_type\": \"listeners\", \"version\": 1646309974795615435}"
	drService.processNotification(payload)

	payload = "{\"operation\": \"DELETE\", \"entity\": \"envoy_config_version\", \"node_group\": \"public-gateway-service\", \"entity_type\": \"listeners\", \"version\": 1646309974795615435}"
	drService.processNotification(payload)
}

func TestProcessNotification_Route(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, mockBus, _, mockStorage, _, _ := getService(ctrl)

	routeId := int32(1)
	mockDao.EXPECT().DeleteRouteById(routeId).Return(nil)
	mockBus.EXPECT().Publish(gomock.Any(), gomock.Any()).Times(0).Return(nil)

	payload := fmt.Sprintf("{\"operation\": \"DELETE\", \"entity\": \"routes\", \"id\": %d}", routeId)
	drService.processNotification(payload)

	route := &domain.Route{
		Id: routeId,
	}
	mockStorage.EXPECT().FindRouteById(routeId).Return(route, nil)
	mockDao.EXPECT().SaveEntity("routes", route)
	payload = fmt.Sprintf("{\"operation\": \"UPSERT\", \"entity\": \"routes\", \"id\": %d}", routeId)
	drService.processNotification(payload)
}

func TestInitStorageAfterConnect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, _, _, _, _, _, mockMasterNodeInitializer := getService(ctrl)

	mockMasterNodeInitializer.EXPECT().InitMaster()
	drService.initStorageAfterConnect()
}

func TestCloseListener(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, _, _, _, _, mockListener, _ := getService(ctrl)

	mockListener.EXPECT().Close()
	drService.Close()
}

func TestDrStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDao := getMockDao(ctrl)
	mockBus := getMockBus(ctrl)
	mockDBProvider := getMockDBProvider(ctrl)
	mockMasterNodeInitializer := getMockMasterNodeInitializer(ctrl)
	mockStorage := getMockStorage(ctrl)
	drService := Service{
		MasterInitializer: mockMasterNodeInitializer,
		DBProvider:        mockDBProvider,
		ConstantStorage:   mockStorage,
		Dao:               mockDao,
		Bus:               mockBus,
	}

	mockListener := getMockPersistentStorageListener(ctrl)
	mockDBProvider.EXPECT().Listen(NotifyChannelName, gomock.Any(), gomock.Any()).Return(mockListener, nil)

	err := drService.Start()
	assert.Nil(t, err)

	mockDBProvider.EXPECT().Listen(NotifyChannelName, gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	err = drService.Start()
	assert.NotNil(t, err)
}

func getService(ctrl *gomock.Controller) (Service, *mock_dao.MockDao, *mock_bus.MockBusPublisher, *mock_db.MockDBProvider, *mock_constancy.MockStorage, *mock_db.MockPersistentStorageListener, *mock_clustering.MockMasterNodeInitializer) {
	mockDao := getMockDao(ctrl)
	mockBus := getMockBus(ctrl)
	mockDBProvider := getMockDBProvider(ctrl)
	mockMasterNodeInitializer := getMockMasterNodeInitializer(ctrl)
	mockStorage := getMockStorage(ctrl)
	mockListener := getMockPersistentStorageListener(ctrl)

	drService := Service{
		MasterInitializer: mockMasterNodeInitializer,
		DBProvider:        mockDBProvider,
		ConstantStorage:   mockStorage,
		Dao:               mockDao,
		Bus:               mockBus,
		mutex:             &sync.Mutex{},
		dbListener:        mockListener,
	}
	changes := []memdb.Change{
		{},
	}
	mockDao.EXPECT().WithWTx(gomock.Any()).AnyTimes().Return(changes, nil).Do(func(args ...interface{}) {
		args[0].(func(dao dao.Repository) error)(mockDao)
	})

	return drService, mockDao, mockBus, mockDBProvider, mockStorage, mockListener, mockMasterNodeInitializer
}

func getMockDao(ctrl *gomock.Controller) *mock_dao.MockDao {
	mock := mock_dao.NewMockDao(ctrl)
	return mock
}

func getMockBus(ctrl *gomock.Controller) *mock_bus.MockBusPublisher {
	mock := mock_bus.NewMockBusPublisher(ctrl)
	return mock
}

func getMockDBProvider(ctrl *gomock.Controller) *mock_db.MockDBProvider {
	mock := mock_db.NewMockDBProvider(ctrl)
	return mock
}

func getMockMasterNodeInitializer(ctrl *gomock.Controller) *mock_clustering.MockMasterNodeInitializer {
	mock := mock_clustering.NewMockMasterNodeInitializer(ctrl)
	return mock
}

func getMockPersistentStorageListener(ctrl *gomock.Controller) *mock_db.MockPersistentStorageListener {
	mock := mock_db.NewMockPersistentStorageListener(ctrl)
	return mock
}

func getMockStorage(ctrl *gomock.Controller) *mock_constancy.MockStorage {
	mock := mock_constancy.NewMockStorage(ctrl)
	return mock
}

func TestProcessNotification_TlgConfigNodeGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	drService, mockDao, _, _, mockStorage, _, _ := getService(ctrl)

	entity := &domain.TlsConfigsNodeGroups{
		TlsConfigId:   int32(1),
		NodeGroupName: "name",
	}
	mockStorage.EXPECT().FindTlsConfigByIdAndNodeGroupName(entity.TlsConfigId, entity.NodeGroupName).Return(entity, nil)
	mockDao.EXPECT().SaveEntity("tls_configs_node_groups", entity)

	payload := "{\"operation\": \"UPSERT\", \"entity\": \"tls_configs_node_groups\", \"tls_config_id\": 1, \"node_group_name\": \"name\"}"
	drService.processNotification(payload)

	mockDao.EXPECT().DeleteTlsConfigByIdAndNodeGroupName(entity)
	payload = "{\"operation\": \"DELETE\", \"entity\": \"tls_configs_node_groups\", \"tls_config_id\": 1, \"node_group_name\": \"name\"}"
	drService.processNotification(payload)
}
