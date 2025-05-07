package bus

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/netcracker/qubership-core-control-plane/data"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
	"time"
)

const TestTopic = "changes"
const TestReadTopic = "read"

func TestSubscribersCache(t *testing.T) {
	cache := NewSubscribersCache()

	subs := make([]string, 100)
	for i := 0; i < 100; i++ {
		subs[i] = fmt.Sprintf("subscriber_%v", i)
	}
	for goroutineNum := 0; goroutineNum < 10; goroutineNum++ {
		go addTenSubsToCache(cache, subs, goroutineNum*10)
	}
	delay := 100 * time.Millisecond
	attemtsLeft := 1 * time.Minute / delay
	topicCacheSize := countCachedSubsForTopic(cache, TestTopic)
	for topicCacheSize < 100 && attemtsLeft > 0 {
		time.Sleep(delay)
		topicCacheSize = countCachedSubsForTopic(cache, TestTopic)
		attemtsLeft--
	}
	assert.Equal(t, 100, topicCacheSize)

	anotherTopicSubs := make([][]byte, 20)
	for i := 0; i < 20; i++ {
		anotherTopicSubs[i] = []byte(fmt.Sprintf("another_sub_representation_%v", i))
		cache.Add("AnotherTopic", anotherTopicSubs[i])
	}
	cachedTestTopicSubs := make([]string, 0)
	cachedAnotherTopicSubs := make([][]byte, 0)
	cache.ForEachSubInTopic(TestTopic, func(sub interface{}) bool {
		cachedTestTopicSubs = append(cachedTestTopicSubs, sub.(string))
		return true
	})
	cache.ForEachSubInTopic("AnotherTopic", func(sub interface{}) bool {
		cachedAnotherTopicSubs = append(cachedAnotherTopicSubs, sub.([]byte))
		return true
	})
	assertSlicesEqualIgnoreOrder(t, subs, cachedTestTopicSubs)
	assertSlicesEqualIgnoreOrder(t, anotherTopicSubs, cachedAnotherTopicSubs)

	cachedTestTopicSubs = make([]string, 0)
	cachedAnotherTopicSubs = make([][]byte, 0)
	cache.ForEach(func(topic string, subs []interface{}) bool {
		if topic == TestTopic {
			for _, sub := range subs {
				cachedTestTopicSubs = append(cachedTestTopicSubs, sub.(string))
			}
		} else if topic == "AnotherTopic" {
			for _, sub := range subs {
				cachedAnotherTopicSubs = append(cachedAnotherTopicSubs, sub.([]byte))
			}
		} else {
			t.Fatalf("Unexpected topic in subscribers cache: %v", topic)
		}
		return true
	})

	assertSlicesEqualIgnoreOrder(t, subs, cachedTestTopicSubs)
	assertSlicesEqualIgnoreOrder(t, anotherTopicSubs, cachedAnotherTopicSubs)

	for goroutineNum := 0; goroutineNum < 10; goroutineNum++ {
		go removeTenSubsFromCache(cache, subs, goroutineNum*10)
	}
	attemtsLeft = 1 * time.Minute / delay
	topicCacheSize = countCachedSubsForTopic(cache, TestTopic)
	for topicCacheSize > 0 && attemtsLeft > 0 {
		time.Sleep(delay)
		topicCacheSize = countCachedSubsForTopic(cache, TestTopic)
		attemtsLeft--
	}
	assert.Equal(t, 0, topicCacheSize)
}

func TestSubscribersCacheBreak(t *testing.T) {
	cache := NewSubscribersCache()
	requestedSubNunber := 1
	requestedSubName := fmt.Sprintf("subscriber_%v", requestedSubNunber)
	totalSubs := 10

	for i := 0; i < totalSubs; i++ {
		cache.Add(TestTopic, fmt.Sprintf("subscriber_%v", i))
	}
	topicCacheSize := countCachedSubsForTopic(cache, TestTopic)
	assert.Equal(t, totalSubs, topicCacheSize)

	var requestedSub interface{}
	cache.ForEachSubInTopic(TestTopic, func(sub interface{}) bool {
		if sub == requestedSubName {
			requestedSub = sub
			return false
		}
		cache.Add(TestReadTopic, sub)
		return true
	})
	assert.Equal(t, requestedSubName, requestedSub)

	readTopicCacheSize := countCachedSubsForTopic(cache, TestReadTopic)
	assert.Equal(t, requestedSubNunber, readTopicCacheSize)
}

func assertSlicesEqualIgnoreOrder(t *testing.T, slice1 interface{}, slice2 interface{}) {
	assert.Subset(t, slice1, slice2)
	assert.Subset(t, slice2, slice1)
}

