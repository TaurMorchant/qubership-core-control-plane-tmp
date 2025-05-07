package tls

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/configresources"
	"reflect"
)

type tlsDefResource struct {
	service *Service
}

func (h tlsDefResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "TlsDef",
	}
}

func (h tlsDefResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.TlsConfig{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			tlsConfig, ok := entity.(*dto.TlsConfig)
			if !ok {
				return false, "bad TlsDef payload"
			}
			if len(tlsConfig.Name) == 0 {
				return false, mandatory("spec.name")
			}
			if len(tlsConfig.TrustedForGateways) != 0 && (len(tlsConfig.TrustedForGateways) > 1 || cache.EgressGateway != tlsConfig.TrustedForGateways[0]) {
				return false, "global TLS supported only for " + cache.EgressGateway
			}
			return true, ""
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			tlsConfig := entity.(*dto.TlsConfig)
			err := h.service.SaveTlsConfig(ctx, tlsConfig)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.TlsConfig).Overridden
		},
	}
}

func mandatory(field string) string {
	return field + " is mandatory"
}
