package config

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateAndSaveInternalGatewayRoutesReturnsError_OnPutRoutesFails(t *testing.T) {
	serviceMock := &EntityServiceMock{}
	err := createAndSaveInternalGatewayRoutes(nil, serviceMock, 0, "", "")
	assert.Error(t, err)
}

func Test_CreateAndSavePrivateGatewayRoutesReturnsError_OnPutRoutesFails(t *testing.T) {
	serviceMock := &EntityServiceMock{}
	err := createAndSavePrivateGatewayRoutes(nil, serviceMock, 0, "", "")
	assert.Error(t, err)
}

func Test_CreateAndSavePublicGatewayRoutesReturnsError_OnPutRoutesFails(t *testing.T) {
	serviceMock := &EntityServiceMock{}
	err := createAndSavePublicGatewayRoutes(nil, serviceMock, 0, "", "")
	assert.Error(t, err)
}

type EntityServiceMock struct {
	entity.Service
}

func (srv *EntityServiceMock) PutRoutes(dao dao.Repository, routes []*domain.Route) error {
	return fmt.Errorf("PutRoutes has failed")
}
