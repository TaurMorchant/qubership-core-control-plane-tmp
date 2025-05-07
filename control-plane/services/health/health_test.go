package health

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CheckReadiness_MasterNotReadyAndNoDataLoaded_NotReady(t *testing.T) {
	service := HealthService{}
	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)
	result := service.CheckReadiness()
	assert.Equal(t, NotReady, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckReadiness_MasterNotReadyButHasLoadedData_NotReady(t *testing.T) {
	service := HealthService{}
	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)
	clustering.CurrentNodeState.SetMasterReady()
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)

	result := service.CheckReadiness()
	assert.Equal(t, Ready, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckReadiness_MasterReady_Ready(t *testing.T) {
	service := HealthService{}
	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)
	clustering.CurrentNodeState.SetMasterReady()

	result := service.CheckReadiness()
	assert.Equal(t, Ready, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckReadiness_MasterHasErrors_Ready(t *testing.T) {
	service := HealthService{}
	node := clustering.NodeInfo{}
	clustering.AppendFatal(fmt.Errorf("some error"))

	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)
	clustering.CurrentNodeState.SetMasterReady()

	result := service.CheckReadiness()
	assert.Equal(t, Ready, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckReadiness_SlaveWithReceiverStarted_Ready(t *testing.T) {
	service := NewHealthService(&ConfiguratorMock{true})
	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Slave)

	result := service.CheckReadiness()
	assert.Equal(t, Ready, result.Status)
	assert.Equal(t, clustering.Slave, result.Role)
}

type ConfiguratorMock struct {
	IsStarted bool
}

func (c ConfiguratorMock) SetUpNodesCommunication(info clustering.NodeInfo, role clustering.Role) error {
	return nil
}

func (c ConfiguratorMock) IsReceiverStarted() bool {
	return c.IsStarted
}

func Test_CheckReadiness_SlaveWithReceiverNotStarted_NotReady(t *testing.T) {
	service := NewHealthService(&ConfiguratorMock{false})
	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Slave)

	result := service.CheckReadiness()
	assert.Equal(t, NotReady, result.Status)
	assert.Equal(t, clustering.Slave, result.Role)
}

func Test_CheckReadiness_Phantom_Ready(t *testing.T) {
	service := NewHealthService(&ConfiguratorMock{true})
	clustering.AppendFatal(fmt.Errorf("some error"))

	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Phantom)

	result := service.CheckReadiness()
	assert.Equal(t, Ready, result.Status)
	assert.Equal(t, clustering.Phantom, result.Role)
}

func Test_CheckLiveness_NodeHasErrorButNotInitMasterError_Problem(t *testing.T) {
	service := NewHealthService(&ConfiguratorMock{true})
	clustering.AppendFatal(fmt.Errorf("some error"))

	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)

	result := service.CheckLiveness()
	assert.Equal(t, Problem, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckLiveness_NodeHasNoErrors_Up(t *testing.T) {
	clustering.CleanFatalErrors()
	service := NewHealthService(&ConfiguratorMock{true})

	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)

	result := service.CheckLiveness()
	assert.Equal(t, Up, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}

func Test_CheckLiveness_HasMaterInitError_Up(t *testing.T) {
	clustering.CleanFatalErrors()
	service := NewHealthService(&ConfiguratorMock{true})
	clustering.AppendFatal(clustering.NewInitMasterError(fmt.Errorf("some error")))

	node := clustering.NodeInfo{}
	clustering.CurrentNodeState.ChangeNodeState(node, clustering.Master)

	result := service.CheckLiveness()
	assert.Equal(t, Up, result.Status)
	assert.Equal(t, clustering.Master, result.Role)
}
