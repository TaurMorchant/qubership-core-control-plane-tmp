package cache

import (
	"context"
	"errors"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/action"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/builder"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/builder/common"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache/event"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"sync"
	"time"
)

var (
	logger         logging.Logger
	runtimeVersion string
)

func init() {
	logger = logging.GetLogger("update-manager")
	runtimeVersion = string(rune(time.Now().Nanosecond()))
}

type UpdateManager struct {
	dao                dao.Dao
	cache              cache.SnapshotCache
	envoyConfigBuilder builder.EnvoyConfigBuilder
	updateAction       action.UpdateActionFactory
	eventParser        event.ChangeEventParser
	nodeGroupLocks     *sync.Map
}

func NewUpdateManager(dao dao.Dao, cache cache.SnapshotCache, envoyConfigBuilder builder.EnvoyConfigBuilder) *UpdateManager {
	updateActionFactory := action.NewUpdateActionFactory(envoyConfigBuilder)
	return &UpdateManager{
		dao:                dao,
		cache:              cache,
		envoyConfigBuilder: envoyConfigBuilder,
		updateAction:       updateActionFactory,
		eventParser:        event.NewChangeEventParser(dao, updateActionFactory, envoyConfigBuilder),
		nodeGroupLocks:     &sync.Map{},
	}
}

func (cacheManager *UpdateManager) UpdateSnapshot(nodeGroup string, action action.SnapshotUpdateAction) error {
	logger.Debugf("Start updating snapshot for group %v with action %#v", nodeGroup, action)
	lock := cacheManager.getNodeGroupLock(nodeGroup)
	lock.Lock()
	defer lock.Unlock()

	snapshot, err := cacheManager.cache.GetSnapshot(nodeGroup)
	if err != nil {
		snapshot = cacheManager.generateEmptySnapshot("0")
		_ = cacheManager.cache.SetSnapshot(context.Background(), nodeGroup, snapshot)
	}
	newSnapshot, err := action.Perform(snapshot.(*cache.Snapshot))
	if err != nil {
		logger.Errorf("Failed to update envoy cache snapshot (nodeGroup: %v, action: %v): %v", nodeGroup, action, err)
	} else if err = cacheManager.cache.SetSnapshot(context.Background(), nodeGroup, newSnapshot); err != nil {
		logger.Errorf("Failed to set new envoy cache snapshot (nodeGroup: %v, action: %v): %v", nodeGroup, action, err)
	}
	return err
}

func (cacheManager *UpdateManager) HandleChangeEvent(changeEvent *events.ChangeEvent) {
	actions := cacheManager.eventParser.ParseChangeEvent(changeEvent)
	if err := cacheManager.UpdateSnapshot(changeEvent.NodeGroup, actions.CompositeAction()); err != nil {
		logger.Errorf("Failed to update envoy cache snapshot for nodeGroup %v: %v", changeEvent.NodeGroup, err)
	}
}

// HandleMultipleChangeEvent handles MultipleChangeEvent which is an event with changes for multiple node groups.
func (cacheManager *UpdateManager) HandleMultipleChangeEvent(changeEvent *events.MultipleChangeEvent) {
	actionsByNodeGroup := cacheManager.eventParser.ParseMultipleChangeEvent(changeEvent)
	cacheManager.performUpdateActions(actionsByNodeGroup)
}

func (cacheManager *UpdateManager) HandlePartialReloadEvent(event *events.PartialReloadEvent) {
	logger.Debugf("Handling PartialReloadEvent: %v", event)
	actionsByNodeGroup := cacheManager.eventParser.ParsePartialReloadEvent(event)
	cacheManager.performUpdateActions(actionsByNodeGroup)
}

func (cacheManager *UpdateManager) performUpdateActions(actionsByNodeGroup map[string]action.SnapshotUpdateAction) {
	for nodeGroup, updateAction := range actionsByNodeGroup {
		if err := cacheManager.UpdateSnapshot(nodeGroup, updateAction); err != nil {
			logger.Errorf("Failed to update envoy cache snapshot for nodeGroup %v: %v", nodeGroup, err)
		}
	}
}

// TODO change init to Reload or reconstruct method
func (cacheManager *UpdateManager) HandleReloadEvent() {
	cacheManager.initConfigWithRecovery()
}

func (cacheManager *UpdateManager) UpdateListener(nodeGroup string, version int64, value *domain.Listener, namespaceMapping string) error {
	updateAction := action.NewListenerUpdateAction(cacheManager.envoyConfigBuilder, common.VersionToString(version), value, namespaceMapping)
	return cacheManager.UpdateSnapshot(nodeGroup, updateAction)
}

