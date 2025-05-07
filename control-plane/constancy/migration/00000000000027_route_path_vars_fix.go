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
			log.Info("#27 route path variables migration")

			if err := fixRegexpPathVars(&tx, ctx); err != nil {
				return err
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return err
			}

			log.Info("Route path variables fix (migration #27) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func fixRegexpPathVars(db *bun.Tx, ctx context.Context) error {
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

	removeOldStylePathVars(routesWithRegexp)
	uniqueRoutes, routesForDeletion := splitToUniqueAndDuplicatesByRegexpKey(routesWithRegexp)

	for _, route := range uniqueRoutes {
		if _, err = db.NewUpdate().Model(route).Column("rm_regexp").WherePK().Exec(ctx); err != nil {
			log.Errorf("Error updating routes with non-empty regexp: %v", err)
			return err
		}
	}

	if err := deleteRoutesAndHeaderMatchers(ctx, db, routesForDeletion); err != nil {
		return err
	}

	return nil
}

func removeOldStylePathVars(routes []*V16Route) {
	for _, route := range routes {
		withoutTrail := strings.TrimSuffix(route.Regexp, "(/.*)?")
		newStyledPathVarsWithoutTrail := strings.ReplaceAll(withoutTrail, ".*", "([^/]+)")
		route.Regexp = newStyledPathVarsWithoutTrail + "(/.*)?"
	}
}

// Due to not accurate previous migration (#18) we can face an situation when there are old-style and new-style routes
// exists in database, so after migration from old-style to new-style they are the same.
// Let's find them and drop duplicate routes
func splitToUniqueAndDuplicatesByRegexpKey(routes []*V16Route) ([]*V16Route, []*V16Route) {
	regexps := make(map[string][]*V16Route, len(routes))
	uniqueRegexpRoutes := make([]*V16Route, 0, len(routes))
	regexpRoutesForDeletion := make([]*V16Route, 0, len(routes))

	contains := func(routes []*V16Route, target *V16Route) bool {
		for _, r := range routes {
			if target.Regexp == r.Regexp && target.Prefix == r.Prefix && target.DeploymentVersion == r.DeploymentVersion &&
				target.PrefixRewrite == r.PrefixRewrite && target.RegexpRewrite == r.RegexpRewrite &&
				target.VirtualHostId == r.VirtualHostId && target.ClusterName == r.ClusterName && target.Version == r.Version {
				return true
			}
		}
		return false
	}
	addToUnique := func(target *V16Route) {
		if len(regexps[target.Regexp]) == 0 {
			regexps[target.Regexp] = make([]*V16Route, 0)
		}
		regexps[target.Regexp] = append(regexps[target.Regexp], target)
		uniqueRegexpRoutes = append(uniqueRegexpRoutes, target)
	}

	for _, route := range routes {
		if _, ok := regexps[route.Regexp]; !ok {
			addToUnique(route)
		} else {
			if contains(regexps[route.Regexp], route) {
				regexpRoutesForDeletion = append(regexpRoutesForDeletion, route)
			} else {
				addToUnique(route)
			}
		}
	}

	return uniqueRegexpRoutes, regexpRoutesForDeletion
}
