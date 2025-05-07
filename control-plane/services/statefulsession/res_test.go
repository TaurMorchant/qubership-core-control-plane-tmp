package statefulsession

import (
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetKey(t *testing.T) {
	srv := NewService(nil, nil, nil)
	res := srv.GetStatefulSessionResource()

	resource, ok := res.(*statefulSessionResource)
	assert.True(t, ok)
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.validator)
	assert.NotNil(t, resource.service)
	assert.Equal(t, srv, resource.service)

	resourceKey := resource.GetKey()
	assert.NotNil(t, resourceKey)
	assert.Equal(t, "StatefulSession", resourceKey.Kind)
	assert.Equal(t, "nc.core.mesh/v3", resourceKey.APIVersion)
}

func TestEnrichSpecWithNamespace(t *testing.T) {
	err := enrichSpecWithNamespace(&dto.StatefulSession{
		Version:   "v1",
		Namespace: "",
		Cluster:   "cluster",
	}, configresources.Metadata{
		"namespace": 1,
	})
	assert.NotNil(t, err)

	ses := &dto.StatefulSession{
		Version:   "v1",
		Namespace: "",
		Cluster:   "cluster",
	}
	err = enrichSpecWithNamespace(ses, configresources.Metadata{
		"namespace": "wat",
	})
	assert.Nil(t, err)
	assert.Equal(t, "wat", ses.Namespace)
}

func TestService_SingleOverriddenWithTrueValueForRoutingConfigRequestV3(t *testing.T) {
	resource := statefulSessionResource{}
	isOverridden := resource.GetDefinition().IsOverriddenByCR(nil, nil, &dto.StatefulSession{
		Version:    "v1",
		Namespace:  "",
		Cluster:    "cluster",
		Overridden: true,
	})
	assert.True(t, isOverridden)
}
