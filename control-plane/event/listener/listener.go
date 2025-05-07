package listener

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
)

type ChangeEventListener struct {
	updateManager *cache.UpdateManager
}

func NewChangeEventListener(updateManager *cache.UpdateManager) *ChangeEventListener {
	return &ChangeEventListener{updateManager: updateManager}
}

func (l *ChangeEventListener) HandleEvent(data interface{}) {
	event := data.(*events.ChangeEvent)
	l.updateManager.HandleChangeEvent(event)
}

// TODO Get rid of ReloadEvent. Make Generic event with switch
type ReloadEventListener struct {
	updateManager *cache.UpdateManager
}

func NewReloadEventListener(updateManager *cache.UpdateManager) *ReloadEventListener {
	return &ReloadEventListener{updateManager: updateManager}
}

func (l *ReloadEventListener) HandleEvent(data interface{}) {
	l.updateManager.HandleReloadEvent()
}

// TODO Get rid of MultipleChangeEvent. Make Generic event with switch
// MultipleChangeEventListener listens for MultipleChangeEvent which is an event that supports changes in several nodeGroups.
type MultipleChangeEventListener struct {
	updateManager *cache.UpdateManager
}

func NewMultipleChangeEventListener(updateManager *cache.UpdateManager) *MultipleChangeEventListener {
	return &MultipleChangeEventListener{updateManager: updateManager}
}

func (l *MultipleChangeEventListener) HandleEvent(data interface{}) {
	event := data.(*events.MultipleChangeEvent)
	l.updateManager.HandleMultipleChangeEvent(event)
}

type PartialReloadEventListener struct {
	updateManager *cache.UpdateManager
}

func NewPartialReloadEventListener(updateManager *cache.UpdateManager) *PartialReloadEventListener {
	return &PartialReloadEventListener{updateManager: updateManager}
}

func (l *PartialReloadEventListener) HandleEvent(data interface{}) {
	event := data.(*events.PartialReloadEvent)
	l.updateManager.HandlePartialReloadEvent(event)
}
