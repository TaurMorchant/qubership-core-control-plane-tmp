package com9n

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/data"
)

type DataReceiver struct {
	snapshotProcessingFunc func(snapshot *data.Snapshot) error
	snapshotSubscriber     ActivatableSubscriber
	snapshotProvider       RemoteSnapshotProvider
	snapshotsChan          chan *data.Snapshot
	quitChan               chan int
	stopped                bool
}

func NewDataReceiver(
	snapshotProcessingFunc func(snapshot *data.Snapshot) error,
	snapshotSubscriber ActivatableSubscriber,
	snapshotProvider RemoteSnapshotProvider) *DataReceiver {
	return &DataReceiver{
		snapshotProcessingFunc: snapshotProcessingFunc,
		snapshotSubscriber:     snapshotSubscriber,
		snapshotProvider:       snapshotProvider,
		snapshotsChan:          make(chan *data.Snapshot),
		quitChan:               make(chan int),
		stopped:                true}
}

func (r *DataReceiver) IsStarted() bool {
	return !r.stopped
}

func (r *DataReceiver) StartReceiving(busAddress string) error {
	log.Infof("Starting snapshots receiver from address %v", busAddress)
	err := r.snapshotSubscriber.Activate(busAddress)
	if err != nil {
		return err
	}
	r.snapshotSubscriber.Subscribe(Topic, func(binData interface{}) {
		r.snapshotsChan <- binData.(*data.Snapshot)
	})
	snapshot, err := r.snapshotProvider.GetSnapshot()
	if err != nil {
		r.snapshotSubscriber.Shutdown()
		return err
	}
	go func() {
		for {
			select {
			case snapshot := <-r.snapshotsChan:
				r.snapshotProcessingFunc(snapshot)
				break
			case <-r.quitChan:
				log.Debugf("Receiver of snapshots has been stopped")
				return
			}
		}
	}()
	r.snapshotsChan <- snapshot
	r.stopped = false
	log.Infof("Receiver of snapshots has been started")
	return nil
}

func (r *DataReceiver) StopReceiving() error {
	log.Infof("Stopping snapshots receiver")
	r.snapshotSubscriber.Shutdown()
	r.quitChan <- 0
	r.stopped = true
	log.Infof("Receiver of snapshots has been stopped")
	return nil
}
