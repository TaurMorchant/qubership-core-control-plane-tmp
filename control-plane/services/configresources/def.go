package configresources

import (
	"context"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/errorcodes"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"reflect"
)

var (
	registry        = map[ResourceKey]ResourceDef{}
	log             = logging.GetLogger("config-resources")
	ErrIsOverridden = errors.New("Error for overriding config")
)

type ResourceDef struct {
	Type             reflect.Type
	Validate         func(ctx context.Context, metadata Metadata, entity interface{}) (bool, string)
	Handler          func(ctx context.Context, metadata Metadata, entity interface{}) (interface{}, error)
	IsOverriddenByCR func(ctx context.Context, metadata Metadata, entity interface{}) bool
}

func RegisterResource(resource Resource) {
	registry[resource.GetKey()] = resource.GetDefinition()
}

func RegisterResources(resources []Resource) {
	for _, resource := range resources {
		RegisterResource(resource)
	}
}

func HandleConfigResource(ctx context.Context, config ConfigResource) (interface{}, *errorcodes.CpErrCodeError) {
	if resDef, ok := registry[config.Key()]; ok {
		if config.Metadata == nil {
			config.Metadata = make(map[string]interface{})
		}
		config.Metadata["nodeGroup"] = config.NodeGroup
		entity := reflect.New(resDef.Type).Interface()
		cfg := &mapstructure.DecoderConfig{
			Result:     &entity,
			DecodeHook: domain.ToNullsHookFunc,
		}
		decoder, err := mapstructure.NewDecoder(cfg)
		if err != nil {
			log.ErrorC(ctx, "Creating new decoder for spec section caused error: %v", err)
			return nil, errorcodes.NewCpError(errorcodes.ValidationRequestError, err.Error(), err)
		}
		err = decoder.Decode(config.Spec)
		if err != nil {
			log.ErrorC(ctx, "Decoding spec section caused error: %v", err)
			return nil, errorcodes.NewCpError(errorcodes.ValidationRequestError, err.Error(), err)
		}
		if ok := resDef.IsOverriddenByCR(ctx, config.Metadata, entity); ok {
			log.InfoC(ctx, "Configuration resource was not applied because field overridden has true value")
			return nil, errorcodes.NewCpError(errorcodes.ValidationRequestError, ErrIsOverridden.Error(), ErrIsOverridden)
		}
		if ok, msg := resDef.Validate(ctx, config.Metadata, entity); !ok {
			log.WarnC(ctx, "Configuration resource is invalid: %v", msg)
			return nil, errorcodes.NewCpError(errorcodes.ValidationRequestError, msg, nil)
		}
		result, err := resDef.Handler(ctx, config.Metadata, entity)
		if err != nil {
			log.ErrorC(ctx, "Handling function for resource %v threw error %v", config.Key(), err)
			return nil, errorcodes.NewCpError(errorcodes.MultiCauseApplyConfigError, err.Error(), err)
		}
		return result, nil
	} else {
		log.WarnC(ctx, "There is no applicable implementation for resource %v", config.Key())
		return nil, errorcodes.NewCpError(errorcodes.ValidationRequestError, fmt.Sprintf("can't find resource definition for apiVersion: %s and kind: %s", config.APIVersion, config.Kind), nil)
	}
}

type Resource interface {
	GetKey() ResourceKey
	GetDefinition() ResourceDef
}

type ResourceProto struct {
	GetKeyFunc func() ResourceKey
	GetDefFunc func() ResourceDef
}

func (r ResourceProto) GetKey() ResourceKey {
	return r.GetKeyFunc()
}

func (r ResourceProto) GetDefinition() ResourceDef {
	return r.GetDefFunc()
}

type ConfigResource struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	NodeGroup  string      `json:"nodeGroup"`
	Metadata   Metadata    `json:"metadata"`
	Spec       interface{} `json:"spec"`
}

type Metadata map[string]interface{}

type ResourceKey struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
}

func (r ConfigResource) Key() ResourceKey {
	return ResourceKey{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
	}
}
