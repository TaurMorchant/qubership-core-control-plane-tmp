package clustering

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotAppliesWhenCalledFirstlyAndNotAppliedWhenCalledWithTheSameMaster(t *testing.T) {
	masterNode := NodeInfo{IP: "127.0.0.1", SWIMPort: 1234, BusPort: 5431, HttpPort: 8080}
	role := Slave

	underlyingCallback := callback{lastcallRoleArg: Phantom}
	onMasterChange := ApplyOnMasterChange(underlyingCallback.Handle)

	onMasterChange(masterNode, role)
	assert.False(t, underlyingCallback.called)
	assert.Equal(t, underlyingCallback.lastcallRoleArg, Phantom) // Stays the same -- do not changes!
	assert.Equal(t, underlyingCallback.lastCallNodeInfoArg, NodeInfo{})

	onMasterChange(masterNode, role)
	assert.False(t, underlyingCallback.called)
	assert.Equal(t, underlyingCallback.lastcallRoleArg, Phantom)
	assert.Equal(t, underlyingCallback.lastCallNodeInfoArg, NodeInfo{})
}

func TestAppliedWhenCalledWithDifferentMaster(t *testing.T) {
	firstMasterNode := NodeInfo{}
	role := Slave

	underlyingCallback := callback{lastcallRoleArg: Phantom}
	onMasterChange := ApplyOnMasterChange(underlyingCallback.Handle)
	onMasterChange(firstMasterNode, role)

	newMasterNode := NodeInfo{IP: "127.0.0.1", SWIMPort: 1234, BusPort: 5431, HttpPort: 8080}
	onMasterChange(newMasterNode, Slave)
	assert.True(t, underlyingCallback.called)
	assert.Equal(t, underlyingCallback.lastcallRoleArg, Slave)
	assert.Equal(t, underlyingCallback.lastCallNodeInfoArg, newMasterNode)
}

type callback struct {
	called              bool
	lastCallNodeInfoArg NodeInfo
	lastcallRoleArg     Role
}

func (c *callback) Handle(nodeInfo NodeInfo, role Role) {
	c.called = true
	c.lastCallNodeInfoArg = nodeInfo
	c.lastcallRoleArg = role
}
