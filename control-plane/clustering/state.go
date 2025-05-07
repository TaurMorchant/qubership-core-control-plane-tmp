package clustering

import (
	"sync"
	"time"
)

// TODO don't use this object directly, better way define interface
var CurrentNodeState NodeState

func init() {
	CurrentNodeState = NodeState{role: Initial, masterInfo: NodeInfo{}, masterReady: false}
}

type NodeState struct {
	role                Role
	masterInfo          NodeInfo
	masterReady         bool
	masterHasLoadedData bool
	lock                sync.RWMutex
}

func (s *NodeState) ChangeNodeState(record NodeInfo, state Role) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.role = state
	s.masterInfo = record
	s.masterReady = false
}

func (s *NodeState) GetRole() Role {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.role
}

func (s *NodeState) GetHttpAddress() string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.masterInfo.GetHttpAddress()
}

func (s *NodeState) IsSlave() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.role == Slave {
		return true
	}
	return false
}

func (s *NodeState) IsMaster() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.role == Master {
		return true
	}
	return false
}

func (s *NodeState) IsMasterReady() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.masterReady
}

func (s *NodeState) IsMasterHasLoadedData() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.masterHasLoadedData
}

func (s *NodeState) WaitMasterReady() {
	for !CurrentNodeState.IsMasterReady() {
		log.Infof("Incoming request is waiting for master ready")
		time.Sleep(3 * time.Second)
	}
}

func (s *NodeState) SetMasterReady() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.masterReady = true
	s.masterHasLoadedData = true
	log.Infof("Master is ready to server requests.")
}

func (s *NodeState) IsPhantom() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.role == Phantom
}
