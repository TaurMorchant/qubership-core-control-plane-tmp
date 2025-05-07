package statefulsession

import (
	"context"
	"errors"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"reflect"
	"strings"
)

type RequestValidator interface {
	ValidateStatefulSession(req dto.StatefulSession) (bool, string)
}

type statefulSessionResource struct {
	validator RequestValidator
	service   Service
}

func (r statefulSessionResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "StatefulSession",
	}
}

func (r statefulSessionResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.StatefulSession{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			config, ok := entity.(*dto.StatefulSession)
			if !ok {
				return false, "Bad StatefulSession payload"
			}
			if err := enrichSpecWithNamespace(config, md); err != nil {
				return false, err.Error()
			}
			return r.validator.ValidateStatefulSession(*config)
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			config := entity.(*dto.StatefulSession)
			if err := enrichSpecWithNamespace(config, md); err != nil {
				return nil, err
			}
			err := r.service.ApplyStatefulSession(ctx, config)
			return "StatefulSession configuration applied successfully", err
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.StatefulSession).Overridden
		},
	}
}

func enrichSpecWithNamespace(spec *dto.StatefulSession, md cfgres.Metadata) error {
	if spec.Namespace == "" {
		if namespaceRaw, ok := md["namespace"]; ok {
			if namespace, ok := namespaceRaw.(string); ok {
				spec.Namespace = strings.TrimSpace(namespace)
			} else {
				return errors.New("statefulsession: field 'namespace' must have string type")
			}
		}
	}
	return nil
}
