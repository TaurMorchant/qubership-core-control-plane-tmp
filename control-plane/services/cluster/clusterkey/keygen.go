package clusterkey

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"strings"
)

type Generator interface {
	GenerateKey(clusterName string, msAddr *msaddr.MicroserviceAddress) string
	ExtractFamilyName(clusterKey string) string
	ExtractNamespace(clusterKey string) msaddr.Namespace
	// BuildKeyPrefix builds value to be used in search by cluster key prefix operations.
	BuildKeyPrefix(clusterName string, namespace msaddr.Namespace) string
}

type defaultGenerator struct {
	keyPattern          string
	searchPrefixPattern string
}

var DefaultClusterKeyGenerator = &defaultGenerator{keyPattern: "%s||%s||%v", searchPrefixPattern: "%s||%s||"}

func (keyGen *defaultGenerator) GenerateKey(clusterName string, msAddr *msaddr.MicroserviceAddress) string {
	clusterName = strings.TrimSpace(clusterName)
	if len(clusterName) == 0 {
		return fmt.Sprintf(keyGen.keyPattern,
			msAddr.GetMicroserviceName(),
			keyGen.getNamespacedClusterHost(msAddr.Host, msAddr.Namespace),
			msAddr.GetPort())
	} else {
		return fmt.Sprintf(keyGen.keyPattern,
			clusterName,
			keyGen.getNamespacedClusterHost(clusterName, msAddr.Namespace),
			msAddr.GetPort())
	}
}

func (keyGen *defaultGenerator) ExtractFamilyName(clusterKey string) string {
	if idx := strings.Index(clusterKey, "||"); idx != -1 {
		return clusterKey[:idx]
	}
	return clusterKey
}

func (keyGen *defaultGenerator) ExtractNamespace(clusterKey string) msaddr.Namespace {
	namespacedName := keyGen.extractNamespacedName(clusterKey)
	if dotIdx := strings.Index(namespacedName, "."); dotIdx != -1 {
		namespace := namespacedName[dotIdx+1:]
		return msaddr.Namespace{Namespace: namespace}
	}
	return msaddr.Namespace{}
}

func (keyGen *defaultGenerator) extractNamespacedName(clusterKey string) string {
	keyParts := strings.Split(clusterKey, "||")
	if len(keyParts) > 1 {
		return keyParts[1]
	}
	return clusterKey
}

func (keyGen *defaultGenerator) BuildKeyPrefix(clusterName string, namespace msaddr.Namespace) string {
	return fmt.Sprintf(keyGen.searchPrefixPattern, clusterName, keyGen.getNamespacedClusterHost(clusterName, namespace))
}

func (keyGen *defaultGenerator) getNamespacedClusterHost(clusterName string, namespace msaddr.Namespace) string {
	if namespace.IsCurrentNamespace() {
		return clusterName
	}
	return clusterName + "." + namespace.Namespace
}
