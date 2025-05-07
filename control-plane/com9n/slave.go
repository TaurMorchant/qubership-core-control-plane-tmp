package com9n

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/data"
)

func NewSlaveNodeDataReceiver(storage data.RestorableStorage, subscriber ActivatableSubscriber, provider RemoteSnapshotProvider, configUpdater ConfigUpdater) *DataReceiver {
	snapshotProcessor := NewSlaveNodeSnapshotProcessor(storage, configUpdater)
	return NewDataReceiver(snapshotProcessor.ProcessSnapshot, subscriber, provider)
}

type SlaveNodeSnapshotProcessor struct {
	storage       data.RestorableStorage
	configUpdater ConfigUpdater
}

func NewSlaveNodeSnapshotProcessor(storage data.RestorableStorage, configUpdater ConfigUpdater) *SlaveNodeSnapshotProcessor {
	return &SlaveNodeSnapshotProcessor{storage: storage, configUpdater: configUpdater}
}

func (p *SlaveNodeSnapshotProcessor) ProcessSnapshot(snapshot *data.Snapshot) error {
	log.Debug("In-memory storage snapshot has been received")
	if err := p.storage.Restore(*snapshot); err != nil {
		return fmt.Errorf("restoring data caused error: %w", err)
	}
	p.configUpdater.InitConfigWithRetry()
	return nil
}
