package cluster

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache/builder/common"
	"os"
	"strings"

	tlsV3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
)

type EgressClusterBuilder struct {
	BaseClusterBuilder
	ecdhCurves []string
}

func NewEgressClusterBuilder(dao dao.Repository, routeProperties *common.RouteProperties) *EgressClusterBuilder {
	var curves []string
	if value, exists := os.LookupEnv("ECDH_CURVES"); exists {
		curves = strings.Split(value, ",")
	}
	builder := &EgressClusterBuilder{ecdhCurves: curves}
	baseBuilder := BaseClusterBuilder{
		dao:                      dao,
		routeProperties:          routeProperties,
		enrichUpstreamTlsContext: builder.enrichUpstreamTlsContext,
	}
	builder.BaseClusterBuilder = baseBuilder
	return builder
}

func (builder *EgressClusterBuilder) enrichUpstreamTlsContext(upstreamTlsContext *tlsV3.UpstreamTlsContext) {
	upstreamTlsContext.CommonTlsContext.TlsParams = &tlsV3.TlsParameters{
		TlsMaximumProtocolVersion: tlsV3.TlsParameters_TLSv1_3,
	}
	if len(builder.ecdhCurves) > 0 {
		upstreamTlsContext.CommonTlsContext.TlsParams.EcdhCurves = builder.ecdhCurves
	}
}