func (cacheManager *UpdateManager) BulkUpdateClusters(version int64, clusters []*domain.Cluster, updateActionProvider action.UpdateActionProvider) error {
	ver := common.VersionToString(version)
	actionsByNodeGroup := make(map[string][]action.SnapshotUpdateAction)
	for _, cluster := range clusters {
		if cluster.NodeGroups == nil {
			nodeGroups, err := cacheManager.dao.FindNodeGroupsByCluster(cluster)
			if err != nil {
				logger.Errorf("Failed to load nodeGroups for cluster %v using DAO: %v", cluster.Name, err)
				return err
			}
			cluster.NodeGroups = nodeGroups
		}
		for _, node := range cluster.NodeGroups {
			nodeGroup := node.Name
			if actionsByNodeGroup[nodeGroup] == nil {
				actionsByNodeGroup[nodeGroup] = make([]action.SnapshotUpdateAction, 0)
			}
			updateAction := updateActionProvider(nodeGroup, ver, cluster)
			actionsByNodeGroup[nodeGroup] = append(actionsByNodeGroup[nodeGroup], updateAction)
		}
	}
	var resultErr error
	resultErr = nil
	for nodeGroup, actions := range actionsByNodeGroup {
		updateAction := action.NewCompositeUpdateAction(actions)
		if err := cacheManager.UpdateSnapshot(nodeGroup, updateAction); err != nil {
			resultErr = errors.New("cache: failed to update some clusters")
			logger.Errorf("Failed to update clusters in envoy config for node group %v: %v", nodeGroup, err)
		}
	}
	return resultErr
}

