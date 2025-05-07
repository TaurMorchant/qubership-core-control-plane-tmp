package migration

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

func getOOBNodeGroups() []string {
	return []string{"public-gateway-service", "private-gateway-service", "internal-gateway-service"}
}

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			col_name, err := tx.Query("SELECT column_name FROM information_schema.columns WHERE table_name='clusters' and column_name='host'")
			if CountRows(col_name) == 0 {
				log.Infof("It looks like migration # 4 was done earlier. Skip migration #4.")
				return nil
			}
			log.Infof("#4 New format for clusters and routes")
			var routes []*v4Route
			log.Debugf("Started updating routes")
			err = tx.NewSelect().Model(&routes).Scan(ctx)
			if err != nil {
				return errors.Wrap(err, "Finding all routes has failed")
			}
			for _, route := range routes {
				route.DeploymentVersion = "v1"
				route.RouteKey = route.RouteKey + "||v1"
				_, err := tx.NewUpdate().Model(route).WherePK().Exec(ctx)
				if err != nil {
					return errors.Wrapf(err, "Updating columns 'deployment_version' and 'routeKey' for route '%v' has failed", *route)
				}
			}
			log.Debugf("Routes updated successfully.")

			log.Debugf("Started updating clusters")
			var clusters []*v4Cluster
			err = tx.NewSelect().Model(&clusters).Scan(ctx)
			if err != nil {
				return errors.Wrap(err, "Finding all clusters in old format has failed")
			}
			for _, cluster := range clusters {
				endpoint := &v4Endpoint{
					Address:           cluster.Host,
					Port:              cluster.Port,
					DeploymentVersion: "v1",
					ClusterId:         cluster.Id,
				}
				_, err = tx.NewInsert().Model(endpoint).Exec(ctx)
				if err != nil {
					return errors.Wrapf(err, "Inserting new endpoint '%v' has failed", *endpoint)
				}

				err := insertClusterNodegroup(&tx, cluster.Id, getOOBNodeGroups(), ctx)
				if err != nil {
					return err
				}
			}
			log.Debugf("Endpoints updated successfully")
			log.Debugf("Clusters -> NodeGroups relations updated successfully")

			if _, err := tx.Exec("ALTER TABLE clusters DROP COLUMN IF EXISTS nodeGroup;"); err != nil {
				return errors.Wrapf(err, "Dropping column nodeGroup for 'clusters' has failed")
			}
			if _, err := tx.Exec("ALTER TABLE clusters DROP COLUMN IF EXISTS host;"); err != nil {
				return errors.Wrapf(err, "Dropping column host for 'clusters' has failed")
			}
			if _, err := tx.Exec("ALTER TABLE clusters DROP COLUMN IF EXISTS port"); err != nil {
				return errors.Wrapf(err, "Dropping column port for 'clusters' has failed")
			}
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func insertClusterNodegroup(tx *bun.Tx, clusterId int32, nodeGroups []string, ctx context.Context) error {
	for _, nodeGroup := range nodeGroups {
		entity := v4ClustersNodeGroup{ClusterId: clusterId, NodeGroupName: nodeGroup}
		_, err := tx.NewInsert().Model(&entity).Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "Inserting clusters -> nodeGroups relation '%v' has failed", entity)
		}
	}
	return nil
}

type v4Route struct {
	bun.BaseModel     `bun:"routes"`
	Id                int32
	RouteKey          string `bun:"routekey"`
	DeploymentVersion string `bun:"deployment_version"`
}

type v4Cluster struct {
	bun.BaseModel `bun:"table:clusters"`
	Id            int32
	Host          string
	Port          int32
}

type v4Endpoint struct {
	bun.BaseModel     `bun:"endpoints"`
	Id                int32
	Address           string
	Port              int32
	DeploymentVersion string `bun:"deployment_version,notnull"`
	ClusterId         int32  `bun:"clusterid"`
}

type v4ClustersNodeGroup struct {
	bun.BaseModel `bun:"table:clusters_node_groups"`
	ClusterId     int32  `bun:"clusters_id,pk"`
	NodeGroupName string `bun:"nodegroups_name,pk"`
}