func countCachedSubsForTopic(cache *SubscribersCache, topic string) int {
	subsNum := 0
	cache.ForEachSubInTopic(topic, func(sub interface{}) bool {
		subsNum++
		return true
	})
	return subsNum
}

func addTenSubsToCache(cache *SubscribersCache, allSubs []string, firstIdx int) {
	for i := firstIdx; i < firstIdx+10; i++ {
		cache.Add(TestTopic, allSubs[i])
	}
}

func removeTenSubsFromCache(cache *SubscribersCache, allSubs []string, firstIdx int) {
	for i := firstIdx; i < firstIdx+10; i++ {
		cache.Remove(TestTopic, allSubs[i])
	}
}

func TestBus(t *testing.T) {
	masterBusPort := findFreePort()
	log.Infof("Event bus test will be using free port %v", masterBusPort)
	nodeInf := clustering.NodeInfo{
		IP:       "127.0.0.1",
		SWIMPort: 0,
		BusPort:  uint16(masterBusPort),
	}

	syncChan := make(chan int32, 10)
	master := NewTestNode("master", nodeInf, clustering.Master, syncChan)
	slave1 := NewTestNode("slave1", nodeInf, clustering.Slave, syncChan)
	slave2 := NewTestNode("slave2", nodeInf, clustering.Slave, syncChan)
	// TODO get rid of using global object
	clustering.CurrentNodeState.SetMasterReady()

	master.Bus.Subscribe(TestTopic, master.HandleTestEvent)
	slave1.Bus.Subscribe(TestTopic, slave1.HandleTestEvent)
	slave2.Bus.Subscribe(TestTopic, slave2.HandleTestEvent)

	grpcServiceImpl := master.Bus.externalBusPublisher.(*GRPCBusPublisher).serviceImpl
	waitForGrpcSubscriptions(t, grpcServiceImpl)

	log.Info("Publishing test event...")
	if err := master.Bus.Publish(TestTopic, &data.Snapshot{}); err != nil {
		t.Fatalf("Test bus event publish error: %v", err)
	}

	expectedNotificationsNum := 5 // 3 internal handlers (cause really all 3 nodes use same local bus) + 2 slaves must be notified by gRPC
	notifiedSubsNum := 0
	for notifiedSubsNum < expectedNotificationsNum {
		select {
		case _ = <-syncChan:
			notifiedSubsNum++
		case <-time.After(10 * time.Second):
			t.Fatal("Event bus test failed by timeout")
		}
	}
	assert.Equal(t, expectedNotificationsNum, notifiedSubsNum)
}

func waitForGrpcSubscriptions(t *testing.T, serviceImpl *EventBusServerImpl) {
	delay := 300 * time.Millisecond
	attemptsLeft := 2 * time.Minute / delay
	var grpcSubscribersNum int
	for attemptsLeft > 0 {
		time.Sleep(300 * time.Millisecond)
		grpcSubscribersNum = 0
		serviceImpl.subs.ForEachSubInTopic(TestTopic, func(sub interface{}) bool {
			grpcSubscribersNum++
			return true
		})
		if grpcSubscribersNum >= 2 {
			break
		}
		attemptsLeft--
	}
	assert.Equal(t, 2, grpcSubscribersNum)
}

func findFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	return port
}

type TestNode struct {
	Name     string
	Role     clustering.Role
	Bus      *EventBusAggregator
	SyncChan chan int32
}

func NewTestNode(name string, masterNode clustering.NodeInfo, role clustering.Role, syncChan chan int32) *TestNode {
	internalBus := GetInternalBusInstance()
	grpcSub := NewGRPCBusSubscriber()
	grpcPub := NewGRPCBusPublisher(nil)
	bus := NewEventBusAggregator(nil, internalBus, internalBus, grpcSub, grpcPub)
	switch role {
	case clustering.Slave:
		if err := grpcSub.Activate(masterNode.BusAddress()); err != nil {
			panic(err)
		}
	case clustering.Master:
		if err := grpcPub.Activate(masterNode.BusAddress()); err != nil {
			panic(err)
		}
	default:
		panic("unknown role")
	}
	bus.RestartEventBus(masterNode, role)
	return &TestNode{Name: name, Role: role, Bus: bus, SyncChan: syncChan}
}

func (node *TestNode) ChangeRole(masterNode clustering.NodeInfo, role clustering.Role) {
	node.Role = role
	node.Bus.RestartEventBus(masterNode, role)
}

func (node *TestNode) HandleTestEvent(d interface{}) {
	log.Infof("Node %v (role: %v) received event: %v", node.Name, node.Role, d.(*data.Snapshot))
	node.SyncChan <- 1
}
