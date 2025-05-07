package clustering

import (
	"fmt"
	"time"
)

const (
	MasterLoopSleepTimeDefault   = 15 * time.Second
	SyncClockTimeShiftDefault    = 30 * time.Second
	ElectionLoopSleepTimeDefault = SyncClockTimeShiftDefault
)

type ElectorConfig struct {
	ElectionService     ElectionService
	LifeCycleManager    *LifeCycleManager
	SlaveLoopSleepTime  time.Duration
	MasterLoopSleepTime time.Duration
	SyncClockTimeShift  time.Duration
}

type Elector struct {
	record           MasterMetadata
	config           *ElectorConfig
	lifeCycleManager *LifeCycleManager
	ticker           chan interface{}
	stopped          bool
}

func NewElector(config ElectorConfig) (*Elector, error) {
	elector := &Elector{}
	if config.ElectionService == nil {
		return nil, fmt.Errorf("ElectionService is required")
	}

	if config.LifeCycleManager == nil {
		return nil, fmt.Errorf("LifeCycleManager is required")
	}

	if config.SlaveLoopSleepTime == 0 {
		config.SlaveLoopSleepTime = ElectionLoopSleepTimeDefault
	}

	if config.MasterLoopSleepTime == 0 {
		config.MasterLoopSleepTime = MasterLoopSleepTimeDefault
	}

	if config.SyncClockTimeShift == 0 {
		config.SyncClockTimeShift = SyncClockTimeShiftDefault
	}

	elector.record = config.LifeCycleManager.GetThisNodeMetadata()
	elector.lifeCycleManager = config.LifeCycleManager
	elector.config = &config
	return elector, nil
}

func (e *Elector) Start() error {
	e.ticker = make(chan interface{})
	e.stopped = false
	go e.startElectionLoop()
	go e.startElectionTicker()
	return nil
}

func (e *Elector) Stop() {
	log.InfoC(ctx, "Election has been stopped")
	e.stopped = true
}

func (e *Elector) startElectionTicker() {
	for {
		time.Sleep(e.config.SlaveLoopSleepTime)
		e.ticker <- struct{}{}
	}
}

func (e *Elector) startElectionLoop() {
	for !e.stopped {
		log.DebugC(ctx, "start election")
		curRecord := &e.record
		if ok := e.config.ElectionService.TryWriteAsMaster(curRecord); ok {
			log.DebugC(ctx, "current node has been saved as master: %v ", curRecord)
			e.lifeCycleManager.defineNodeRole(curRecord)
			if HasInitErrorInNode() {
				log.Warnf("node has got init master error")
				time.Sleep(e.config.SlaveLoopSleepTime) // let's give a chance to another node
				continue
			}
			e.startMasterProlongation()
			continue
		} else {
			masterRecord, err := e.config.ElectionService.GetMaster()
			log.DebugC(ctx, "master record: %v", masterRecord)
			if err != nil {
				log.Warnf("Can't get info about master node. Cause: %v", err)
			}
			e.lifeCycleManager.defineNodeRole(masterRecord)
		}

		<-e.ticker
	}
}

func (e *Elector) startMasterProlongation() {
	log.Infof("Master prolongation has started.")
	for !e.stopped {
		time.Sleep(e.config.MasterLoopSleepTime)
		masterRecord, err := e.config.ElectionService.GetMaster()
		if err != nil {
			log.Warnf("Can't prolong master record. Falling back to election. Cause: %v", err)
			break
		}
		if masterRecord.Name != e.config.LifeCycleManager.currentMaster.Name ||
			masterRecord.NodeInfo != e.config.LifeCycleManager.currentMaster.NodeInfo {
			log.Warnf("Master has changed. Falling back to election.")
			log.Debug("New master record: %v", masterRecord)
			break
		}
		err = e.config.ElectionService.ShiftSyncClock(e.config.SyncClockTimeShift)
		if err != nil {
			log.Warnf("Can't prolong master record. Falling back to election. Cause: %v", err)
			break
		}
	}
}

func (e *Elector) ForceElection(masterNodeName string) {
	log.Debugf("Forced election of master '%s'.", masterNodeName)
	if err := e.config.ElectionService.ResetSyncClock(masterNodeName); err != nil {
		log.Errorf("Error during reset sync clock: '%v'", err)
	}
	e.ticker <- struct{}{}
}
