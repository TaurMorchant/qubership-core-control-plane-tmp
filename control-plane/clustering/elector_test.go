package clustering

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

type TestEnvironment struct {
	electionService  *ElectionServiceMock
	lifeCycleManager *LifeCycleManager
	electorConfig    *ElectorConfig
}

func initElectorTestEnvironment() (TestEnvironment, error) {
	electionService := NewElectionServiceMock()
	err := electionService.tryAddInternalRecord()
	if err != nil {
		return TestEnvironment{}, err
	}

	lifeCycleManager := NewLifeCycleManager(
		"test-node",
		"default",
		NodeInfo{
			IP:       "0.0.0.0",
			SWIMPort: 0,
			BusPort:  0,
			HttpPort: 8080,
		},
		MasterNodeInitializerMock{},
	)

	electorConfig := ElectorConfig{
		ElectionService:     electionService,
		LifeCycleManager:    lifeCycleManager,
		MasterLoopSleepTime: 1 * time.Second,
		SlaveLoopSleepTime:  1 * time.Second,
		SyncClockTimeShift:  1 * time.Second,
	}

	return TestEnvironment{
		electionService:  electionService,
		lifeCycleManager: lifeCycleManager,
		electorConfig:    &electorConfig,
	}, nil
}

func TestElector_ElectorCreation(t *testing.T) {
	electorConfig := ElectorConfig{}
	_, err := NewElector(electorConfig)
	assert.NotNil(t, err)
	electionService := NewElectionServiceMock()
	electorConfig = ElectorConfig{
		ElectionService: electionService,
	}
	_, err = NewElector(electorConfig)
	assert.NotNil(t, err)
	lifeCycleManager := NewLifeCycleManager("", "", NodeInfo{}, nil)
	electorConfig = ElectorConfig{
		ElectionService:  electionService,
		LifeCycleManager: lifeCycleManager,
	}
	elector, err := NewElector(electorConfig)
	assert.Nil(t, err)
	assert.NotNil(t, elector)
}

