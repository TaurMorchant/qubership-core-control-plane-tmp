package cache

//
//import (
//	"context"
//	"control-plane/dao"
//	"control-plane/domain"
//	"control-plane/envoy/cache/action"
//	"control-plane/envoy/cache/builder"
//	"control-plane/envoy/cache/builder/common"
//	"control-plane/envoy/cache/event"
//	"control-plane/event/events"
//	"control-plane/ram"
//	mock_dao "control-plane/test/mock/dao"
//	mock_action "control-plane/test/mock/envoy/cache/action"
//	mock_event "control-plane/test/mock/envoy/cache/event"
//	"errors"
//	v3listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
//	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
//	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
//	"github.com/envoyproxy/go-control-plane/pkg/server/stream/v3"
//	"github.com/golang/mock/gomock"
//	"github.com/hashicorp/go-memdb"
//	"github.com/stretchr/testify/assert"
//	"strconv"
//	"sync"
//	"sync/atomic"
//	"testing"
//)
//
//func TestUpdateManagerInitConfigWithRetry(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	snapshot := getSnapshot()
//	mockDao, mockUpdateAction, mockSnapshotUpdateAction, mockChangeEventParser, _ := getUpdateManagerMocks(ctrl)
//	mockDao.EXPECT().FindEnvoyConfigVersion(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
//
//	mockSnapshotUpdateAction.EXPECT().Perform(gomock.Any()).Return(snapshot, nil).Times(4)
//
//	mockUpdateAction.EXPECT().ListenerUpdateAction(gomock.Any(), gomock.Any()).Return(mockSnapshotUpdateAction)
//	mockUpdateAction.EXPECT().ClusterUpdateAction(gomock.Any(), gomock.Any()).Return(mockSnapshotUpdateAction)
//	mockUpdateAction.EXPECT().RouteConfigUpdateAction(gomock.Any(), gomock.Any()).Return(mockSnapshotUpdateAction)
//	mockUpdateAction.EXPECT().RuntimeUpdateAction(gomock.Any()).Return(mockSnapshotUpdateAction)
//
//	updateManager := getUpdateManagerWithMocks(mockDao, mockUpdateAction, mockChangeEventParser)
//
//	httpVersion := int32(1)
//	nodeGroup := &domain.NodeGroup{
//		Name: "nodeGroup1",
//		Clusters: []*domain.Cluster{
//			{
//				Id:          int32(1),
//				Name:        "cluster name1",
//				HttpVersion: &httpVersion,
//			},
//		},
//	}
//	mockDao.EXPECT().FindAllNodeGroups().Return([]*domain.NodeGroup{
//		nodeGroup,
//	}, nil)
//
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroup.Name, snapshot)
//	mockDao.EXPECT().FindListenersByNodeGroupId(nodeGroup.Name).Return([]*domain.Listener{{}}, nil)
//	mockDao.EXPECT().FindClusterByNodeGroup(nodeGroup).Return([]*domain.Cluster{{}}, nil)
//	mockDao.EXPECT().FindRouteConfigsByNodeGroupId(nodeGroup.Name).Return([]*domain.RouteConfiguration{{}}, nil)
//
//	updateManager.InitConfigWithRetry()
//}
//
//func TestUpdateManagerBulkUpdateClusters(t *testing.T) {
//	version := int64(2)
//	httpVersion := int32(1)
//
//	updateManager := getUpdateManager()
//	snapshot := getSnapshot()
//
//	clusters := []*domain.Cluster{
//		{
//			Id:          int32(1),
//			Name:        "cluster name1",
//			HttpVersion: &httpVersion,
//		},
//		{
//			Id:          int32(2),
//			Name:        "cluster name2",
//			HttpVersion: &httpVersion,
//		},
//	}
//
//	nodeGroups := []domain.NodeGroup{
//		{
//			Name: "nodeGroup1",
//			Clusters: []*domain.Cluster{
//				clusters[0],
//			},
//		},
//		{
//			Name: "nodeGroup2",
//			Clusters: []*domain.Cluster{
//				clusters[1],
//			},
//		},
//	}
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroups[0].Name, snapshot)
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroups[1].Name, snapshot)
//
//	updateManager.dao.WithWTx(func(dao dao.Repository) error {
//		dao.SaveNodeGroup(&nodeGroups[0])
//		dao.SaveNodeGroup(&nodeGroups[1])
//		return nil
//	})
//
//	clusterNodeGroup0 := &domain.ClustersNodeGroup{
//		ClustersId:     clusters[0].Id,
//		Cluster:        clusters[0],
//		NodegroupsName: nodeGroups[0].Name,
//		NodeGroup:      &nodeGroups[0],
//	}
//
//	clusterNodeGroup1 := &domain.ClustersNodeGroup{
//		ClustersId:     clusters[1].Id,
//		Cluster:        clusters[1],
//		NodegroupsName: nodeGroups[1].Name,
//		NodeGroup:      &nodeGroups[1],
//	}
//
//	updateManager.dao.WithWTx(func(dao dao.Repository) error {
//		dao.SaveClustersNodeGroup(clusterNodeGroup0)
//		dao.SaveClustersNodeGroup(clusterNodeGroup1)
//		return nil
//	})
//
//	updateActionProvider := func(nodeGroup, version string, entity interface{}) action.SnapshotUpdateAction {
//		return &testType{
//			Snapshot: &cache.Snapshot{
//				Resources: [types.UnknownType]cache.Resources{
//					{
//						Version: "10",
//					},
//				},
//				VersionMap: map[string]map[string]string{
//					"namespace-new-1": {
//						"/fromA": "/toA",
//					},
//					"namespace-new-2": {
//						"/fromB": "/toB",
//					},
//				},
//			},
//		}
//	}
//
//	firstNodeSnapshot, _ := updateManager.cache.GetSnapshot(nodeGroups[0].Name)
//	assert.Equal(t, "1", firstNodeSnapshot.(*cache.Snapshot).Resources[0].Version)
//
//	secondNodeSnapshot, _ := updateManager.cache.GetSnapshot(nodeGroups[1].Name)
//	assert.Equal(t, "1", secondNodeSnapshot.(*cache.Snapshot).Resources[0].Version)
//
//	err := updateManager.BulkUpdateClusters(version, clusters, updateActionProvider)
//	assert.Nil(t, err)
//
//	firstNodeSnapshot, _ = updateManager.cache.GetSnapshot(nodeGroups[0].Name)
//	assert.Equal(t, "10", firstNodeSnapshot.(*cache.Snapshot).Resources[0].Version)
//
//	secondNodeSnapshot, _ = updateManager.cache.GetSnapshot(nodeGroups[1].Name)
//	assert.Equal(t, "10", secondNodeSnapshot.(*cache.Snapshot).Resources[0].Version)
//}
//
//func TestUpdateManagerUpdateCluster(t *testing.T) {
//	nodeGroup := "test"
//	clusterName := "cluster name"
//	version := int64(2)
//	httpVersion := int32(1)
//	cluster := &domain.Cluster{
//		Name:        clusterName,
//		HttpVersion: &httpVersion,
//	}
//
//	updateManager := getUpdateManager()
//	snapshot := getSnapshot()
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroup, snapshot)
//
//	err := updateManager.UpdateCluster(nodeGroup, version, cluster)
//	assert.Nil(t, err)
//	resultSnapshot, _ := updateManager.cache.GetSnapshot(nodeGroup)
//	assert.NotNil(t, resultSnapshot.(*cache.Snapshot).Resources[types.Cluster].Items[clusterName])
//	assert.Equal(t, strconv.FormatInt(version, 10), resultSnapshot.(*cache.Snapshot).Resources[types.Cluster].Version)
//}
//
//func TestUpdateManagerUpdateListener(t *testing.T) {
//	nodeGroup := "test"
//	listenerName := "test name"
//	version := int64(2)
//	listener := &domain.Listener{
//		NodeGroupId: nodeGroup,
//		Name:        listenerName,
//	}
//	namespaceMapping := "namespace"
//
//	updateManager := getUpdateManager()
//	snapshot := getSnapshot()
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroup, snapshot)
//
//	err := updateManager.UpdateListener(nodeGroup, version, listener, namespaceMapping)
//	assert.Nil(t, err)
//	resultSnapshot, _ := updateManager.cache.GetSnapshot(nodeGroup)
//	assert.NotNil(t, resultSnapshot.(*cache.Snapshot).Resources[types.Listener].Items[listenerName])
//	assert.Equal(t, strconv.FormatInt(version, 10), resultSnapshot.(*cache.Snapshot).Resources[types.Listener].Version)
//}
//
//func TestUpdateManagerHandleMultipleChangeEvent(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockDao, mockUpdateAction, mockSnapshotUpdateAction, mockChangeEventParser, _ := getUpdateManagerMocks(ctrl)
//	updateManager := getUpdateManagerWithMocks(mockDao, mockUpdateAction, mockChangeEventParser)
//
//	nodeGroup := "1"
//
//	changeEvent := &events.MultipleChangeEvent{
//		Changes: map[string][]memdb.Change{
//			domain.EnvoyConfigVersionTable: {
//				{
//					Table: "table",
//					After: &domain.EnvoyConfigVersion{
//						NodeGroup:  nodeGroup,
//						EntityType: "type",
//						Version:    int64(12345),
//					},
//				},
//			},
//		},
//	}
//	snapshot := getSnapshot()
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroup, snapshot)
//
//	actionTest := map[string]action.SnapshotUpdateAction{
//		domain.EnvoyConfigVersionTable: mockSnapshotUpdateAction,
//	}
//	mockChangeEventParser.EXPECT().ParseMultipleChangeEvent(changeEvent).Return(actionTest)
//	mockSnapshotUpdateAction.EXPECT().Perform(gomock.Any()).Return(snapshot, nil).Times(1)
//
//	updateManager.HandleMultipleChangeEvent(changeEvent)
//}
//
//func TestUpdateManagerHandleChangeEven(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	mockDao, mockUpdateAction, mockSnapshotUpdateAction, mockChangeEventParser, mockActionsMap := getUpdateManagerMocks(ctrl)
//	updateManager := getUpdateManagerWithMocks(mockDao, mockUpdateAction, mockChangeEventParser)
//
//	nodeGroup := "1"
//
//	changeEvent := &events.ChangeEvent{
//		NodeGroup: nodeGroup,
//		Changes: map[string][]memdb.Change{
//			domain.EnvoyConfigVersionTable: {
//				{
//					Table: "table",
//					After: &domain.EnvoyConfigVersion{
//						NodeGroup:  nodeGroup,
//						EntityType: "type",
//						Version:    int64(12345),
//					},
//				},
//			},
//		},
//	}
//	snapshot := getSnapshot()
//	updateManager.cache.SetSnapshot(context.Background(), nodeGroup, snapshot)
//
//	compositeUpdateAction := action.NewCompositeUpdateAction([]action.SnapshotUpdateAction{
//		mockSnapshotUpdateAction,
//	})
//	mockActionsMap.EXPECT().CompositeAction().Return(compositeUpdateAction)
//	mockChangeEventParser.EXPECT().ParseChangeEvent(changeEvent).Return(mockActionsMap)
//	mockSnapshotUpdateAction.EXPECT().Perform(gomock.Any()).Return(snapshot, nil).Times(1)
//
//	updateManager.HandleChangeEvent(changeEvent)
//}
//
//func TestUpdateManagerUpdateSnapshot_shouldSetEmptySnapshot_whenGetCacheReturnError(t *testing.T) {
//	updateManager := getUpdateManager()
//
//	testAction := testType{
//		Snapshot: getSnapshot(),
//		Err:      errors.New("error"),
//	}
//	nodeGroup := "error"
//
//	err := updateManager.UpdateSnapshot(nodeGroup, testAction)
//	assert.NotNil(t, err)
//	snapshotCache, _ := updateManager.cache.GetSnapshot(nodeGroup)
//	assert.NotEqual(t, testAction.Snapshot, &snapshotCache)
//	assert.Nil(t, snapshotCache.(*cache.Snapshot).VersionMap)
//}
//
//func TestUpdateManagerUpdateSnapshot_shouldSetSnapshot_whenNoErrors(t *testing.T) {
//	updateManager := getUpdateManager()
//
//	testAction := testType{
//		Snapshot: getSnapshot(),
//		Err:      nil,
//	}
//	nodeGroup := "1"
//
//	err := updateManager.UpdateSnapshot(nodeGroup, testAction)
//	assert.Nil(t, err)
//	snapshotCache, _ := updateManager.cache.GetSnapshot(nodeGroup)
//	assert.Equal(t, testAction.Snapshot, &snapshotCache)
//}
//
//func getMockUpdateAction(ctrl *gomock.Controller) *mock_action.MockUpdateActionFactory {
//	mockUpdateAction := mock_action.NewMockUpdateActionFactory(ctrl)
//	return mockUpdateAction
//}
//
//func getMockSnapshotUpdateAction(ctrl *gomock.Controller) *mock_action.MockSnapshotUpdateAction {
//	mockSnapshotUpdateAction := mock_action.NewMockSnapshotUpdateAction(ctrl)
//	return mockSnapshotUpdateAction
//}
//
//func getMockDao(ctrl *gomock.Controller) *mock_dao.MockDao {
//	mockDao := mock_dao.NewMockDao(ctrl)
//	mockDao.EXPECT().WithWTx(gomock.Any()).AnyTimes().Return(nil, nil)
//	return mockDao
//}
//
//func getMockChangeEventParser(ctrl *gomock.Controller) *mock_event.MockChangeEventParser {
//	mockChangeEventParser := mock_event.NewMockChangeEventParser(ctrl)
//	return mockChangeEventParser
//}
//
//func getMockActionsMap(ctrl *gomock.Controller) *mock_action.MockActionsMap {
//	mockActionsMap := mock_action.NewMockActionsMap(ctrl)
//	return mockActionsMap
//}
//
//func getMockActionsByNodeGroup(ctrl *gomock.Controller) *mock_event.MockActionsByNodeGroup {
//	mockActionsByNodeGroup := mock_event.NewMockActionsByNodeGroup(ctrl)
//	return mockActionsByNodeGroup
//}
//
//func getUpdateManagerMocks(ctrl *gomock.Controller) (*mock_dao.MockDao, *mock_action.MockUpdateActionFactory, *mock_action.MockSnapshotUpdateAction, *mock_event.MockChangeEventParser, *mock_action.MockActionsMap) {
//	return getMockDao(ctrl), getMockUpdateAction(ctrl), getMockSnapshotUpdateAction(ctrl), getMockChangeEventParser(ctrl), getMockActionsMap(ctrl)
//}
//
//func getUpdateManagerWithMocks(dao dao.Dao, updateAction action.UpdateActionFactory, eventParser event.ChangeEventParser) *UpdateManager {
//	cache := &snapshotCache{
//		Snapshots: map[string]cache.ResourceSnapshot{},
//	}
//	props := &common.EnvoyProxyProperties{
//		Routes: &common.RouteProperties{},
//	}
//	defaultListenerBuilder := &StubConfigBuilder{}
//
//	envoyConfigBuilder := builder.NewEnvoyConfigBuilder(dao, props, defaultListenerBuilder, nil)
//
//	return &UpdateManager{
//		dao:                dao,
//		cache:              cache,
//		envoyConfigBuilder: envoyConfigBuilder,
//		updateAction:       updateAction,
//		eventParser:        eventParser,
//		nodeGroupLocks:     &sync.Map{},
//	}
//}
//
//func getUpdateManager() *UpdateManager {
//	//// create an instance of our test object
//	//mockDao := new(test.MockedDao)
//	//// setup expectations
//	//mockDao.On("DoSomething", 123).Return(true, nil)
//
//	mockDao := dao.NewInMemDao(ram.NewStorage(), &GeneratorMock{}, nil)
//	v1 := &domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage}
//	mockDao.WithWTx(func(dao dao.Repository) error {
//		return dao.SaveDeploymentVersion(v1)
//	})
//	cache := &snapshotCache{
//		Snapshots: map[string]cache.ResourceSnapshot{},
//	}
//	props := &common.EnvoyProxyProperties{
//		Routes: &common.RouteProperties{},
//	}
//	defaultListenerBuilder := &StubConfigBuilder{}
//
//	envoyConfigBuilder := builder.NewEnvoyConfigBuilder(mockDao, props, defaultListenerBuilder, nil)
//
//	return NewUpdateManager(mockDao, cache, envoyConfigBuilder)
//}
//
//type StubConfigBuilder struct {
//}
//
//func (ecb *StubConfigBuilder) BuildListener(listener *domain.Listener, namespaceMapping string, withTls bool) (*v3listener.Listener, error) {
//	return &v3listener.Listener{
//		Name: "",
//	}, nil
//}
//
//func getSnapshot() cache.ResourceSnapshot {
//	snapshot := &cache.Snapshot{
//		Resources: [types.UnknownType]cache.Resources{
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-1": {},
//					"resource-2": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-3": {},
//					"resource-4": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-5": {},
//					"resource-6": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-7": {},
//					"resource-8": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-9":  {},
//					"resource-10": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-11": {},
//					"resource-12": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-13": {},
//					"resource-14": {},
//				},
//			},
//			{
//				Version: "1",
//				Items: map[string]types.ResourceWithTTL{
//					"resource-15": {},
//					"resource-16": {},
//				},
//			},
//		},
//		VersionMap: map[string]map[string]string{
//			"namespace1": {
//				"/fromA": "/toA",
//			},
//			"namespace2": {
//				"/fromB": "/toB",
//			},
//		},
//	}
//	return snapshot
//}
//
//type testType struct {
//	Snapshot cache.ResourceSnapshot
//	Err      error
//}
//
//func (t testType) Perform(original *cache.Snapshot) (*cache.Snapshot, error) {
//	return t.Snapshot.(*cache.Snapshot), t.Err
//}
//
//type snapshotCache struct {
//	Snapshots map[string]cache.ResourceSnapshot
//}
//
//func (c *snapshotCache) SetSnapshot(ctx context.Context, node string, snapshot cache.ResourceSnapshot) error {
//	c.Snapshots[node] = snapshot
//	return nil
//}
//
//func (c *snapshotCache) GetSnapshot(node string) (cache.ResourceSnapshot, error) {
//	if node == "error" {
//		return &cache.Snapshot{}, errors.New("error")
//	}
//	return c.Snapshots[node], nil
//}
//
//func (c *snapshotCache) ClearSnapshot(node string) {
//}
//
//func (c *snapshotCache) GetStatusInfo(node string) cache.StatusInfo {
//	return nil
//}
//
//func (c *snapshotCache) GetStatusKeys() []string {
//	return nil
//}
//
//func (c *snapshotCache) CreateWatch(request *cache.Request, streamState stream.StreamState, value chan cache.Response) func() {
//	return nil
//}
//
//func (c *snapshotCache) CreateDeltaWatch(request *cache.DeltaRequest, state stream.StreamState, value chan cache.DeltaResponse) func() {
//	return nil
//}
//
//func (c *snapshotCache) Fetch(ctx context.Context, request *cache.Request) (cache.Response, error) {
//	return nil, nil
//}
//
//type GeneratorMock struct {
//	counter int32
//}
//
//func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
//	if uniqEntity.GetId() == 0 {
//		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
//	}
//	return nil
//}
