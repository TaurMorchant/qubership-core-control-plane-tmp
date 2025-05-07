package listener

import (
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/builder/common"
)

const facadeDefaultHost = "facade"
const facadeStatPrefix = "facade_http"

type FacadeListenerBuilder struct {
	BaseListenerBuilder
}

func NewFacadeListenerBuilder(tracingProperties *common.TracingProperties) *FacadeListenerBuilder {
	return &FacadeListenerBuilder{BaseListenerBuilder{
		defaultHost:       facadeDefaultHost,
		statPrefix:        facadeStatPrefix,
		tracingProperties: tracingProperties,
		enrichConnManager: func(connManager *hcm.HttpConnectionManager, namespaceMapping string) error {
			return nil
		},
		enrichListener: func(v *listener.Listener) error {
			return nil
		},
	}}
}
