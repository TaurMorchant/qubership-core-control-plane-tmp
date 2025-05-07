package configresources

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenericConfig(t *testing.T) {
	srv := &TestApplier[TestRes]{Applied: make([]TestRes, 0, 1)}

	cfgRes := srv.GetConfigRes()

	RegisterResource(cfgRes)

	assert.Equal(t, "TestRes", cfgRes.GetKey().Kind)
	assert.Equal(t, "nc.core.mesh/v3", cfgRes.GetKey().APIVersion)

	var configResources []ConfigResource
	err := json.Unmarshal([]byte(`[
{
  "kind": "TestRes",
  "apiVersion": "nc.core.mesh/v3",
  "spec": {
    "name": "res1",
    "valid": true
  }
}, 
{
  "kind": "TestRes",
  "apiVersion": "nc.core.mesh/v3",
  "spec": {
    "name": "res2",
    "valid": false
  }
}]`), &configResources)
	assert.Nil(t, err)

	ctx := context.Background()
	for _, resource := range configResources {
		res, err := HandleConfigResource(ctx, resource)
		log.InfoC(ctx, "ConfigResource apply result: %+v, err: %v", res, err)
	}
	assert.Equal(t, 1, len(srv.Applied))
	assert.Equal(t, "res1", srv.Applied[0].Name)
}

type TestRes struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
}

type TestApplier[R TestRes] struct {
	Applied []TestRes
}

func (t *TestApplier[R]) Validate(ctx context.Context, res TestRes) (bool, string) {
	if !res.Valid {
		return false, "test err"
	}
	return true, ""
}

func (t *TestApplier[R]) IsOverriddenByCR(ctx context.Context, res TestRes) bool {
	return false
}

func (t *TestApplier[R]) Apply(ctx context.Context, res TestRes) (any, error) {
	t.Applied = append(t.Applied, res)
	return "OK", nil
}

func (t *TestApplier[R]) GetConfigRes() ConfigRes[TestRes] {
	return ConfigRes[TestRes]{
		Key: ResourceKey{
			APIVersion: "nc.core.mesh/v3",
			Kind:       "TestRes",
		},
		Applier: t,
	}
}
