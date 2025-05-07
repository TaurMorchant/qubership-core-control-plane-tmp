package health

import (
	"bytes"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/clustering"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/com9n"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var log = logging.GetLogger("health")

type HealthService struct {
	com9nCfg com9n.IConfigurator
}

func NewHealthService(com9nCfg com9n.IConfigurator) *HealthService {
	return &HealthService{
		com9nCfg: com9nCfg,
	}
}

//swagger:model Readiness
type Readiness struct {
	Status  ReadyStatus
	Role    clustering.Role
	Details string `json:",omitempty"`
}

//swagger:model Health
type Health struct {
	Status  HealthStatus
	Role    clustering.Role
	Details string   `json:",omitempty"`
	Errors  []string `json:",omitempty"`
}

type HealthStatus int
type ReadyStatus int

const (
	Up HealthStatus = iota
	Problem
)

const (
	Ready ReadyStatus = iota
	NotReady
)

func (s HealthStatus) String() string {
	switch s {
	case Up:
		return "UP"
	case Problem:
		return "PROBLEM"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

func (s ReadyStatus) String() string {
	switch s {
	case Ready:
		return "READY"
	case NotReady:
		return "NOT_READY"
	default:
		return fmt.Sprintf("%d", int(s))
	}
}

func (s HealthStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (s ReadyStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(s.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (s *HealthService) CheckReadiness() Readiness {
	nodeState := clustering.CurrentNodeState.GetRole()

	switch nodeState {
	case clustering.Master:
		if clustering.CurrentNodeState.IsMasterReady() || clustering.CurrentNodeState.IsMasterHasLoadedData() {
			return Readiness{Status: Ready, Role: nodeState}
		} else {
			return Readiness{Status: NotReady, Role: nodeState, Details: "Master is not ready to serve requests"}
		}
	case clustering.Slave:
		if s.com9nCfg.IsReceiverStarted() {
			return Readiness{Status: Ready, Role: nodeState}
		} else {
			return Readiness{Status: NotReady, Role: nodeState, Details: "Receiver data from master is not active"}
		}
	case clustering.Phantom:
		return Readiness{Status: Ready, Role: nodeState}
	default:
		return Readiness{Status: NotReady, Role: nodeState, Details: fmt.Sprintf("Unknown state of node: %v", nodeState)}
	}
}

func (s *HealthService) CheckLiveness() Health {
	fatalErrors := clustering.GetFatalErrorsExceptInitErrors()
	if len(fatalErrors) > 0 {
		strErrs := make([]string, 0, len(fatalErrors))
		for _, strErr := range fatalErrors {
			strErrs = append(strErrs, strErr.Error())
		}
		return Health{Status: Problem, Errors: strErrs, Role: clustering.CurrentNodeState.GetRole()}
	}
	return Health{Status: Up, Role: clustering.CurrentNodeState.GetRole()}
}
