package events

import (
	"github.com/hashicorp/go-memdb"
)

func NewChangeEventByNodeGroup(nodeGroup string, changes []memdb.Change) *ChangeEvent {
	changesByType := changesToMap(changes)
	return &ChangeEvent{
		NodeGroup: nodeGroup,
		Changes:   changesByType,
	}
}

func NewChangeEvent(changes []memdb.Change) *ChangeEvent {
	changesByType := changesToMap(changes)
	return &ChangeEvent{
		Changes: changesByType,
	}
}

func NewMultipleChangeEvent(changes []memdb.Change) *MultipleChangeEvent {
	changesByType := changesToMap(changes)
	return &MultipleChangeEvent{Changes: changesByType}
}

func NewReloadEvent(changes []memdb.Change) *ReloadEvent {
	changesByType := changesToMap(changes)
	return &ReloadEvent{
		Changes: changesByType,
	}
}

func changesToMap(changes []memdb.Change) map[string][]memdb.Change {
	changesByType := make(map[string][]memdb.Change)
	for _, change := range changes {
		var typedChanges []memdb.Change
		if value, ok := changesByType[change.Table]; ok {
			typedChanges = value
		} else {
			typedChanges = make([]memdb.Change, 0)
		}
		typedChanges = append(typedChanges, change)
		changesByType[change.Table] = typedChanges
	}
	return changesByType
}
