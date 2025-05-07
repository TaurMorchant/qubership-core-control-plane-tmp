package migration

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Infof("#6 Set column 'initial_deployment_version'")
			var routes []*v6Route
			err := tx.NewSelect().Model(&routes).Where("initialdeploymentversion is null").Scan(ctx)
			if err != nil {
				return errors.Wrap(err, "Finding routes where initialdeploymentversion is null has failed")
			}
			for _, route := range routes {
				route.InitialDeploymentVersion = route.DeploymentVersion
				_, err := tx.NewUpdate().Model(route).WherePK().Exec(ctx)
				if err != nil {
					return errors.Wrapf(err, "Updating column 'initialdeploymentversion' for route '%v' has failed", *route)
				}
			}
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

type v6Route struct {
	bun.BaseModel            `bun:"routes"`
	Id                       int32
	DeploymentVersion        string `bun:"deployment_version"`
	InitialDeploymentVersion string `bun:"initialdeploymentversion"`
}
