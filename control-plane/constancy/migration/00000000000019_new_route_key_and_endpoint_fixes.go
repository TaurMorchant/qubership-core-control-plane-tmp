package migration

import (
	"context"
	"database/sql"
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/services/route/routekey"
	"github.com/uptrace/bun"
	"regexp"
	"time"
)

var versionFromAddressRegexp = regexp.MustCompile("^.*-(v[[:digit:]]+)$")

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#19 new route key and endpoint fixes")

			if err := removeDuplicatedControlPlaneHealthRoute(&tx); err != nil {
				return errors.Wrap(err, 0)
			}

			if err := addAndFillInitialVersionForEndpoints(&tx, ctx); err != nil {
				return errors.Wrap(err, 0)
			}

			if err := updateInitialVersionForRoutes(&tx, ctx); err != nil {
				return errors.Wrap(err, 0)
			}

			if err := regenerateRouteKey(&tx, ctx); err != nil {
				return errors.Wrap(err, 0)
			}

			var activeVersion V19DeploymentVersion
			if err := tx.NewSelect().Model(&activeVersion).Where("stage = ?", domain.ActiveStage).Scan(ctx); err != nil {
				log.Errorf("couldn't find ACTIVE deployment_version: %v", err)
				return errors.Wrap(err, 0)
			}

			if err := removeEndpointsDuplicate(&tx, ctx, activeVersion); err != nil {
				return errors.Wrap(err, 0)
			}

			var legacyVersion V19DeploymentVersion
			if err := tx.NewSelect().Model(&legacyVersion).Where("stage = ?", domain.LegacyStage).Scan(ctx); err == nil {
				if err := moveRoutesToActualVersion(&tx, ctx, legacyVersion, activeVersion); err != nil {
					return errors.Wrap(err, 0)
				}
			} else {
				if !errors.Is(sql.ErrNoRows, err) {
					return errors.Wrap(err, 0)
				}
				log.Warnf("Legacy version has not found, Skip moving routes to actual version")
			}

			if err := updateEnvoyRouteConfigVersion(ctx, &tx); err != nil {
				return errors.Wrap(err, 0)
			}
			if err := updateEnvoyClusterVersion(&tx); err != nil {
				return errors.Wrap(err, 0)
			}

			log.Info("New route key and endpoint fixes (migration #19) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func removeDuplicatedControlPlaneHealthRoute(db *bun.Tx) interface{} {
	_, err := db.Exec(`delete from routes where virtualhostid in 
(select id from virtual_hosts where name in ('private-gateway-service', 'public-gateway-service', 'internal-gateway-service'))
and rm_prefix = '/health' and (ra_clustername = '||control-plane||control-plane||8080' OR ra_clustername is null)`)
	if err != nil {
		log.Errorf("couldn't delete control-plane routes with ra_clustername '||control-plane||control-plane||8080': %v", err)
		return errors.Wrap(err, 1)
	}
	return nil
}

func moveRoutesToActualVersion(db *bun.Tx, ctx context.Context, legacyVersion V19DeploymentVersion, activeVersion V19DeploymentVersion) error {
	var clusters []V19Cluster
	if err := db.NewSelect().Model(&clusters).Relation("Endpoints").Scan(ctx); err != nil {
		log.Errorf("Failed to load clusters with endpoints from pg: %v", err)
		return errors.Wrap(err, 1)
	}
	for _, cluster := range clusters {
		legacyRoutesCount, err := db.NewSelect().
			Model(&V16Route{}).
			Where("deployment_version = ? AND ra_clustername = ?", legacyVersion.Version, cluster.Name).
			Count(ctx)
		if err != nil {
			log.Errorf("couldn't find route by deployment_version = '%s' and ra_clustername = '%s': %v", legacyVersion.Version, cluster.Name, err)
			return err
		}

		if legacyRoutesCount > 0 {
			activeRoutesCount, err := db.NewSelect().
				Model(&V16Route{}).
				Where("initialdeploymentversion = ? AND ra_clustername = ?", activeVersion.Version, cluster.Name).
				Count(ctx)
			if err != nil {
				log.Errorf("couldn't find route by initialdeploymentversion = '%s' and ra_clustername = '%s': %v", activeVersion.Version, cluster.Name, err)
				return err
			}
			if activeRoutesCount == 0 {
				_, err := db.NewUpdate().Model(&V16Route{}).Set("deployment_version = ?", activeVersion.Version).
					Where("ra_clustername = ? AND deployment_version = ?", cluster.Name, legacyVersion.Version).
					Exec(ctx)
				if err != nil {
					log.Errorf("couldn't update routes with active deployment_version '%s': %v", activeVersion.Version, err)
				}
			}
		}
	}
	return nil
}

func removeEndpointsDuplicate(db *bun.Tx, ctx context.Context, activeVersion V19DeploymentVersion) error {
	addressEndpointsMap := make(map[string][]V19Endpoint)
	var endpoints []V19Endpoint
	err := db.NewSelect().Model(&endpoints).Scan(ctx)
	if err != nil {
		log.Errorf("couldn't find all endpoints: %v", err)
		return errors.Wrap(err, 1)
	}
	for _, endpoint := range endpoints {
		key := endpoint.Address + string(endpoint.Port)
		addressEndpointsMap[key] = append(addressEndpointsMap[key], endpoint)
	}
	for _, endpoints := range addressEndpointsMap {
		if len(endpoints) > 1 {
			for _, endpoint := range endpoints {
				if endpoint.DeploymentVersion != activeVersion.Version {
					_, err := db.NewDelete().Model(&endpoint).WherePK().Exec(ctx)
					if err != nil {
						log.Errorf("couldn't delete route with PK '%d': %v", endpoint.Id, err)
						return err
					}
				}
			}
		}
	}
	return nil
}

