package tls

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_tlsDefResource_GetDefinition_only_for_egress(t *testing.T) {
	resource := tlsDefResource{}
	valid, msg := resource.GetDefinition().Validate(context.Background(), nil, &dto.TlsConfig{
		Name:               "test-tls",
		TrustedForGateways: []string{"not-egress-gateway"},
	})
	assert.False(t, valid)
	assert.Equal(t, "global TLS supported only for "+cache.EgressGateway, msg)
}

func Test_tlsDefResource_GetDefinition_egress(t *testing.T) {
	resource := tlsDefResource{}
	valid, msg := resource.GetDefinition().Validate(context.Background(), nil, &dto.TlsConfig{
		Name:               "test-tls",
		TrustedForGateways: []string{"egress-gateway"},
	})
	assert.True(t, valid)
	assert.Empty(t, msg)
}

func TestService_SingleOverriddenWithTrueValueForRoutingConfigRequestV3(t *testing.T) {
	resource := tlsDefResource{}
	isOverridden := resource.GetDefinition().IsOverriddenByCR(nil, nil, &dto.TlsConfig{
		Name:               "test-tls",
		TrustedForGateways: []string{"egress-gateway"},
		Overridden:         true,
	})
	assert.True(t, isOverridden)
}
