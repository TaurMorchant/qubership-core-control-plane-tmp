package httpFilter

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"reflect"
)

type httpFiltersResourceAdd struct {
	service *Service
}

func (h httpFiltersResourceAdd) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "HttpFilters",
	}
}

func (h httpFiltersResourceAdd) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.HttpFiltersConfigRequestV3{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			return h.service.ValidateApply(ctx, entity.(*dto.HttpFiltersConfigRequestV3))
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			return nil, h.service.Apply(ctx, entity.(*dto.HttpFiltersConfigRequestV3))
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.HttpFiltersConfigRequestV3).Overridden
		},
	}
}
