package dao

import (
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/cluster/clusterkey"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"strings"
)

func (d *InMemRepo) FindAllStatefulSessionConfigs() ([]*domain.StatefulSession, error) {
	return FindAll[domain.StatefulSession](d, domain.StatefulSessionTable)
}

func (d *InMemRepo) FindStatefulSessionConfigById(id int32) (*domain.StatefulSession, error) {
	return FindById[domain.StatefulSession](d, domain.StatefulSessionTable, id)
}

func (d *InMemRepo) SaveStatefulSessionConfig(cookie *domain.StatefulSession) error {
	namespace := msaddr.Namespace{Namespace: cookie.Namespace}
	if namespace.IsCurrentNamespace() {
		cookie.Namespace = "default" // always use 'default' instead of empty so index works properly
	}
	return d.SaveUnique(domain.StatefulSessionTable, cookie)
}

func (d *InMemRepo) FindStatefulSessionConfigsByCookieName(cookieName string) ([]*domain.StatefulSession, error) {
	return FindByIndex[domain.StatefulSession](d, domain.StatefulSessionTable, "cookieName", cookieName)
}

func (d *InMemRepo) FindStatefulSessionConfigsByClusterAndVersion(clusterName string, namespace msaddr.Namespace, version *domain.DeploymentVersion) ([]*domain.StatefulSession, error) {
	var resolvedNamespace string
	if namespace.IsCurrentNamespace() {
		resolvedNamespace = "default" // always use 'default' instead of empty so index works properly
	} else {
		resolvedNamespace = namespace.Namespace
	}
	return FindByIndex[domain.StatefulSession](d, domain.StatefulSessionTable, "clusterNamespaceAndVersion", clusterName, resolvedNamespace, version.Version)
}

func (d *InMemRepo) FindStatefulSessionConfigsByCluster(cluster *domain.Cluster) ([]*domain.StatefulSession, error) {
	familyName := clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(cluster.Name)
	namespace := clusterkey.DefaultClusterKeyGenerator.ExtractNamespace(cluster.Name)
	return d.FindStatefulSessionConfigsByClusterName(familyName, namespace)
}

func (d *InMemRepo) FindStatefulSessionConfigsByClusterName(clusterName string, namespace msaddr.Namespace) ([]*domain.StatefulSession, error) {
	if strings.Contains(clusterName, "||") {
		clusterName = clusterkey.DefaultClusterKeyGenerator.ExtractFamilyName(clusterName)
	}
	return d.findStatefulSessionsByCondition(func(cookie *domain.StatefulSession) bool {
		return clusterName == cookie.ClusterName &&
			namespace.Equals(msaddr.Namespace{Namespace: cookie.Namespace})
	})
}

func (d *InMemRepo) findStatefulSessionsByCondition(condition func(cookie *domain.StatefulSession) bool) ([]*domain.StatefulSession, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	allCookies, err := d.FindAllStatefulSessionConfigs()
	if err != nil {
		return nil, errors.WrapPrefix(err, "error loading all StatefulSessionConfigs in memstorage wrapper", 1)
	}
	result := make([]*domain.StatefulSession, 0, 10)
	for _, cookie := range allCookies {
		if condition(cookie) {
			result = append(result, cookie)
		}
	}
	return result, nil
}

func (d *InMemRepo) DeleteStatefulSessionConfig(id int32) error {
	return d.DeleteById(domain.StatefulSessionTable, id)
}
