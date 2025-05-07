package configresources

import (
	"context"
	"reflect"
)

type ResourceService[R any] interface {
	ResourceApplier[R]
	GetAll(ctx context.Context) ([]R, error)
}

type ResourceApplier[R any] interface {
	Validate(ctx context.Context, res R) (bool, string)

	IsOverriddenByCR(ctx context.Context, res R) bool

	Apply(ctx context.Context, res R) (any, error)

	GetConfigRes() ConfigRes[R]
}

type ConfigRes[R any] struct {
	Key     ResourceKey
	Applier ResourceApplier[R]
}

func (r ConfigRes[R]) GetKey() ResourceKey {
	return r.Key
}

func (r ConfigRes[R]) GetDefinition() ResourceDef {
	var res R
	return ResourceDef{
		Type: reflect.TypeOf(res),
		Validate: func(ctx context.Context, metadata Metadata, entity any) (bool, string) {
			payload, ok := entity.(*R)
			if !ok {
				return false, "bad payload"
			}
			return r.Applier.Validate(ctx, *payload)
		},
		Handler: func(ctx context.Context, md Metadata, entity any) (any, error) {
			payload := entity.(*R)
			return r.Applier.Apply(ctx, *payload)
		},
		IsOverriddenByCR: func(ctx context.Context, metadata Metadata, entity interface{}) bool {
			payload := entity.(*R)
			return r.Applier.IsOverriddenByCR(ctx, *payload)
		},
	}
}
