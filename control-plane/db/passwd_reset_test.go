package db

import (
	"context"
	"sync"
	"testing"

	"github.com/go-errors/errors"
	"github.com/golang/mock/gomock"
	pgdbaas "github.com/netcracker/qubership-core-lib-go-dbaas-postgres-client/v4"
	"github.com/netcracker/qubership-core-lib-go-dbaas-postgres-client/v4/model"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func TestNewDB(t *testing.T) {
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

	db, err := defaultDBProvider.NewDB(stubEvent)
	assert.Nil(t, err)
	assert.NotNil(t, db)

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
		defaultDBProvider.NewDB(stubEvent)
	})

	defaultDBProvider = &DefaultDBProvider{
		serviceDB: &stubServiceDb{
			connectionProperties: &model.PgConnProperties{
				Url:      "postgresql://localhost:localhost:localhost",
				Username: "username",
				Password: "password",
			},
		},
	}

	_, err = defaultDBProvider.NewDB(stubEvent)
	assert.NotNil(t, err)

	defaultDBProvider = &DefaultDBProvider{
		serviceDB: &stubServiceDb{
			connectionPropertiesError: true,
		},
	}

	db, err = defaultDBProvider.NewDB(stubEvent)
	assert.NotNil(t, err)
}

func TestPGDbWithPasswordReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pgDbWithPasswordReset := &pgDbWithPasswordReset{
		dbMux: &sync.RWMutex{},
		pgDB:  &bun.DB{},
	}

	pgDB := pgDbWithPasswordReset.Get()
	assert.Equal(t, pgDbWithPasswordReset.pgDB, pgDB)
}

type stubServiceDb struct {
	connectionProperties      *model.PgConnProperties
	connectionPropertiesError bool
}

func (db stubServiceDb) GetPgClient(options ...*model.PgOptions) (pgdbaas.PgClient, error) {
	return nil, nil
}
func (db stubServiceDb) GetConnectionProperties(ctx context.Context) (*model.PgConnProperties, error) {
	if db.connectionPropertiesError {
		return nil, errors.New("test error")
	}
	return db.connectionProperties, nil
}
func (db stubServiceDb) FindConnectionProperties(ctx context.Context) (*model.PgConnProperties, error) {
	return nil, nil
}

func stubEvent(db PgDB, event ConnEvent, err error) {

}
