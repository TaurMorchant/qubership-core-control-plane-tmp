package websocket

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
)

type Message struct {
	Changes []Change                    `json:"changes,omitempty"`
	State   []*domain.DeploymentVersion `json:"state,omitempty"`
}

type Change struct {
	New *domain.DeploymentVersion `json:"new"`
	Old *domain.DeploymentVersion `json:"old"`
}

type MessageType int

const (
	Error MessageType = iota
	Versions
)

var toString = map[MessageType]string{
	Error:    "Error",
	Versions: "Versions",
}

var toID = map[string]MessageType{
	"Error":    Error,
	"Versions": Versions,
}

func (t MessageType) String() string {
	return toString[t]
}

func (t MessageType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(t.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *MessageType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	*t = toID[j]
	return nil
}

func NewMessage(state []*domain.DeploymentVersion, changes []Change) (*Message, error) {
	return &Message{
		State:   state,
		Changes: changes,
	}, nil
}

func NewChange(change memdb.Change) Change {
	var newCh *domain.DeploymentVersion
	var oldCh *domain.DeploymentVersion
	if change.After != nil {
		// Struct links sent from internal bus, struct values sent from external bus
		if val, ok := change.After.(domain.DeploymentVersion); ok {
			newCh = &val
		}
		if val, ok := change.After.(*domain.DeploymentVersion); ok {
			newCh = val
		}
	}
	if change.Before != nil {
		if val, ok := change.Before.(domain.DeploymentVersion); ok {
			oldCh = &val
		}
		if val, ok := change.Before.(*domain.DeploymentVersion); ok {
			oldCh = val
		}
	}
	return Change{
		New: newCh,
		Old: oldCh,
	}
}
