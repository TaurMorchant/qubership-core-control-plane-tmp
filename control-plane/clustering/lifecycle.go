package clustering

import (
	"github.com/pkg/errors"
	"sync"
)

//go:generate mockgen -source=lifecycle.go -destination=../test/mock/clustering/stub_lifecycle.go -package=mock_clustering
type MasterNodeInitializer interface {
	InitMaster() error
}

var fatalErrors fatalErrorStack

type fatalErrorStack struct {
	fatalErrors []error
	lock        sync.Mutex
}

func AppendFatal(err error) {
	fatalErrors.lock.Lock()
	defer fatalErrors.lock.Unlock()
	log.Errorf("Appending fatal error: %v", err)
	if isInitMasterError(err) && hasInitMasterErrorInNodeInArray(fatalErrors.fatalErrors) {
		return
	}
	fatalErrors.fatalErrors = append(fatalErrors.fatalErrors, err)
}

func GetFatalErrors() []error {
	fatalErrors.lock.Lock()
	defer fatalErrors.lock.Unlock()
	return fatalErrors.fatalErrors
}

func GetFatalErrorsExceptInitErrors() []error {
	fatalErrors.lock.Lock()
	defer fatalErrors.lock.Unlock()
	var result []error
	for _, err := range fatalErrors.fatalErrors {
		if isInitError(err) {
			continue
		}
		result = append(result, err)
	}
	return result
}

func CleanFatalErrors() {
	fatalErrors.lock.Lock()
	defer fatalErrors.lock.Unlock()
	fatalErrors.fatalErrors = make([]error, 0)
	log.Debugf("Fatal error collection cleaned")
}

// be carefull when compare the thisNode with another MasterMetadata structure
// thisNode is initialized once on start and it is necessary to exclude id from comparing
// use thisNode.EqualsByNameAndNodeInfo(anotherNode) for this
type LifeCycleManager struct {
	currentMaster *MasterMetadata
	thisNode      MasterMetadata
	masterInit    MasterNodeInitializer
	callbacks     []func(NodeInfo, Role) error
}

func NewLifeCycleManager(nodeName, namespace string, nodeInfo NodeInfo, initializer MasterNodeInitializer) *LifeCycleManager {
	return &LifeCycleManager{
		masterInit:    initializer,
		currentMaster: &MasterMetadata{},
		thisNode: MasterMetadata{
			Name:      nodeName,
			NodeInfo:  nodeInfo,
			Namespace: namespace,
		},
	}
}

func (m *LifeCycleManager) defineNodeRole(newMaster *MasterMetadata) {
	role := Initial
	if newMaster == nil {
		log.DebugC(ctx, "newMaster is nil")
		if CurrentNodeState.IsPhantom() {
			log.DebugC(ctx, "current node is Phantom")
			return
		}
		role = Phantom
		log.Infof("Role of node '%s' defined as %s", m.thisNode.Name, role)
		m.handleNodeRoleChange(newMaster, role)
	} else if !m.currentMaster.Equals(newMaster) {
		if m.thisNode.EqualsByNameAndNodeInfo(newMaster) {
			role = Master
			log.Infof("Role of node '%s' defined as %s.", m.thisNode.Name, role)
		} else {
			role = Slave
			log.Infof("Role of node '%s' defined as %s. Master node is '%s'", m.thisNode.Name, role, newMaster.Name)
		}
		m.handleNodeRoleChange(newMaster, role)
	} else if HasInitErrorInNode() {
		m.handleNodeRoleChange(newMaster, CurrentNodeState.GetRole())
	}
	m.currentMaster = newMaster
}

func HasInitErrorInNode() bool {
	return hasInitErrorInNodeInArray(GetFatalErrors())
}

func hasInitMasterErrorInNodeInArray(errs []error) bool {
	for _, err := range errs {
		if isInitMasterError(err) {
			return true
		}
	}
	return false
}

func hasInitErrorInNodeInArray(errs []error) bool {
	for _, err := range errs {
		if isInitError(err) {
			return true
		}
	}
	return false
}

func isInitMasterError(err error) bool {
	_, ok := err.(initMasterError)
	return ok
}

func isInitError(err error) bool {
	_, ok := err.(initError)
	return ok
}

type initError interface {
	initError() error
}

type initMasterError struct {
	error
}

type initCallbacksError struct {
	error
}

func (e initCallbacksError) Error() string {
	return e.error.Error()
}

func (e initCallbacksError) initError() error {
	return e.error
}

func (e initMasterError) Error() string {
	return e.error.Error()
}

func (e initMasterError) initError() error {
	return e.error
}

func NewInitMasterError(err error) error {
	return initMasterError{errors.Wrapf(err, "can't initialize master")}
}

func NewCallbacksError(err error) error {
	return initCallbacksError{errors.Wrapf(err, "can't execute callback in node")}
}

func (m *LifeCycleManager) handleNodeRoleChange(record *MasterMetadata, role Role) {
	var nodeInfo NodeInfo
	if record != nil {
		nodeInfo = record.NodeInfo
	}
	CurrentNodeState.ChangeNodeState(nodeInfo, role)
	if role == Master {
		if err := m.masterInit.InitMaster(); err != nil {
			AppendFatal(NewInitMasterError(err))
			return
		}
	}
	for _, callback := range m.callbacks {
		if err := callback(nodeInfo, role); err != nil {
			AppendFatal(NewCallbacksError(err))
			return
		}
	}
	CleanFatalErrors()
	if role == Master {
		CurrentNodeState.SetMasterReady()
	}
	log.DebugC(ctx, "Handling node role %v finished", role)
}

func (m *LifeCycleManager) GetThisNodeMetadata() MasterMetadata {
	return m.thisNode
}

func (m *LifeCycleManager) AddOnRoleChanged(callback func(NodeInfo, Role) error) {
	m.callbacks = append(m.callbacks, callback)
}
