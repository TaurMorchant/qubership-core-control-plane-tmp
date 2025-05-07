package events

import (
	"fmt"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

type ChangeEvent struct {
	NodeGroup string
	Changes   map[string][]memdb.Change
}

func (ce *ChangeEvent) ToString() string {
	str := "{ChangeEvent: nodeGroup='" + ce.NodeGroup + "', changes=[\n"
	for key, changes := range ce.Changes {
		str += key + ": ["
		for _, change := range changes {
			if change.Before == nil {
				str += "nil, "
			} else {
				str += fmt.Sprintf("%v, ", change.Before)
			}
			if change.After == nil {
				str += "nil, "
			} else {
				str += fmt.Sprintf("%v, ", change.After)
			}
		}
		str += "], \n"
	}
	return str + "]}"
}

// MultipleChangeEvent supports changes in multiple nodeGroups.
type MultipleChangeEvent struct {
	Changes map[string][]memdb.Change
}

type ReloadEvent struct {
	Changes map[string][]memdb.Change
}

func (ce *ChangeEvent) MarshalPrepare() error {
	for _, changes := range ce.Changes {
		if err := marshalPrepareForChanges(changes); err != nil {
			return err
		}
	}
	return nil
}

func (mce *MultipleChangeEvent) MarshalPrepare() error {
	for _, changes := range mce.Changes {
		if err := marshalPrepareForChanges(changes); err != nil {
			return err
		}
	}
	return nil
}

func (re *ReloadEvent) MarshalPrepare() error {
	for _, changes := range re.Changes {
		if err := marshalPrepareForChanges(changes); err != nil {
			return err
		}
	}
	return nil
}

func marshalPrepareForChanges(changes []memdb.Change) error {
	for _, change := range changes {
		if change.Before != nil {
			if err := change.Before.(domain.MarshalPreparer).MarshalPrepare(); err != nil {
				return err
			}
		}
		if change.After != nil {
			if err := change.After.(domain.MarshalPreparer).MarshalPrepare(); err != nil {
				return err
			}
		}
	}
	return nil
}

// PartialReloadEvent signals that all envoy cache entries for the specified nodeGroup:entityType pairs
// must be reloaded from in-memory storage.
type PartialReloadEvent struct {
	EnvoyVersions []*domain.EnvoyConfigVersion
}
