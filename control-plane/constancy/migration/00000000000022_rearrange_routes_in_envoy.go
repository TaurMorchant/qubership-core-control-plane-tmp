package migration

import (
	"context"
	"database/sql"
	"github.com/go-errors/errors"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#22 rearrange routes order in envoy routing table")

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return errors.Wrap(err, 0)
			}

			log.Info("rearrange routes order in envoy (migration #22) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}
