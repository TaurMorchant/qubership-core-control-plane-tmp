package bus

import (
	"context"
	"github.com/google/uuid"
	"github.com/mustafaturan/bus/v3"
	"github.com/mustafaturan/monoton/v3"
	"github.com/mustafaturan/monoton/v3/sequencer"
	"sync"
)

var instance *InternalBusImpl
var once sync.Once

func init() {
	// init bus and register initial topics
	GetInternalBusInstance()
}

// InternalBusImpl provides implementation of the internal event bus used by EventBusAggregator. Implements both
// BusPublisher and BusSubscriber interfaces so the same instance is used for publishing and subscribing.
type InternalBusImpl struct {
	bus *bus.Bus
}

func GetInternalBusInstance() *InternalBusImpl {
	once.Do(func() {
		// configure id generator (it doesn't have to be monotone)
		node := uint64(1)
		initialTime := uint64(0)
		m, err := monoton.New(sequencer.NewMillisecond(), node, initialTime)
		if err != nil {
			log.Errorf("Failed to create monoton ID generator: %v", err)
			panic(err)
		}

		var idGenerator bus.Next = m.Next

		// configure bus
		b, err := bus.NewBus(idGenerator)
		if err != nil {
			log.Errorf("Failed to configure internal event bus (mustafaturan implementation): %v", err)
			panic(err)
		}
		b.RegisterTopics(TopicChanges)
		b.RegisterTopics(TopicBgRegistry)
		b.RegisterTopics(TopicMultipleChanges)
		b.RegisterTopics(TopicReload)
		b.RegisterTopics(TopicPartialReapply)

		instance = &InternalBusImpl{b}
	})
	return instance
}

func (pub *InternalBusImpl) Publish(topic string, data interface{}) error {
	return pub.bus.Emit(context.Background(), topic, data)
}

func (sub *InternalBusImpl) Subscribe(topic string, handler func(data interface{})) {
	adaptedHandler := bus.Handler{Handle: mustafaTuranEventHandler(handler), Matcher: topic}
	registryKey, err := uuid.NewRandom()
	if err != nil {
		log.Errorf("Failed to generate mustafaturan bus registry key (uuid): %v", err)
		panic(err)
	}
	sub.bus.RegisterHandler(registryKey.String(), adaptedHandler)
}

func mustafaTuranEventHandler(handler func(data interface{})) func(ctx context.Context, e bus.Event) {
	return func(ctx context.Context, e bus.Event) {
		handler(e.Data)
	}
}

func (sub *InternalBusImpl) Shutdown() {
	for _, handlerKey := range sub.bus.HandlerKeys() {
		sub.bus.DeregisterHandler(handlerKey)
	}
	for _, topic := range sub.bus.Topics() {
		sub.bus.DeregisterTopics(topic)
	}
}
