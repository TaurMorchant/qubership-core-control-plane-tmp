package cluster

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"reflect"
)

type clusterResource struct {
	service *Service
}

func (h clusterResource) GetKey() cfgres.ResourceKey {
	return cfgres.ResourceKey{
		APIVersion: "nc.core.mesh/v3",
		Kind:       "Cluster",
	}
}

func (h clusterResource) GetDefinition() cfgres.ResourceDef {
	return cfgres.ResourceDef{
		Type: reflect.TypeOf(dto.ClusterConfigRequestV3{}),
		Validate: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (bool, string) {
			clusterConfigRequestV3req, ok := entity.(*dto.ClusterConfigRequestV3)
			if !ok {
				return false, "bad ClusterConfigRequestV3 payload"
			}
			if len(clusterConfigRequestV3req.Name) == 0 {
				return false, mandatory("spec.name")
			}
			if len(clusterConfigRequestV3req.Endpoints) == 0 {
				return false, mandatory("spec.endpoints")
			}
			if len(clusterConfigRequestV3req.Gateways) == 0 {
				return false, mandatory("spec.gateways")
			}
			return true, ""
		},
		Handler: func(ctx context.Context, md cfgres.Metadata, entity interface{}) (interface{}, error) {
			clusterConfigRequestV3req := entity.(*dto.ClusterConfigRequestV3)
			for _, nodeGroupId := range clusterConfigRequestV3req.Gateways {
				err := h.service.AddCluster(ctx, nodeGroupId, clusterConfigRequestV3req)
				if err != nil {
					return nil, err
				}
			}

			return nil, nil
		},
		IsOverriddenByCR: func(ctx context.Context, metadata cfgres.Metadata, entity interface{}) bool {
			return entity.(*dto.ClusterConfigRequestV3).Overridden
		},
	}
}

func mandatory(field string) string {
	return field + " is mandatory"
}
