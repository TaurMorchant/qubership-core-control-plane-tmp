package httpFilter

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/services/httpFilter/extAuthz"
	"reflect"
)

type httpFiltersDropResourceDrop struct {
	service         *Service
	extAuthzService extAuthz.Service
}

func (h httpFiltersDropResourceDrop) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "HttpFiltersDrop",
	}
}

func (h httpFiltersDropResourceDrop) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.HttpFiltersDropConfigRequestV3{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			return h.service.ValidateDelete(ctx, entity.(*dto.HttpFiltersDropConfigRequestV3))
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			return nil, h.service.Delete(ctx, entity.(*dto.HttpFiltersDropConfigRequestV3))
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.HttpFiltersDropConfigRequestV3).Overridden
		},
	}
}

func asSliceByName(filters []map[string]interface{}) []string {
	result := make([]string, len(filters))
	for i, f := range filters {
		result[i] = f["name"].(string)
	}
	return result
}
