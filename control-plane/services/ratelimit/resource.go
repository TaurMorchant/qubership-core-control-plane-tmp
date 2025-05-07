package ratelimit

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"reflect"
)

type rateLimitResource struct {
	service *Service
}

func (h rateLimitResource) Validate(rateLimit dto.RateLimit) (bool, string) {
	if len(rateLimit.Name) == 0 {
		return false, mandatory("spec.name")
	}
	return true, ""
}

func (h rateLimitResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "RateLimit",
	}
}

func (h rateLimitResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.RateLimit{}),
		Validate: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) (bool, string) {
			rateLimit, ok := entity.(*dto.RateLimit)
			if !ok {
				return false, "bad RateLimit payload"
			}
			return h.Validate(*rateLimit)
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			rateLimit := entity.(*dto.RateLimit)
			err := h.service.SaveRateLimit(ctx, rateLimit)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.RateLimit).Overridden
		},
	}
}

func mandatory(field string) string {
	return field + " is mandatory"
}
