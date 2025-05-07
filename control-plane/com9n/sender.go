package com9n

import (
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/netcracker/qubership-core-control-plane/data"
	"github.com/netcracker/qubership-core-control-plane/util"
	"sync/atomic"
	"time"
)

type DataSender struct {
	storage           data.RestorableStorage
	snapshotPublisher ActivatablePublisher
	tickerChan        chan int
	quitChan          chan int
	changes           int32
	retryProvider     RetryProvider
	stopped           bool
}

func NewDataSender(storage data.RestorableStorage, publisher ActivatablePublisher) *DataSender {
	sender := DataSender{
		storage:           storage,
		snapshotPublisher: publisher,
		changes:           0,
		tickerChan:        make(chan int),
		quitChan:          make(chan int),
		retryProvider:     util.DefaultRetryProvider{},
		stopped:           true,
	}
	return &sender
}

func (s *DataSender) HandleChangeEvent(data interface{}) {
	atomic.AddInt32(&s.changes, 1)
}

func (s *DataSender) IsStarted() bool {
	return !s.stopped
}

func (s *DataSender) StartSending(info clustering.NodeInfo) error {
	err := s.snapshotPublisher.Activate(info.BusAddress())
	if err != nil {
		return err
	}
	s.startTickerLoop()
	go func() {
		for {
			select {
			case <-s.tickerChan:
				if s.changes > 0 {
					if s.snapshotPublisher.HasSubscribers() {
						snapshot, _ := s.storage.Backup()
						err := s.snapshotPublisher.Publish(Topic, snapshot)
						if err != nil {
							log.Error("Sending snapshot caused error: %v", err)
						}
						log.Debug("In-memory snapshot has been sent successfully")
						atomic.StoreInt32(&s.changes, 0)
					} else {
						log.Debugf("There is no subscribers on snapshot. Do nothing.")
					}
				}
			case <-s.quitChan:
				log.Debugf("Sender of snapshots has been stopped")
				return
			}
		}
	}()
	s.stopped = false
	log.Debugf("Sender of snapshots has been started")
	return nil
}

func (s *DataSender) startTickerLoop() {
	go func() {
		for {
			s.tickerChan <- 0
			time.Sleep(s.retryProvider.SleepPeriodOnSnapshotSend())
		}
	}()
}

func (s *DataSender) StopSending() error {
	s.snapshotPublisher.Shutdown()
	s.quitChan <- 0
	s.stopped = true
	return nil
}
