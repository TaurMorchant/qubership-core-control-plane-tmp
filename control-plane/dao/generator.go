package dao

import "github.com/netcracker/qubership-core-control-plane/domain"

type IDGenerator interface {
	Generate(uniqEntity domain.Unique) error
}
