package action

import (
	"context"
	"errors"
	v3cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	v3listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	v3route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	v3runtime "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/domain"
	mock_builder "github.com/netcracker/qubership-core-control-plane/test/mock/envoy/cache/builder"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/reflect/protoreflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRuntimeUpdateActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	nodeGroup := resource.EndpointType

	envoyConfigBuilder := getMockEnvoyConfigBuilder(ctrl)
	envoyConfigBuilder.EXPECT().BuildRuntime(gomock.Any(), gomock.Any()).Return(&v3runtime.Runtime{}, nil)

	runtimeUpdateAction := NewRuntimeUpdateAction(envoyConfigBuilder, nodeGroup)

	result, err := runtimeUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, snapshot.Resources[types.Runtime].Items["rtds_layer0"].Resource)
	assert.NotEqual(t, 1, snapshot.Resources[types.Runtime].Version)
}

func TestDeleteAllByTypeActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	version := "2"
	envoyTypeURL := resource.EndpointType

	deleteByTypeAction := NewDeleteAllByTypeAction(version, envoyTypeURL)

	result, err := deleteByTypeAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, version, snapshot.Resources[types.Endpoint].Version)
	assert.Equal(t, 0, len(snapshot.Resources[types.Endpoint].Items))
}

func TestRouteConfigDeleteActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	version := snapshot.Resources[types.Route].Version
	newRouteConfig := &domain.RouteConfiguration{
		Id:   int32(0),
		Name: "resource-1",
	}

	routeConfigDeleteAction := NewRouteConfigDeleteAction(version, newRouteConfig)

	result, err := routeConfigDeleteAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, snapshot.Resources[types.Route].Items[newRouteConfig.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Route].Version)
}

func TestRouteConfigUpdateActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	envoyConfigBuilder := getMockEnvoyConfigBuilder(ctrl)
	envoyConfigBuilder.EXPECT().BuildRouteConfig(gomock.Any()).Return(&v3route.RouteConfiguration{}, nil)
	version := snapshot.Resources[types.Route].Version
	newRouteConfig := &domain.RouteConfiguration{
		Id:   int32(0),
		Name: "resource-new",
	}

	routeConfigUpdateAction := NewRouteConfigUpdateAction(envoyConfigBuilder, version, newRouteConfig)

	result, err := routeConfigUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, snapshot.Resources[types.Route].Items[newRouteConfig.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Route].Version)
}

func TestClusterDeleteActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	version := snapshot.Resources[types.Route].Version
	newCluster := &domain.Cluster{
		Id:   int32(0),
		Name: "resource-1",
	}

	clusterDeleteAction := NewClusterDeleteAction(version, newCluster)

	result, err := clusterDeleteAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, snapshot.Resources[types.Cluster].Items[newCluster.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Cluster].Version)
}

func TestNewGenericDeleteAction_ConcurrentAccessToResource(t *testing.T) {
	original := getSnapshot()
	clusterResources := original.GetResourcesAndTTL(resource.ClusterType)
	ctx, cancelF := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelF()

	go func() {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for range clusterResources {
					i = i + 1
				}
			}
		}
	}()

	wd := &sync.WaitGroup{}
	wd.Add(1)
	go func() {
		var newSnapshot *cache.Snapshot
		var err error
		for {
			select {
			case <-ctx.Done():
				assert.Nil(t, err)
				assert.NotNil(t, newSnapshot)
				assert.Nil(t, newSnapshot.Resources[types.Cluster].Items["resource-1"].Resource)
				wd.Done()
				return
			default:
				newSnapshot, err = NewGenericDeleteAction("resource-1", resource.ClusterType).Perform(original)
			}
		}
	}()

	wd.Wait()
}

func TestClusterUpdateActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeGroup := resource.EndpointType
	snapshot := getSnapshot()
	envoyConfigBuilder := getMockEnvoyConfigBuilder(ctrl)
	envoyConfigBuilder.EXPECT().BuildCluster(nodeGroup, gomock.Any()).Return(&v3cluster.Cluster{}, nil)
	version := snapshot.Resources[types.Route].Version
	newCluster := &domain.Cluster{
		Id:   int32(0),
		Name: "resource-new",
	}

	clusterUpdateAction := NewClusterUpdateAction(envoyConfigBuilder, nodeGroup, version, newCluster)

	result, err := clusterUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, snapshot.Resources[types.Cluster].Items[newCluster.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Cluster].Version)
}

func TestNewGenericUpdateAction_ConcurrentAccessToResource(t *testing.T) {
	cl := &v3cluster.Cluster{}
	original := getSnapshot()
	clusterResources := original.GetResourcesAndTTL(resource.ClusterType)
	ctx, cancelF := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelF()

	go func() {
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			default:
				for range clusterResources {
					i = i + 1
				}
			}
		}
	}()

	wd := &sync.WaitGroup{}
	wd.Add(1)
	go func() {
		var newSnapshot *cache.Snapshot
		var err error
		for {
			select {
			case <-ctx.Done():
				assert.Nil(t, err)
				assert.NotNil(t, newSnapshot.Resources[types.Cluster].Items["resource-new"].Resource)
				wd.Done()
				return
			default:
				newSnapshot, err = NewGenericUpdateAction("resource-new", resource.ClusterType, cl).Perform(original)
			}
		}
	}()

	wd.Wait()
}

func TestListenerDeleteActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	version := snapshot.Resources[types.Route].Version
	listener := &domain.Listener{Id: int32(0), Name: "resource-1"}

	listenerUpdateAction := NewListenerDeleteAction(version, listener)

	result, err := listenerUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, snapshot.Resources[types.Listener].Items[listener.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Listener].Version)
}

func TestListenerUpdateActionPerform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	snapshot := getSnapshot()
	envoyConfigBuilder := getMockEnvoyConfigBuilder(ctrl)
	envoyConfigBuilder.EXPECT().BuildListener(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v3listener.Listener{}, nil)
	version := snapshot.Resources[types.Route].Version
	listener := &domain.Listener{Id: int32(0), Name: "resource-new"}
	namespaceMapping := "test"

	listenerUpdateAction := NewListenerUpdateAction(envoyConfigBuilder, version, listener, namespaceMapping)

	result, err := listenerUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, snapshot.Resources[types.Listener].Items[listener.Name].Resource)
	assert.NotEqual(t, version, snapshot.Resources[types.Listener].Version)
}

func TestGenericUpdateActionPerform(t *testing.T) {
	snapshot := getSnapshot()

	resourceName2 := "resource-2"
	envoyTypeURL := resource.ClusterType
	resource := &EmptyResource{
		Name: "test",
	}

	genericUpdateAction := NewGenericUpdateAction(resourceName2, envoyTypeURL, resource)
	assert.NotNil(t, genericUpdateAction)

	assert.Nil(t, snapshot.Resources[1].Items[resourceName2].Resource)

	result, err := genericUpdateAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result.Resources[0].Items[resourceName2].Resource)
	assert.Equal(t, resource, result.Resources[0].Items[resourceName2].Resource)
}

func TestGenericDeleteActionPerform(t *testing.T) {
	snapshot := getSnapshot()

	nextVersion := snapshot.Resources[types.Route].Version
	resourceName1 := "resource-1"
	resourceName2 := "resource-2"
	resourceName3 := "resource-3"
	envoyTypeURL := resource.ClusterType
	genericDeleteAction := NewGenericDeleteAction(resourceName2, envoyTypeURL)
	_, exist := snapshot.Resources[1].Items[resourceName2]
	assert.True(t, exist)

	result, err := genericDeleteAction.Perform(snapshot)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	_, exist = snapshot.Resources[0].Items[resourceName2]
	assert.False(t, exist)

	_, exist = snapshot.Resources[0].Items[resourceName3]
	assert.False(t, exist)

	_, exist = snapshot.Resources[1].Items[resourceName1]
	assert.True(t, exist)

	_, exist = snapshot.Resources[1].Items[resourceName2]
	assert.True(t, exist)

	assert.NotEqual(t, nextVersion, snapshot.Resources[0].Version)
}

func TestCompositeUpdateActionPerform_shouldReturnError_whenErrorPresent(t *testing.T) {
	testAction := testType{
		Snapshot: getSnapshot(),
		Err:      errors.New("test error"),
	}
	compositeUpdateAction := NewCompositeUpdateAction([]SnapshotUpdateAction{testAction})

	result, err := compositeUpdateAction.Perform(&cache.Snapshot{})
	assert.Nil(t, result.VersionMap)
	assert.NotNil(t, err)
	assert.Equal(t, ErrSomeActionsFailed, err)
}

func TestCompositeUpdateActionPerform_shouldReturnSnapshot_whenNoErrors(t *testing.T) {
	testAction := testType{
		Snapshot: getSnapshot(),
		Err:      nil,
	}
	compositeUpdateAction := NewCompositeUpdateAction([]SnapshotUpdateAction{testAction})

	result, err := compositeUpdateAction.Perform(&cache.Snapshot{})
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, testAction.Snapshot.VersionMap, result.VersionMap)
}

func getSnapshot() *cache.Snapshot {
	resourceName1 := "resource-1"
	resourceName2 := "resource-2"

	snapshot := &cache.Snapshot{
		Resources: [types.UnknownType]cache.Resources{
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
			{
				Version: "1",
				Items: map[string]types.ResourceWithTTL{
					resourceName1: {},
					resourceName2: {},
				},
			},
		},
		VersionMap: map[string]map[string]string{
			"namespace1": {
				"/fromA": "/toA",
			},
			"namespace2": {
				"/fromB": "/toB",
			},
		},
	}
	return snapshot
}

func getMockEnvoyConfigBuilder(ctrl *gomock.Controller) *mock_builder.MockEnvoyConfigBuilder {
	mockEnvoyConfigBuilder := mock_builder.NewMockEnvoyConfigBuilder(ctrl)
	return mockEnvoyConfigBuilder
}

type testType struct {
	Snapshot *cache.Snapshot
	Err      error
}

func (t testType) Perform(original *cache.Snapshot) (*cache.Snapshot, error) {
	return t.Snapshot, t.Err
}

type EmptyResource struct {
	Name string
}

func (x *EmptyResource) Reset() {
}

func (x *EmptyResource) String() string {
	return ""
}

func (*EmptyResource) ProtoMessage() {}

func (x *EmptyResource) ProtoReflect() protoreflect.Message {
	return nil
}

func (*EmptyResource) Descriptor() ([]byte, []int) {
	return nil, nil
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
