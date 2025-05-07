package domain

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
)

type MicroserviceKey struct {
	Name      string
	Namespace string
}

func (k MicroserviceKey) GetNamespace() msaddr.Namespace {
	return msaddr.Namespace{Namespace: k.Namespace}
}

func (c *Cluster) GetMicroserviceKey() MicroserviceKey {
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(c.Name)
	return MicroserviceKey{
		Name:      clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(c.Name),
		Namespace: namespace.GetNamespace(),
	}
}

func (m *MicroserviceVersion) GetMicroserviceKey() MicroserviceKey {
	return MicroserviceKey{
		Name:      m.Name,
		Namespace: m.Namespace,
	}
}
