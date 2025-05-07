package bus

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/clustering"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/data"
	"sync"
)

// EventBusAggregator is event bus that aggregates internal and external buses (both publishers and subscribers)
// into the single API, so EventBusAggregator implements both BusPublisher and BusSubscriber interfaces.
//
// The idea of having two buses is that when we are running in Slave mode, we are obtaining events from external bus,
// and when we are running in Master mode, we are broadcasting events to both buses
// (our local listeners will be listening to internal bus and other nodes will obtain events by external bus).
type EventBusAggregator struct {
	mutex *sync.RWMutex
	// subscriptions contains slices of local event handlers (handler functions) mapped on topics.
	// It is used for re-subscribing on external bus in case external bus publisher has changed.
	subscriptions *SubscribersCache

	// runningAsMaster indicates whether external bus publisher is running.
	runningAsMaster bool
	// runningAsSlave indicates whether external bus subscriber is running.
	runningAsSlave bool

	externalBusPublisher  BusPublisher
	InternalBusPublisher  BusPublisher
	externalBusSubscriber BusSubscriber
	InternalBusSubscriber BusSubscriber

	storage data.RestorableStorage
}

func NewEventBusAggregator(storage data.RestorableStorage,
	internalBusSubscriber BusSubscriber, internalBusPublisher BusPublisher,
	externalBusSubscriber BusSubscriber, externalBusPublisher BusPublisher) *EventBusAggregator {
	return &EventBusAggregator{
		mutex:                 &sync.RWMutex{},
		subscriptions:         NewSubscribersCache(),
		runningAsMaster:       false,
		runningAsSlave:        false,
		InternalBusPublisher:  internalBusPublisher,
		InternalBusSubscriber: internalBusSubscriber,
		externalBusPublisher:  externalBusPublisher,
		externalBusSubscriber: externalBusSubscriber,
		storage:               storage,
	}
}

// RestartEventBus restarts aggregated event bus according to current cluster state.
// Can be safely used for the first event bus launch and for adaptation to self role change or Master address change.
func (evb *EventBusAggregator) RestartEventBus(info clustering.NodeInfo, role clustering.Role) error {
	evb.mutex.Lock()
	defer evb.mutex.Unlock()
	log.Infof("(Re)Starting event bus with role: %v", role)

	if role == clustering.Master {
		evb.StartAsMaster(&info)
	} else if role == clustering.Slave {
		evb.StartAsSlave(&info)
	} else {
		log.Errorf("Unsupported cluster node role: %v! Event bus will not be started", role)
	}

	return nil
}

// ResubscribeOnNewMaster iterates over cached local event handlers and creates external bus subscriptions for each of them.
// This function is called by event bus slave each time event bus master address is changed.
func (evb *EventBusAggregator) ResubscribeOnNewMaster() {
	evb.subscriptions.ForEach(func(topic string, handlers []interface{}) bool {
		for _, handler := range handlers {
			evb.externalBusSubscriber.Subscribe(topic, handler.(func(data interface{})))
		}
		return true
	})
}

func (evb *EventBusAggregator) StartAsMaster(masterNode *clustering.NodeInfo) {
	log.Info("Start bus aggregator as master")
	if evb.runningAsMaster {
		log.Infof("EventBus already running as master")
		return
	}

	evb.runningAsSlave = false
	evb.runningAsMaster = true
}

func (evb *EventBusAggregator) StartAsSlave(masterNode *clustering.NodeInfo) {
	log.Info("Start bus aggregator as slave")
	if evb.runningAsSlave {
		log.Infof("EventBus already running as slave")
		return
	}
	evb.runningAsMaster = false
	evb.runningAsSlave = true

	log.Infof("Subscribing on new master: %v", masterNode.BusAddress())
	evb.ResubscribeOnNewMaster()
}

// Publish publishes event for topic in internal and external buses. Note, that event is published in external bus
// only if current node is running in Master mode. If datatype is not supported, then Publish will be a no-operation
func (evb *EventBusAggregator) Publish(topic string, data interface{}) error {
	evb.mutex.RLock()
	defer evb.mutex.RUnlock()

	log.Debugf("Publishing on topic %v via bus aggregator", topic)

	// TODO: do we need to publish into internal bus if we are slave?
	if err := evb.InternalBusPublisher.Publish(topic, data); err != nil {
		log.Errorf("Bus aggregator internal publisher failed with error: %v", err)
		return err
	}

	if evb.runningAsMaster {
		if err := evb.externalBusPublisher.Publish(topic, data); err != nil {
			log.Errorf("Bus aggregator external publisher failed with error: %v", err)
			return err
		}

	}
	return nil
}

// Subscribe subscribes on topic in internal and external buses. Note, that event is published in external bus
// only if current node is running in Slave mode. If datatype is not supported, then Subscribe will be a no-operation
func (evb *EventBusAggregator) Subscribe(topic string, handler func(data interface{})) {
	evb.mutex.RLock()
	defer evb.mutex.RUnlock()

	evb.subscriptions.Add(topic, handler) // add handler to cache for re-subscribing in case of master address change

	evb.InternalBusSubscriber.Subscribe(topic, handler) // internal subscription will be useful when we become master
	if evb.runningAsSlave {
		evb.externalBusSubscriber.Subscribe(topic, handler)
	}
}

func (evb *EventBusAggregator) Shutdown() {
	evb.mutex.Lock()
	defer evb.mutex.Unlock()

	if evb.runningAsMaster {
		evb.externalBusPublisher.Shutdown()
		evb.runningAsMaster = false
	}
	if evb.runningAsSlave {
		evb.externalBusSubscriber.Shutdown()
		evb.runningAsSlave = false
	}
	evb.InternalBusPublisher.Shutdown()
}
