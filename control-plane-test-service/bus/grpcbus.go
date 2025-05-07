package bus

import (
	"context"
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"google.golang.org/grpc"
	"io"
	"net"
	"sync"
	"time"
)

// EventBusServerImpl provides implementation of EventBus gRPC service server stub.
//
// HOW-TO GENERATE GRPC STUBS:
// execute in control-plane-test-service/bus folder (protoc must be installed):
//
// $ protoc -I . test_bus.proto --go-grpc_out=. --go_out=.
type EventBusServerImpl struct {
	// subs holds subscriptions represented as channels for gRPCChannelMsg messages mapped on topics.
	// These channels are used to pass message (event) to thread that is communicating with specific subscriber,
	// so there is one channel for each subscriber on each his subscription.
	subs *SubscribersCache
	// chanBuffer specifies buffer size for every channel that represents subscription in subscribers cache.
	chanBuffer int
}

// gRPCChannelMsg is message that is passed to a subscription channel to be processed
// by thread that is communicating with specific subscriber.
type gRPCChannelMsg struct {
	// terminal indicates whether this is a message to terminate connection. If true, then event field is ignored.
	terminal bool
	// event holds event that needs to be sent to subscriber, serialized in protobuf.
	event *Event
}

var ErrConnGracefullyClosed = errors.New("bus: gRPC server gracefully closed connection")
var ErrSubscriberIsShutDown = errors.New("bus: gRPC subscriber is already shut down")

func NewEventBusServerImpl(chanBuffer int) *EventBusServerImpl {
	masterImpl := EventBusServerImpl{
		chanBuffer: chanBuffer,
		subs:       NewSubscribersCache(),
	}
	return &masterImpl
}

var Bus *GRPCBusPublisher

func init() {
	log.Infof("Starting gGRPC bus publisher on port 8888...")
	Bus = NewStartedGRPCBusPublisher("0.0.0.0:8888")
}

// Subscribe is server-side implementation of EventBus#Subscribe rpc defined in bus.proto.
func (s *EventBusServerImpl) Subscribe(topic *Topic, stream TestEventBus_SubscribeServer) error {
	log.Debugf("Received gRPC event bus subscription on topic %v", topic)
	// channel will be used by other threads to pass us messages so we can read them from channel and stream to the client
	ch := make(chan gRPCChannelMsg, s.chanBuffer)
	s.subs.Add(topic.Name, ch)
	for {
		// we need to block forever until terminal event comes,
		// because if we exit from Subscribe function, gRPC connection will be closed
		event := <-ch
		if event.terminal {
			log.Info("Shutting down gRPC server stream")
			s.subs.Remove(topic.Name, ch)
			return nil
		}
		if err := stream.Send(event.event); err != nil {
			log.Errorf("Failed grpc event %v sending on topic %v: %v", event, topic, err)
			return err
		}
	}
}

func (s *EventBusServerImpl) GetLastSnapshot(ctx context.Context, empty *Empty) (*Event, error) {
	return buildTestSnapshotEvent()
}

func (s *EventBusServerImpl) mustEmbedUnimplementedTestEventBusServer() {
}

// GRPCBusPublisher is a gRPC-based implementation of the BusPublisher interface.
type GRPCBusPublisher struct {
	serviceImpl *EventBusServerImpl
	gRPCServer  *grpc.Server
	active      bool
}

func NewGRPCBusPublisher() *GRPCBusPublisher {
	publisher := &GRPCBusPublisher{active: false}
	return publisher
}

func NewStartedGRPCBusPublisher(addr string) *GRPCBusPublisher {
	publisher := NewGRPCBusPublisher()
	_ = publisher.Activate(addr)
	return publisher
}

func (p *GRPCBusPublisher) Activate(address string) error {
	if !p.active {
		p.serviceImpl = NewEventBusServerImpl(10)
		log.Infof("Starting gRPC server on addr %v", address)

		lis, err := net.Listen("tcp", address)
		if err != nil {
			return errors2.Wrap(err, fmt.Sprintf("Failed to listen on gRPC addr %v", address))
		}
		p.gRPCServer = grpc.NewServer()
		RegisterTestEventBusServer(p.gRPCServer, p.serviceImpl)
		go func() {
			log.Infof("gRPC server shutdown: %v", p.gRPCServer.Serve(lis))
		}()
		p.active = true
	} else {
		return fmt.Errorf("attempt to activate gRPC publisher already activated with addr: %v", address)
	}
	return nil
}

func (p *GRPCBusPublisher) HasSubscribers() bool {
	return p.serviceImpl.subs.size > 0
}

func (p *GRPCBusPublisher) Publish(topic string, data interface{}) error {
	if p.active {
		log.Debugf("Publishing gRPC event bus message on topic %v", topic)
		marshalledData, err := buildProtobufEvent(data)
		if err != nil {
			log.Errorf("Failed to marshall data %v to protobuf: %v", data, err)
			return err
		}
		p.serviceImpl.subs.ForEachSubInTopic(topic, func(sub interface{}) bool {
			sub.(chan gRPCChannelMsg) <- gRPCChannelMsg{
				terminal: false,
				event:    marshalledData,
			}
			return true
		})
	} else {
		log.Warnf("Attempt to publish data to topic '%v' when publisher inactive", topic)
	}
	return nil
}

