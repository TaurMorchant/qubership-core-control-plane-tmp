package db

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-lib-go-dbaas-postgres-client/v4/model"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
)

func TestListen(t *testing.T) {
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	defaultDBProvider := &DefaultDBProvider{
		serviceDB: &stubServiceDb{
			connectionProperties: &model.PgConnProperties{
				Url:      "postgresql://localhost:5432/mydb",
				Username: "username",
				Password: "password",
			},
		},
	}

	listener, err := defaultDBProvider.Listen("channel", stubInitStorageAfterConnect, stubProcessNotification)
	assert.Nil(t, err)
	assert.NotNil(t, listener)

	defaultDBProvider = &DefaultDBProvider{
		serviceDB: &stubServiceDb{
			connectionProperties: &model.PgConnProperties{
				Url:      "postgresql://localhost",
				Username: "username",
				Password: "password",
			},
		},
	}

	assert.Panics(t, func() {
		defaultDBProvider.Listen("channel", stubInitStorageAfterConnect, stubProcessNotification)
	})
}

func stubInitStorageAfterConnect() {
}
func stubProcessNotification(payload string) {
}
