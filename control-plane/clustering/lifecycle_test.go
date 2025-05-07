package clustering

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_AppendFatal_InitMasterError(t *testing.T) {
	CleanFatalErrors()
	AppendFatal(fmt.Errorf("some error"))
	AppendFatal(NewInitMasterError(fmt.Errorf("init error 1")))
	AppendFatal(NewInitMasterError(fmt.Errorf("init error 2")))
	assert.Equal(t, 2, len(GetFatalErrors()))
}

func Test_GetFatalErrorsExceptInitMaster_NoInitMasterError(t *testing.T) {
	CleanFatalErrors()
	AppendFatal(fmt.Errorf("some error"))
	AppendFatal(NewInitMasterError(fmt.Errorf("init error")))
	assert.Equal(t, 1, len(GetFatalErrorsExceptInitErrors()))
}

func Test_InitMasterError(t *testing.T) {
	CleanFatalErrors()
	err := NewInitMasterError(fmt.Errorf("init error"))
	assert.Equal(t, "can't initialize master: init error", err.Error())
	assert.True(t, isInitError(err))
	assert.True(t, isInitMasterError(err))
}

func Test_InitCallbackError(t *testing.T) {
	CleanFatalErrors()
	err := NewCallbacksError(fmt.Errorf("callback error"))
	assert.Equal(t, "can't execute callback in node: callback error", err.Error())
	assert.True(t, isInitError(err))
	assert.False(t, isInitMasterError(err))
}

func Test_defineNodeRoleByNewMaster(t *testing.T) {
	CurrentNodeState.ChangeNodeState(NodeInfo{
		IP:       "10.20.30.40",
		SWIMPort: 10,
		BusPort:  20,
		HttpPort: 30,
	}, Initial)
	CleanFatalErrors()

	nodeName := "node-name-1"
	namespace := "default"
	nodInfo := NodeInfo{
		IP:       "1.1.1.1",
		SWIMPort: 10,
		BusPort:  20,
		HttpPort: 30,
	}

	initer := &StubMasterNodeInitializer{}
	lcm := NewLifeCycleManager(nodeName, namespace, nodInfo, initer)
	newMaster := &MasterMetadata{
		Id:   100,
		Name: "node-name-1",
		NodeInfo: NodeInfo{
			IP:       "1.1.1.1",
			SWIMPort: 10,
			BusPort:  20,
			HttpPort: 30,
		},
		SyncClock: time.Time{},
	}
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Master, CurrentNodeState.GetRole())
	assert.Equal(t, "1.1.1.1:30", CurrentNodeState.GetHttpAddress())
	assert.True(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 0, len(GetFatalErrors()))
	assert.Equal(t, newMaster, lcm.currentMaster)

	newMaster = &MasterMetadata{
		Id:   150,
		Name: "node-name-2",
		NodeInfo: NodeInfo{
			IP:       "2.2.2.2",
			SWIMPort: 10,
			BusPort:  20,
			HttpPort: 30,
		},
		SyncClock: time.Time{},
	}
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Slave, CurrentNodeState.GetRole())
	assert.Equal(t, "2.2.2.2:30", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 0, len(GetFatalErrors()))
	assert.Equal(t, newMaster, lcm.currentMaster)

	initer.returnError = true
	newMaster = &MasterMetadata{
		Id:   110,
		Name: "node-name-1",
		NodeInfo: NodeInfo{
			IP:       "1.1.1.1",
			SWIMPort: 10,
			BusPort:  20,
			HttpPort: 30,
		},
		SyncClock: time.Time{},
	}
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Master, CurrentNodeState.GetRole())
	assert.Equal(t, "1.1.1.1:30", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 1, len(GetFatalErrors()))
	assert.Equal(t, newMaster, lcm.currentMaster)

	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Master, CurrentNodeState.GetRole())
	assert.Equal(t, "1.1.1.1:30", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 1, len(GetFatalErrors()))
	assert.Equal(t, newMaster, lcm.currentMaster)

	initer.returnError = false
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Master, CurrentNodeState.GetRole())
	assert.Equal(t, "1.1.1.1:30", CurrentNodeState.GetHttpAddress())
	assert.True(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 0, len(GetFatalErrors()))
	assert.Equal(t, newMaster, lcm.currentMaster)

	lcm.defineNodeRole(nil)
	assert.Equal(t, Phantom, CurrentNodeState.GetRole())
	assert.Equal(t, ":0", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.Equal(t, 0, len(GetFatalErrors()))
	assert.Nil(t, lcm.currentMaster)
}

type StubMasterNodeInitializer struct {
	returnError bool
}

func (s *StubMasterNodeInitializer) InitMaster() error {
	if s.returnError {
		return fmt.Errorf("something wrong")
	}
	return nil
}

func Test_CallBackErrorHandling(t *testing.T) {
	CurrentNodeState.ChangeNodeState(NodeInfo{
		IP:       "10.20.30.40",
		SWIMPort: 10,
		BusPort:  20,
		HttpPort: 30,
	}, Initial)
	CleanFatalErrors()

	nodeName := "node-name-1"
	namespace := "default"
	nodInfo := NodeInfo{
		IP:       "1.1.1.1",
		SWIMPort: 10,
		BusPort:  20,
		HttpPort: 30,
	}

	initer := &StubMasterNodeInitializer{}
	lcm := NewLifeCycleManager(nodeName, namespace, nodInfo, initer)
	newMaster := &MasterMetadata{
		Id:   150,
		Name: "node-name-2",
		NodeInfo: NodeInfo{
			IP:       "2.2.2.2",
			SWIMPort: 10,
			BusPort:  20,
			HttpPort: 30,
		},
		SyncClock: time.Time{},
	}

	errors := fmt.Errorf("some error during callback")
	lcm.AddOnRoleChanged(func(info NodeInfo, role Role) error {
		assert.Equal(t, Slave, role)
		return errors
	})
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Slave, CurrentNodeState.GetRole())
	assert.Equal(t, "2.2.2.2:30", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.True(t, HasInitErrorInNode())
	assert.Equal(t, 1, len(GetFatalErrors()))
	assert.Equal(t, "can't execute callback in node: some error during callback", GetFatalErrors()[0].Error())
	assert.Equal(t, newMaster, lcm.currentMaster)

	errors = nil // clear error
	lcm.defineNodeRole(newMaster)
	assert.Equal(t, Slave, CurrentNodeState.GetRole())
	assert.Equal(t, "2.2.2.2:30", CurrentNodeState.GetHttpAddress())
	assert.False(t, CurrentNodeState.IsMasterReady())
	assert.False(t, HasInitErrorInNode())
	assert.Equal(t, newMaster, lcm.currentMaster)
}
