package migration

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/route/routekey"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"strings"
	"time"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#18 route regex fix")

			if err := fixRoutesWithRegexp(ctx, &tx); err != nil {
				return err
			}

			if err := fixRouteKeyForNonRegexpRoutes(ctx, &tx); err != nil {
				return err
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return err
			}

			log.Info("Route regex fix (migration #18) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func fixRoutesWithRegexp(ctx context.Context, db *bun.Tx) error {
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
		processRegexp(r)
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

// fixRouteKeyForNonRegexpRoutes fixes bug in migration #16: header matchers were not used for route key building.
func fixRouteKeyForNonRegexpRoutes(ctx context.Context, db *bun.Tx) error {
	log.Info("Applying fix for bug in migration #16: header matchers were not used for route key building")
	routesWithoutRegexp := make([]*V16Route, 0)
	err := db.NewSelect().
		Model(&routesWithoutRegexp).
		Relation("HeaderMatchers").
		Where("rm_regexp is null").
		Scan(ctx)
	if err != nil {
		log.Errorf("Error selecting routes by empty regexp: %v", err)
		return err
	}

	for _, r := range routesWithoutRegexp {
		r.RouteKey = generateRouteKeyV18(r)
	}

	uniqueRoutes, routeDuplicates := findRouteDuplicates(routesWithoutRegexp)
	if err := deleteRoutesAndHeaderMatchers(ctx, db, routeDuplicates); err != nil {
		return err
	}
	for _, route := range uniqueRoutes {
		if _, err = db.NewUpdate().Model(route).Column("routekey").WherePK().Exec(ctx); err != nil {
			log.Errorf("Error fixing routekey for route without regexp: %v", err)
			return err
		}
	}
	return nil
}

// findRouteDuplicates find duplicated routes in a slice.
// First return value is a slice of unique routes that need to be saved;
// second return value is a slice of duplicates to be deleted.
//
// Since regexp matcher now guaranteed to end with a regexp match group instead of trailing slash,
// some of the existing routes can duplicate each other after applying these changes to them.
func findRouteDuplicates(routesToUpdate []*V16Route) ([]*V16Route, []*V16Route) {
	routesByVirtualHost := make(map[int32][]*V16Route, len(routesToUpdate))
	for _, route := range routesToUpdate {
		if _, sliceExists := routesByVirtualHost[route.VirtualHostId]; !sliceExists {
			routesByVirtualHost[route.VirtualHostId] = make([]*V16Route, 0)
		}
		routesByVirtualHost[route.VirtualHostId] = append(routesByVirtualHost[route.VirtualHostId], route)
	}

	uniqueRoutes := make([]*V16Route, 0)
	duplicates := make([]*V16Route, 0)

	for _, routes := range routesByVirtualHost {
		duplicatesInVH := make([]*V16Route, 0)

		for _, route := range routes {
			alreadyMarkedAsDuplicate := false
			for _, duplicate := range duplicatesInVH {
				if route.Id == duplicate.Id {
					alreadyMarkedAsDuplicate = true
					break
				}
			}

			if !alreadyMarkedAsDuplicate {
				uniqueRoutes = append(uniqueRoutes, route)

				for _, anotherRoute := range routes {
					if route.Id != anotherRoute.Id && route.RouteKey == anotherRoute.RouteKey {
						duplicatesInVH = append(duplicatesInVH, anotherRoute)
					}
				}
			}
		}

		duplicates = append(duplicates, duplicatesInVH...)
	}

	return uniqueRoutes, duplicates
}

func deleteRoutesAndHeaderMatchers(ctx context.Context, db *bun.Tx, routesToDelete []*V16Route) error {
	for _, route := range routesToDelete {
		log.Infof("Deleting route since it is a duplicate: %v", *route)
		if _, err := db.Exec("DELETE FROM header_matchers WHERE routeid = ?", route.Id); err != nil {
			log.Errorf("Error deleting duplicated route header matchers: %v", err)
			return err
		}
		if _, err := db.NewDelete().Model(route).WherePK().Exec(ctx); err != nil {
			log.Errorf("Error deleting duplicated route: %v", err)
			return err
		}
	}
	return nil
}

func updateEnvoyRouteConfigVersion(ctx context.Context, db *bun.Tx) error {
	newVersion := time.Now().UnixNano()
	log.Infof("Updating envoy config version for all RouteConfigs with version %v", newVersion)

	_, err := db.Exec("UPDATE envoy_config_versions "+
		"SET version = ? "+
		"WHERE entity_type = ?", newVersion, domain.RouteConfigurationTable)

	if err != nil {
		log.Errorf("Failed to update envoy config version for all RouteConfigs: %v", err)
		return errors.Wrap(err, "Cannot update envoy config version for all RouteConfigs")
	}
	return nil
}

func generateRouteKeyV18(r *V16Route) string {
	return routekey.GenerateFunc(func() routekey.RouteMatch {
		rm := routekey.RouteMatch{
			Prefix:  r.Prefix,
			Regexp:  r.Regexp,
			Path:    r.Path,
			Version: r.DeploymentVersion,
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

func processRegexp(route *V16Route) {
	if route.Regexp == "" {
		return
	}

	varsNumber := strings.Count(route.Regexp, "(.*)")
	if varsNumber == 0 {
		return
	}

	route.Regexp = strings.ReplaceAll(route.Regexp, "(.*)", "([^/]+)")
	for strings.HasSuffix(route.Regexp, "/") {
		route.Regexp = route.Regexp[:len(route.Regexp)-1]
	}
	route.Regexp = route.Regexp + "(/.*)?"

	if route.RegexpRewrite != "" {
		for strings.HasSuffix(route.RegexpRewrite, "/") {
			route.RegexpRewrite = route.RegexpRewrite[:len(route.RegexpRewrite)-1]
		}
		route.RegexpRewrite = fmt.Sprintf("%s\\%v", route.RegexpRewrite, varsNumber+1)
	}
}
