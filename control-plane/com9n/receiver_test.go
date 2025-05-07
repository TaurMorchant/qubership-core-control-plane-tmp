package com9n

import (
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/netcracker/qubership-core-control-plane/data"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReceiver_FailedStartReceiving(t *testing.T) {
	master1NodeInfo := clustering.NodeInfo{
		IP:       "0.0.0.0",
		SWIMPort: 12345,
		BusPort:  54321,
		HttpPort: 8080,
	}

	subscriber := SubscriberMock{
		isActivated:  false,
		isSubscribed: false,
	}

	receiver := DataReceiver{
		snapshotSubscriber: &subscriber,
		snapshotProvider:   &ProviderMock{},
		snapshotsChan:      nil,
		quitChan:           nil,
		stopped:            true,
	}

	receiver.StartReceiving(master1NodeInfo.BusAddress())
	assert.False(t, receiver.IsStarted())
	assert.False(t, subscriber.isActivated)
	assert.False(t, subscriber.isSubscribed)
}

type SubscriberMock struct {
	isActivated  bool
	isSubscribed bool
}

func (s *SubscriberMock) Activate(address string) error {
	s.isActivated = true
	return nil
}

func (s *SubscriberMock) Subscribe(topic string, handler func(data interface{})) {
	s.isSubscribed = true
}

func (s *SubscriberMock) Shutdown() {
	s.isSubscribed = false
	s.isActivated = false
}

type ProviderMock struct {
	SubscriberMock
}

func (p *ProviderMock) GetSnapshot() (*data.Snapshot, error) {
	return nil, errors.New("Getting snapshot failed")
}
