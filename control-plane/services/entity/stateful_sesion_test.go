package entity

import (
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_PutStatefulSession(t *testing.T) {
	entityService, inMemDao := getService(t)
	session := &domain.StatefulSession{
		CookieName:               "",
		CookiePath:               "/",
		Enabled:                  true,
		ClusterName:              "test-cluster",
		Namespace:                "default",
		Gateways:                 []string{"private-gateway-service"},
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	}

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		assert.Nil(t, entityService.PutStatefulSession(dao, session))
		return nil
	})
	assert.Nil(t, err)

	actualSessions, err := inMemDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.NotEmpty(t, actualSessions)
	assert.Equal(t, 1, len(actualSessions))
	assert.Contains(t, actualSessions, session)
}
