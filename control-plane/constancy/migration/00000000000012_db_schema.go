package migration

import (
	"context"
	"database/sql"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Infof("#12 DB schema")
			entities := []interface{}{
				(*domain.EnvoyConfigVersion)(nil),
			}

			for _, entity := range entities {
				_, err := tx.NewCreateTable().IfNotExists().Model(entity).WithForeignKeys().Exec(ctx)
				if err != nil {
					panic(err)
				}
			}

			log.Infof("Crating table '%s'...", clustering.ElectionTableName)

			electionRecord := &clustering.MasterMetadata{
				Name: "internal",
				NodeInfo: clustering.NodeInfo{
					IP:       "0.0.0.0",
					SWIMPort: 0,
					BusPort:  0,
				},
			}
			if _, err := tx.NewCreateTable().Model(electionRecord).IfNotExists().Exec(ctx); err != nil {
				return err
			}
			if _, err := tx.NewInsert().Model(electionRecord).Value("sync_clock", "now() - interval '1 hour'").Exec(ctx); err != nil {
				return err
			}
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
