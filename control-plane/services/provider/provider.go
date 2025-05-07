package provider

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
)

type Env struct {
	tlsService TlsService
}

var env *Env

type TlsService interface {
	GetGlobalTlsConfigs(cluster *domain.Cluster, affectedNodeGroups ...string) ([]*domain.TlsConfig, error)
}

func Init(ts TlsService) Env {
	env = &Env{tlsService: ts}
	return *env
}

func GetTlsService() TlsService {
	if env == nil {
		panic("call Init before use.")
	}
	return env.tlsService
}
