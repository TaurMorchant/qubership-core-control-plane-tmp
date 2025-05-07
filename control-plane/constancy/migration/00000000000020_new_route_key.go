package migration

import (
	"context"
	"database/sql"
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/services/route/routekey"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#20 new route key")

			if err := regenerateRouteKeyV20(&tx, ctx); err != nil {
				return errors.Wrap(err, 0)
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return errors.Wrap(err, 0)
			}

			log.Info("New route key (migration #20) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func regenerateRouteKeyV20(db *bun.Tx, ctx context.Context) error {
	routesToUpdate := make([]*V16Route, 0)
	err := db.NewSelect().Model(&routesToUpdate).Relation("HeaderMatchers").Scan(ctx)
	if err != nil {
		log.Errorf("couldn't find routes with relations: %v", err)
		return errors.Wrap(err, 0)
	}
	for _, r := range routesToUpdate {
		r.RouteKey = generateRouteKeyV20(r)
		_, err = db.NewUpdate().Model(r).Column("routekey").WherePK().Exec(ctx)
		if err != nil {
			log.Errorf("couldn't update routekey for route with id %d: %v", r.Id, err)
			return errors.Wrap(err, 0)
		}
	}
	return nil
}

func generateRouteKeyV20(r *V16Route) string {
	return routekey.GenerateFunc(func() routekey.RouteMatch {
		rm := routekey.RouteMatch{
			Prefix:  r.Prefix,
			Regexp:  r.Regexp,
			Path:    r.Path,
			Version: r.InitialDeploymentVersion,
		}
		headers := make([]routekey.HeaderMatch, len(r.HeaderMatchers))
		for i, hm := range r.HeaderMatchers {
			headers[i] = routekey.HeaderMatch{
				Name:           hm.Name,
				ExactMatch:     hm.ExactMatch,
				SafeRegexMatch: hm.SafeRegexMatch,
				PresentMatch:   hm.PresentMatch,
				PrefixMatch:    hm.PrefixMatch,
				SuffixMatch:    hm.SuffixMatch,
				InvertMatch:    hm.InvertMatch,
			}
			if hm.RangeMatch != nil {
				headers[i].RangeMatch = &routekey.RangeMatch{
					Start: hm.RangeMatch.Start.Int64,
					End:   hm.RangeMatch.End.Int64,
				}
			}
		}
		rm.Headers = headers
		return rm
	})
}