func (p *GRPCBusPublisher) Shutdown() {
	if p.active {
		p.serviceImpl.subs.ForEach(func(topic string, msgChanSlice []interface{}) bool {
			for _, msgChan := range msgChanSlice {
				// send terminal event to each subscription channel
				msgChan.(chan gRPCChannelMsg) <- gRPCChannelMsg{terminal: true}
			}
			return true
		})
		p.gRPCServer.GracefulStop()
		p.active = false
	} else {
		log.Warnf("Attempt to shutdown inactive gRPC publisher")
	}
}

// GRPCBusSubscriber is a gRPC-based implementation of the BusSubscriber interface.
type GRPCBusSubscriber struct {
	targetAddress string
	// Delay before retry of the client gRPC call
	retryDelay time.Duration
	mutex      *sync.Mutex
	clientConn *grpc.ClientConn
	// subscriberIsActive indicates that subscriber was not yet shut down
	// and needs to retry client connection in case of failure
	subscriberIsActive bool
	wg                 *sync.WaitGroup
}

func NewGRPCBusSubscriber() *GRPCBusSubscriber {
	return &GRPCBusSubscriber{
		retryDelay:         100 * time.Millisecond,
		mutex:              &sync.Mutex{},
		wg:                 &sync.WaitGroup{},
		subscriberIsActive: false}
}

func NewStartedGRPCBusSubscriber(addr string) *GRPCBusSubscriber {
	subscriber := NewGRPCBusSubscriber()
	_ = subscriber.Activate(addr)
	return subscriber
}

func (s *GRPCBusSubscriber) Subscribe(topic string, handler func(data interface{})) {
	if s.subscriberIsActive {
		go s.subscribeWithRetry(topic, handler)
	} else {
		log.Warnf("Attempt to subscribe to '%s' topic when gRPC subscriber is not active.", topic)
	}
}

func (s *GRPCBusSubscriber) Activate(address string) error {
	if !s.subscriberIsActive {
		s.targetAddress = address
		s.subscriberIsActive = true
	} else {
		return fmt.Errorf("attempt to activate gRPC suscriber already activated with addr: %v", address)
	}
	return nil
}

func (s *GRPCBusSubscriber) subscribeWithRetry(topic string, handler func(data interface{})) {
	for s.subscriberIsActive {
		s.wg.Add(1)
		s.subscribeWithRecovery(topic, handler)
		time.Sleep(s.retryDelay)
		s.wg.Done()
	}
}

func (s *GRPCBusSubscriber) subscribeWithRecovery(topic string, handler func(data interface{})) {
	defer s.recoverSubscriberPanic()
	s.subscribeInternal(topic, handler)
}

func (s *GRPCBusSubscriber) recoverSubscriberPanic() {
	if recoveryMessage := recover(); recoveryMessage != nil {
		log.Warnf("Recovered gRPC event s subscriber panic: %v", recoveryMessage)
	}
	log.Debug("Executed gRPC event s subscriber recovery function")
}

func (s *GRPCBusSubscriber) openConnection() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.subscriberIsActive {
		log.Debug("Tried to open gRPC connection for already shut down subscriber")
		panic(ErrSubscriberIsShutDown)
	}

	conn, err := grpc.Dial(s.targetAddress, grpc.WithInsecure(), grpc.WithDefaultServiceConfig(loadRetryConfig()))
	if err != nil {
		log.Errorf("Failed to dial client gRPC conn: %v", err)
		panic(err)
	}
	s.clientConn = conn
}

func (s *GRPCBusSubscriber) closeConnection() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.clientConn != nil {
		_ = s.clientConn.Close()
	}
}

func (s *GRPCBusSubscriber) subscribeInternal(topic string, handler func(data interface{})) {
	log.Infof("Subscribe on server gRPC events on topic %v from %s", topic, s.targetAddress)
	s.openConnection()
	defer s.closeConnection()

	client := NewTestEventBusClient(s.clientConn)
	stream, err := client.Subscribe(context.Background(), &Topic{Name: topic})
	if err != nil {
		log.Errorf("gRPC client failed to subscribe: %v", err)
		panic(err)
	}
	log.Infof("gRPC client successfully subscribed on topic %v", topic)

	for {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Infof("gRPC client received EOF: %v", err)
				panic(ErrConnGracefullyClosed)
			}
			log.Errorf("gRPC client failed to receive event: %v", err)
			panic(err)
		}
		data, err := readProtobufEvent(event)
		if err != nil {
			panic(err)
		}
		handler(data)
	}
}

// loadRetryConfig provides retry configuration for gRPC service "EventBus",
// BUT THIS RETRY POLICY IS USED ONLY ON BUSINESS LAYER!
// It means that in most cases (connectivity and availability issues) this retryPolicy is not applied so it's almost useless.
func loadRetryConfig() string {
	return `{
            "methodConfig": [{
                "name": [{"service": "EventBus"}],
                "waitForReady": true,

                "retryPolicy": {
                    "MaxAttempts": 100,
                    "InitialBackoff": ".01s",
                    "MaxBackoff": ".01s",
                    "BackoffMultiplier": 1.0,
                    "RetryableStatusCodes": [ "UNAVAILABLE", "DEADLINE_EXCEEDED", "CANCELLED", "UNKNOWN", "INTERNAL", "RESOURCE_EXHAUSTED" ]
                }
            }]
        }`
}

func (s *GRPCBusSubscriber) Shutdown() {
	if s.subscriberIsActive {
		s.subscriberIsActive = false // prevent subscribers from retrying when they lose connection after we close it.
		if s.clientConn != nil {
			log.Infof("Closed gRPC client conn: %v", s.clientConn.Close())
		}
		s.wg.Wait()
	} else {
		log.Warnf("Attempt to shutdown inactive gRPC subscriber")
	}
}