func TestElector_TestMasterElection(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	testEnv.lifeCycleManager.thisNode.Name = "master-node"
	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	masterCalls := make([]NodeInfo, 0)
	callLock := sync.Mutex{}
	roleChangedCallback := func(info NodeInfo, role Role) error {
		callLock.Lock()
		defer callLock.Unlock()
		switch role {
		case Master:
			masterCalls = append(masterCalls, info)
		}
		return nil
	}
	testEnv.lifeCycleManager.AddOnRoleChanged(roleChangedCallback)

	err = elector.Start()
	assert.Nil(t, err)

	timeout := time.Now().Add(2 * time.Minute)
	for len(masterCalls) == 0 {
		time.Sleep(1 * time.Second)
		if time.Now().After(timeout) {
			break
		}
	}
	elector.Stop()

	assert.NotEmpty(t, masterCalls)

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestElector_TestSlaveElection(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	testEnv.lifeCycleManager.thisNode.Name = "master-node"
	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	elector.config.ElectionService.TryWriteAsMaster(&elector.record)
	testEnv.lifeCycleManager.thisNode.Name = "slave-node"
	elector.record.Name = "slave-node"

	slaveCalls := make([]NodeInfo, 0)
	callLock := sync.Mutex{}
	roleChangedCallback := func(info NodeInfo, role Role) error {
		callLock.Lock()
		defer callLock.Unlock()
		switch role {
		case Slave:
			slaveCalls = append(slaveCalls, info)
		}
		return nil
	}
	testEnv.lifeCycleManager.AddOnRoleChanged(roleChangedCallback)

	err = elector.Start()
	assert.Nil(t, err)

	timeout := time.Now().Add(2 * time.Minute)
	for len(slaveCalls) == 0 {
		time.Sleep(1 * time.Second)
		if time.Now().After(timeout) {
			break
		}
	}
	elector.Stop()

	assert.NotEmpty(t, slaveCalls)

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestElector_TestMasterElection_Compound(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	electorConfig := ElectorConfig{
		testEnv.electorConfig.ElectionService,
		testEnv.electorConfig.LifeCycleManager,
		15 * time.Second,
		testEnv.electorConfig.MasterLoopSleepTime,
		SyncClockTimeShiftDefault,
	}

	testEnv.lifeCycleManager.thisNode.Name = "master-node"
	elector, err := NewElector(electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	masterCalls := make([]NodeInfo, 0)
	slaveCalls := make([]NodeInfo, 0)
	callLock := sync.Mutex{}
	roleChangedCallback := func(info NodeInfo, role Role) error {
		callLock.Lock()
		defer callLock.Unlock()
		switch role {
		case Master:
			masterCalls = append(masterCalls, info)
		case Slave:
			slaveCalls = append(slaveCalls, info)
		}
		return nil
	}
	testEnv.lifeCycleManager.AddOnRoleChanged(roleChangedCallback)

	err = elector.Start()
	assert.Nil(t, err)

	timeout := time.Now().Add(electorConfig.SlaveLoopSleepTime / 2)
	for len(masterCalls) == 0 {
		time.Sleep(1 * time.Second)
		if time.Now().After(timeout) {
			break
		}
	}

	assert.NotEmpty(t, masterCalls)
	testEnv.lifeCycleManager.thisNode.Name = "slave-node"
	elector.record.Name = "slave-node"

	timeout = time.Now().Add(electorConfig.SlaveLoopSleepTime / 2)
	for CurrentNodeState.IsMaster() {
		time.Sleep(1 * time.Second)
		if time.Now().After(timeout) {
			break
		}
	}

	elector.Stop()

	assert.NotEmpty(t, slaveCalls)
	assert.False(t, CurrentNodeState.IsMaster())

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestElector_TestSlaveElection_NoRecords(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	testEnv.electionService.clearTable()

	testEnv.lifeCycleManager.thisNode.Name = "slave-node"
	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	phantomCalls := make([]NodeInfo, 0)
	callLock := sync.Mutex{}
	roleChangedCallback := func(info NodeInfo, role Role) error {
		callLock.Lock()
		defer callLock.Unlock()
		switch role {
		case Phantom:
			phantomCalls = append(phantomCalls, info)
		}
		return nil
	}
	testEnv.lifeCycleManager.AddOnRoleChanged(roleChangedCallback)

	err = elector.Start()
	assert.Nil(t, err)

	timeout := time.Now().Add(2 * time.Minute)
	for len(phantomCalls) == 0 {
		time.Sleep(1 * time.Second)
		if time.Now().After(timeout) {
			break
		}
	}
	elector.Stop()

	assert.NotEmpty(t, phantomCalls)

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestElector_TestMasterProlongation(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	elector.config.ElectionService.TryWriteAsMaster(&elector.record)
	*elector.lifeCycleManager.currentMaster = elector.record

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := &MasterMetadata{}
	*masterRecordBefore = *masterRecord

	// Stop elector on timeout
	go func() {
		time.Sleep(2 * time.Second)
		elector.Stop()
	}()
	// Prolong
	elector.startMasterProlongation()

	masterRecordAfter, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	assert.NotEqual(t, masterRecordBefore.SyncClock, masterRecordAfter.SyncClock)
	*masterRecordBefore = *masterRecordAfter
	elector.stopped = false
}

func TestElector_TestMasterProlongation_NoRecords(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	elector.config.ElectionService.TryWriteAsMaster(&elector.record)
	*elector.lifeCycleManager.currentMaster = elector.record

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := &MasterMetadata{}
	*masterRecordBefore = *masterRecord

	// No master record case
	testEnv.electionService.clearTable()
	elector.startMasterProlongation()

	_, err = elector.config.ElectionService.GetMaster()
	assert.NotNil(t, err)
}

func TestElector_TestMasterProlongation_MasterHasChangedBreak(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	elector.config.ElectionService.TryWriteAsMaster(&elector.record)
	*elector.lifeCycleManager.currentMaster = elector.record

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := &MasterMetadata{}
	*masterRecordBefore = *masterRecord

	// Master has changed break case
	elector.config.LifeCycleManager.currentMaster.Name = "other-master-node"
	elector.startMasterProlongation()

	masterRecordAfter, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	assert.Equal(t, masterRecordBefore.SyncClock, masterRecordAfter.SyncClock)
}

func TestElector_TestMasterProlongation_ShiftSyncClockBreak(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	elector.config.ElectionService.TryWriteAsMaster(&elector.record)
	*elector.lifeCycleManager.currentMaster = elector.record

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := MasterMetadata{}
	masterRecordBefore = *masterRecord

	// Shift sync clock break case
	elector.config.SyncClockTimeShift = 0 * time.Second // To invoke error
	elector.startMasterProlongation()

	masterRecordAfter, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	assert.Equal(t, masterRecordBefore.SyncClock, masterRecordAfter.SyncClock)
}

func TestElector_ForceElection(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector.ticker = make(chan interface{})

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := &MasterMetadata{}
	*masterRecordBefore = *masterRecord

	go elector.ForceElection("internal")

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			masterRecordAfter, err := elector.config.ElectionService.GetMaster()
			assert.Nil(t, err)
			assert.NotEqual(t, masterRecordBefore.SyncClock, masterRecordAfter.SyncClock)
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func TestElector_ForceElection_NoRecords(t *testing.T) {
	testEnv, err := initElectorTestEnvironment()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector, err := NewElector(*testEnv.electorConfig)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	elector.ticker = make(chan interface{})

	masterRecord, err := elector.config.ElectionService.GetMaster()
	assert.Nil(t, err)
	masterRecordBefore := &MasterMetadata{}
	*masterRecordBefore = *masterRecord

	testEnv.electionService.clearTable()
	go elector.ForceElection("internal")

	// wait for election loop
	for {
		select {
		case <-elector.ticker:
			_, err = elector.config.ElectionService.GetMaster()
			assert.NotNil(t, err)
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

type MasterNodeInitializerMock struct {
}

func (i MasterNodeInitializerMock) InitMaster() error {
	return nil
}

type ElectionServiceMock struct {
	masterMetadata []*MasterMetadata
}

func NewElectionServiceMock() *ElectionServiceMock {
	return &ElectionServiceMock{
		masterMetadata: []*MasterMetadata{},
	}
}

func (e *ElectionServiceMock) GetMaster() (*MasterMetadata, error) {
	if len(e.masterMetadata) == 0 {
		return nil, fmt.Errorf("no rows in db")
	}
	return e.masterMetadata[0], nil
}

func (e *ElectionServiceMock) ResetSyncClock(master string) error {
	if len(e.masterMetadata) == 0 {
		return fmt.Errorf("no rows in db")
	}
	for _, metadata := range e.masterMetadata {
		if metadata.Name == master {
			metadata.SyncClock = time.Now()
		}
	}
	return nil
}

func (e *ElectionServiceMock) TryWriteAsMaster(electionRecord *MasterMetadata) bool {
	rowsAffected := 0
	curTime := time.Now()
	for _, metadata := range e.masterMetadata {
		if metadata.SyncClock.Before(curTime) {
			*metadata = *electionRecord
			metadata.SyncClock = curTime.Add(60 * time.Second)
			rowsAffected++
		}
	}
	return rowsAffected != 0
}

func (e *ElectionServiceMock) ShiftSyncClock(d time.Duration) error {
	if d < 1*time.Second {
		// Purposed only for error test
		return fmt.Errorf("duration less than a second")
	}
	if len(e.masterMetadata) == 0 {
		return fmt.Errorf("no rows in db")
	}
	e.masterMetadata[0].SyncClock = time.Now().Add(d)
	return nil
}

func (e *ElectionServiceMock) DeleteSeveralRecordsFromDb() error {
	var newMetadata []*MasterMetadata
	for _, metadata := range e.masterMetadata {
		if metadata.Name == "internal" {
			newMetadata = append(newMetadata, metadata)
		}
	}
	e.masterMetadata = newMetadata

	//err := e.tryAddInternalRecord()
	//if err != nil {
	//	return err
	//}

	return nil
}

func (e *ElectionServiceMock) tryAddInternalRecord() error {
	count := len(e.masterMetadata)
	if count == 1 {
		return nil
	}

	if count > 1 {
		e.masterMetadata = []*MasterMetadata{}
	}

	electionRecord := MasterMetadata{
		Name: "internal",
		NodeInfo: NodeInfo{
			IP:       "0.0.0.0",
			SWIMPort: 0,
			BusPort:  0,
		},
	}
	e.masterMetadata = append(e.masterMetadata, &electionRecord)
	return nil
}

func (e *ElectionServiceMock) clearTable() {
	e.masterMetadata = make([]*MasterMetadata, 0)
}