func updateInitialVersionForRoutes(db *bun.Tx, ctx context.Context) error {
	var clusters []V19Cluster
	if err := db.NewSelect().Model(&clusters).Relation("Endpoints").Scan(ctx); err != nil {
		log.Errorf("Failed to load clusters with endpoints from pg: %v", err)
		return errors.Wrap(err, 0)
	}

	for _, cluster := range clusters {
		for _, endpoint := range cluster.Endpoints {
			result, err := db.NewUpdate().
				Model(&V16Route{}).
				Set("initialdeploymentversion = ?", endpoint.InitialDeploymentVersion).
				Where("ra_clustername = ? AND deployment_version = ?", cluster.Name, endpoint.DeploymentVersion).
				Exec(ctx)
			if err != nil {
				log.Errorf("couldn't update initialdeploymentversion for endpoints: %v", err)
				return errors.Wrap(err, 0)
			}
			rowsAffected, err := result.RowsAffected()
			if err == nil {
				log.Infof("Update initialdeploymenversion with value '%s', for '%d' routes", endpoint.InitialDeploymentVersion, rowsAffected)
			}
		}
	}
	return nil
}

func regenerateRouteKey(db *bun.Tx, ctx context.Context) error {
	routesToUpdate := make([]*V16Route, 0)
	err := db.NewSelect().Model(&routesToUpdate).Relation("HeaderMatchers").Scan(ctx)
	if err != nil {
		log.Errorf("couldn't find routes with relations: %v", err)
		return errors.Wrap(err, 0)
	}
	for _, r := range routesToUpdate {
		r.RouteKey = generateRouteKeyV19(r)
		_, err = db.NewUpdate().Model(r).Column("routekey").WherePK().Exec(ctx)
		if err != nil {
			log.Errorf("couldn't update routekey for route with id %d: %v", r.Id, err)
			return errors.Wrap(err, 0)
		}
	}
	return nil
}

func addAndFillInitialVersionForEndpoints(db *bun.Tx, ctx context.Context) error {
	_, err := db.Exec("ALTER TABLE endpoints ADD COLUMN IF NOT EXISTS initialdeploymentversion varchar(64);")
	if err != nil {
		err = errors.WrapPrefix(err, "could not add initialdeploymentversion column to endpoints table", 1)
		log.Errorf("%v", err)
		return errors.Wrap(err, 0)
	}

	var endpoints []*V19Endpoint
	if err := db.NewSelect().Model(&endpoints).Scan(ctx); err != nil {
		log.Errorf("couldn't find endpoints: %v", err)
		return errors.Wrap(err, 0)
	}
	for _, endpoint := range endpoints {
		subMatches := versionFromAddressRegexp.FindStringSubmatch(endpoint.Address)
		if len(subMatches) == 0 {
			endpoint.InitialDeploymentVersion = "v1"
		} else {
			endpoint.InitialDeploymentVersion = subMatches[1]
		}
	}

	if len(endpoints) > 0 {
		if _, err = db.NewUpdate().Model(&endpoints).Column("initialdeploymentversion").Bulk().Exec(ctx); err != nil {
			log.Errorf("Error setting initialdeploymentversion for endpoint: %v", err)
			return errors.Wrap(err, 1)
		}
	}
	return nil
}

func updateEnvoyClusterVersion(db *bun.Tx) error {
	newVersion := time.Now().UnixNano()
	log.Infof("Updating envoy config version for all Clusters with version %v", newVersion)

	_, err := db.Exec("UPDATE envoy_config_versions "+
		"SET version = ? "+
		"WHERE entity_type = ?", newVersion, domain.ClusterTable)

	if err != nil {
		log.Errorf("Failed to update envoy config version for all Clusters: %v", err)
		return errors.WrapPrefix(err, "Cannot update envoy config version for all Clusters", 1)
	}
	return nil
}

func generateRouteKeyV19(r *V16Route) string {
	return routekey.GenerateFunc(func() routekey.RouteMatch {
		rm := routekey.RouteMatch{
			Prefix: r.Prefix,
			Regexp: r.Regexp,
			Path:   r.Path,
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

type V19Cluster struct {
	bun.BaseModel `bun:"table:clusters"`
	Id            int32         `bun:",pk" json:"id"`
	Name          string        `bun:",notnull" json:"name"`
	LbPolicy      string        `bun:"lbpolicy,notnull" json:"lbPolicy"`
	DiscoveryType string        `bun:"column:type,notnull" json:"type"`
	Version       int32         `bun:",notnull"`
	EnableH2      bool          `bun:"enableh2,nullzero,notnull,default:false" json:"enableH2"`
	Endpoints     []V19Endpoint `bun:"rel:has-many,join:id=clusterid"`
}

type V19Endpoint struct {
	bun.BaseModel            `bun:"table:endpoints"`
	Id                       int32  `bun:",pk"`
	Address                  string `bun:",notnull"`
	Port                     int32  `bun:",notnull"`
	ClusterId                int32  `bun:"clusterid,notnull"`
	DeploymentVersion        string `bun:"deployment_version,notnull"`
	InitialDeploymentVersion string `bun:"initialdeploymentversion,notnull"`
}

type V19DeploymentVersion struct {
	bun.BaseModel `bun:"table:deployment_versions"`
	Version       string `bun:",pk" json:"version"`
	Stage         string `bun:",notnull" json:"stage"`
}
