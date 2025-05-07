package migration

import (
	"context"
	"database/sql"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Infof("create tables: RetryPolicy, HealthCheck")
			entities := []interface{}{
				(*domain.RetryPolicy)(nil),
				(*domain.HealthCheck)(nil),
			}
			for _, entity := range entities {
				_, err := tx.NewCreateTable().Model(entity).IfNotExists().WithForeignKeys().Exec(ctx)
				if err != nil {
					panic(err)
				}
			}
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