func (cacheManager *UpdateManager) InitConfigWithRetry() {
	maxAttempts := 40 // TODO: smarter retry
	for curAttempt := 1; curAttempt <= maxAttempts; curAttempt++ {
		logger.Debugf("Attempting to initialize envoy configuration (attempt %v of %v)", curAttempt, maxAttempts)
		if cacheManager.initConfigWithRecovery() {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	logger.Panicf("Envoy configuration initialization has failed after %v retries", maxAttempts)
}

func (cacheManager *UpdateManager) initConfigWithRecovery() bool {
	defer func() bool {
		if recoveryMessage := recover(); recoveryMessage != nil {
			// TODO: might be not WARN
			logger.Warnf("Recovered envoy initial configuration panic: %v", recoveryMessage)
		}
		return false
	}()
	cacheManager.InitConfig()
	return true
}

func (cacheManager *UpdateManager) InitConfig() {
	logger.Debugf("Starting to initialize envoy configuration")
	nodeGroups, err := cacheManager.dao.FindAllNodeGroups()
	if err != nil {
		logger.Panicf("Failed to load initial node groups: %v", err)
	}
	actionsByNodeGroup := make(map[string][]action.SnapshotUpdateAction, len(nodeGroups))
	for _, nodeGroup := range nodeGroups {
		if err := cacheManager.envoyConfigBuilder.RegisterGateway(nodeGroup); err != nil {
			logger.Panicf("Update Manager failed to register gateway configuration builder for %+v", *nodeGroup)
		}

		actionsByNodeGroup[nodeGroup.Name] = make([]action.SnapshotUpdateAction, 0)

		if listenerActions := cacheManager.loadListenersForNodeGroup(nodeGroup); listenerActions != nil {
			actionsByNodeGroup[nodeGroup.Name] = append(actionsByNodeGroup[nodeGroup.Name], listenerActions...)
		}

		if clusterActions := cacheManager.loadClustersForNodeGroup(nodeGroup); clusterActions != nil {
			actionsByNodeGroup[nodeGroup.Name] = append(actionsByNodeGroup[nodeGroup.Name], clusterActions...)
		}

		if routeCfgActions := cacheManager.loadRouteConfigsForNodeGroup(nodeGroup); routeCfgActions != nil {
			actionsByNodeGroup[nodeGroup.Name] = append(actionsByNodeGroup[nodeGroup.Name], routeCfgActions...)
		}

		if runtimeActions := cacheManager.loadRuntimesForNodeGroup(nodeGroup); runtimeActions != nil {
			actionsByNodeGroup[nodeGroup.Name] = append(actionsByNodeGroup[nodeGroup.Name], runtimeActions...)
		}
	}
	for nodeGroup, actions := range actionsByNodeGroup {
		updateAction := action.NewCompositeUpdateAction(actions)
		if err := cacheManager.UpdateSnapshot(nodeGroup, updateAction); err != nil {
			logger.Panicf("Failed to update whole envoy config for node group %v: %v", nodeGroup, err)
		}
	}
	logger.Infof("Envoy configuration has been initialized successfully")
}

func (cacheManager *UpdateManager) loadListenersForNodeGroup(nodeGroup *domain.NodeGroup) []action.SnapshotUpdateAction {
	listeners, err := cacheManager.dao.FindListenersByNodeGroupId(nodeGroup.Name)
	if err != nil {
		logger.Panicf("Failed to load initial listeners: %v", err)
	}
	if len(listeners) == 0 {
		return nil
	}
	configVersion, err := cacheManager.dao.FindEnvoyConfigVersion(nodeGroup.Name, domain.ListenerTable)
	if err != nil {
		logger.Panicf("Failed to load envoy config version from DAO: %v", err)
	}
	if configVersion == nil {
		configVersion = domain.NewEnvoyConfigVersion(nodeGroup.Name, domain.ListenerTable)
		_, err := cacheManager.dao.WithWTx(func(dao dao.Repository) error {
			return dao.SaveEnvoyConfigVersion(configVersion)
		})
		if err != nil {
			return nil
		}
	}
	version := common.VersionToString(configVersion.Version)
	actions := make([]action.SnapshotUpdateAction, len(listeners))
	for i, listener := range listeners {
		actions[i] = cacheManager.updateAction.ListenerUpdateAction(version, listener)
	}
	return actions
}

func (cacheManager *UpdateManager) loadClustersForNodeGroup(nodeGroup *domain.NodeGroup) []action.SnapshotUpdateAction {
	clusters, err := cacheManager.dao.FindClusterByNodeGroup(nodeGroup)
	if err != nil {
		logger.Panicf("Failed to load initial clusters: %v", err)
	}
	if len(clusters) == 0 {
		return nil
	}
	configVersion, err := cacheManager.dao.FindEnvoyConfigVersion(nodeGroup.Name, domain.ClusterTable)
	if err != nil {
		logger.Panicf("Failed to load envoy config version from DAO: %v", err)
	}
	if configVersion == nil {
		configVersion = domain.NewEnvoyConfigVersion(nodeGroup.Name, domain.ClusterTable)
		_, err := cacheManager.dao.WithWTx(func(dao dao.Repository) error {
			return dao.SaveEnvoyConfigVersion(configVersion)
		})
		if err != nil {
			return nil
		}
	}
	version := common.VersionToString(configVersion.Version)
	actions := make([]action.SnapshotUpdateAction, len(clusters))
	for i, cluster := range clusters {
		actions[i] = cacheManager.updateAction.ClusterUpdateAction(nodeGroup.Name, version, cluster)
	}
	return actions
}

func (cacheManager *UpdateManager) loadRouteConfigsForNodeGroup(nodeGroup *domain.NodeGroup) []action.SnapshotUpdateAction {
	routeConfigs, err := cacheManager.dao.FindRouteConfigsByNodeGroupId(nodeGroup.Name)
	if err != nil {
		logger.Panicf("Failed to load initial route configs: %v", err)
	}
	if len(routeConfigs) == 0 {
		return nil
	}
	configVersion, err := cacheManager.dao.FindEnvoyConfigVersion(nodeGroup.Name, domain.RouteConfigurationTable)
	if err != nil {
		logger.Panicf("Failed to load envoy config version from DAO: %v", err)
	}
	if configVersion == nil {
		configVersion = domain.NewEnvoyConfigVersion(nodeGroup.Name, domain.RouteConfigurationTable)
		_, err := cacheManager.dao.WithWTx(func(dao dao.Repository) error {
			return dao.SaveEnvoyConfigVersion(configVersion)
		})
		if err != nil {
			return nil
		}
	}
	version := common.VersionToString(configVersion.Version)
	actions := make([]action.SnapshotUpdateAction, len(routeConfigs))
	for i, routeConfig := range routeConfigs {
		actions[i] = cacheManager.updateAction.RouteConfigUpdateAction(version, routeConfig)
	}
	return actions
}

func (cacheManager *UpdateManager) loadRuntimesForNodeGroup(nodeGroup *domain.NodeGroup) []action.SnapshotUpdateAction {
	actions := make([]action.SnapshotUpdateAction, 1)
	actions[0] = cacheManager.updateAction.RuntimeUpdateAction(nodeGroup.Name)
	return actions
}

func (cacheManager *UpdateManager) generateEmptySnapshot(version string) cache.ResourceSnapshot {
	var clusters, endpoints, routes, listeners, runtimes, secrets []types.Resource
	snapshot, err := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.EndpointType: endpoints,
			resource.ClusterType:  clusters,
			resource.RouteType:    routes,
			resource.ListenerType: listeners,
			resource.RuntimeType:  runtimes,
			resource.SecretType:   secrets,
		})
	if err != nil {
		logger.Panicf("Failed to generate empty snapshot:\n %v", err)
	}
	return snapshot
}

func (cacheManager *UpdateManager) getNodeGroupLock(nodeGroup string) *sync.Mutex {
	mutex, _ := cacheManager.nodeGroupLocks.LoadOrStore(nodeGroup, &sync.Mutex{})
	return mutex.(*sync.Mutex)
}
