package dao

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"

type IDGenerator interface {
	Generate(uniqEntity domain.Unique) error
}
