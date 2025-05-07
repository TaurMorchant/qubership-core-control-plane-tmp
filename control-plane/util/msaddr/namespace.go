package msaddr

import (
	"os"
	"strings"
)

type Namespace struct {
	Namespace string
}

func NewNamespace(namespace string) *Namespace {
	return &Namespace{Namespace: namespace}
}

const (
	LocalDevNamespacePostfix = ".nip.io"
	DefaultNamespace         = "default"
	CloudNamespace           = "CLOUD_NAMESPACE"
	LocalNamespace           = "local"
	UnknownNamespace         = "unknown"
)

func CurrentNamespace() Namespace {
	return Namespace{Namespace: CurrentNamespaceAsString()}
}

func CurrentNamespaceAsString() string {
	return gatewayNamespace()
}

func (this *Namespace) IsCurrentNamespace() bool {
	return len(strings.TrimSpace(this.Namespace)) == 0 || this.Namespace == DefaultNamespace ||
		this.Namespace == UnknownNamespace ||
		this.Namespace == gatewayNamespace()
}

func (this *Namespace) Equals(another Namespace) bool {
	return this.Namespace == another.Namespace || (this.IsCurrentNamespace() && another.IsCurrentNamespace())
}

func gatewayNamespace() string {
	ns := os.Getenv(CloudNamespace)
	if len(ns) > 0 {
		return ns
	} else {
		return LocalNamespace
	}
}

func (this *Namespace) IsLocalDevNamespace() bool {
	return this.Namespace != "" && strings.HasSuffix(this.Namespace, LocalDevNamespacePostfix)
}

// GetNamespace function returns namespace name as string, never returning empty string or any default value.
func (this *Namespace) GetNamespace() string {
	if this.IsCurrentNamespace() {
		return gatewayNamespace()
	}
	return this.Namespace
}
