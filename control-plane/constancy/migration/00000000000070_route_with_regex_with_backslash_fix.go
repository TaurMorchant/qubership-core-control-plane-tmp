package migration

import (
	"context"
	"database/sql"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#70 route with regex with backslash fix")

			if err := fixRoutesWithRegexpWithBackslash(&tx, ctx); err != nil {
				return err
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return err
			}

			log.Info("Route with regex with backslash fix (migration #70) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func fixRoutesWithRegexpWithBackslash(db *bun.Tx, ctx context.Context) error {
	routesWithRegexp := make([]*V16Route, 0)
	err := db.NewSelect().
		Model(&routesWithRegexp).
		Relation("HeaderMatchers").
		Where("rm_regexp is not null").
		Scan(ctx)

	if err != nil {
		log.Errorf("Error selecting routes by non-empty regexp: %v", err)
		return err
	}

	for _, r := range routesWithRegexp {
		if r.Regexp != "" {
			r.Regexp = strings.ReplaceAll(r.Regexp, "\\", "")
		}
		r.RouteKey = generateRouteKeyV18(r)
	}

	uniqueRoutes, routeDuplicates := findRouteDuplicates(routesWithRegexp)
	if err := deleteRoutesAndHeaderMatchers(ctx, db, routeDuplicates); err != nil {
		return err
	}
	for _, route := range uniqueRoutes {
		if _, err = db.NewUpdate().Model(route).Column("routekey", "rm_regexp", "ra_regexprewrite").WherePK().Exec(ctx); err != nil {
			log.Errorf("Error updating routes with non-empty regexp: %v", err)
			return err
		}
	}
	return nil
}
