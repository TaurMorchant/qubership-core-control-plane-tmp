package migration

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#21 route regexp rewrite fix")

			if err := fixBrokenRouteRegexRewriteV21(&tx, ctx); err != nil {
				return errors.Wrap(err, 0)
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return errors.Wrap(err, 0)
			}

			log.Info("route regexp rewrite fix (migration #21) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

// Broken ra_regexpRewrite example: "/api/v1/ext-frontend-api/customers_vars/.*/subscriptions_vars/.*/test\1";
//
// fixed ra_regexpRewrite example: "/api/v1/ext-frontend-api/customers_vars/\1/subscriptions_vars/\2/test\3".
func fixBrokenRouteRegexRewriteV21(db *bun.Tx, ctx context.Context) error {
	routesWithRegexpRewrite := make([]*V16Route, 0)
	err := db.NewSelect().Model(&routesWithRegexpRewrite).Where("ra_regexprewrite is not null").Scan(ctx)
	if err != nil {
		log.Errorf("Couldn't load routes with non-null ra_regexprewrite: %v", err)
		return errors.Wrap(err, 0)
	}

	for _, r := range routesWithRegexpRewrite {
		if r.RegexpRewrite != "" && strings.Contains(r.RegexpRewrite, ".*") {
			r.RegexpRewrite = fixRouteRegexRewriteV21(r.RegexpRewrite)
			_, err = db.NewUpdate().Model(r).Column("ra_regexprewrite").WherePK().Exec(ctx)
			if err != nil {
				log.Errorf("couldn't update ra_regexpRewrite for route with id %d: %v", r.Id, err)
				return errors.Wrap(err, 0)
			}
		}
	}
	return nil
}

func fixRouteRegexRewriteV21(regexpRewrite string) string {
	brokenRegexpNum := strings.Count(regexpRewrite, ".*")
	for orderNum := 1; orderNum <= brokenRegexpNum; orderNum++ {
		regexpRewrite = strings.Replace(regexpRewrite, ".*", fmt.Sprintf("\\%d", orderNum), 1)
	}
	return fmt.Sprintf("%s\\%d", regexpRewrite[:len(regexpRewrite)-2], brokenRegexpNum+1)
}
