package migration

import (
	"context"
	"database/sql"
	"github.com/uptrace/bun"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#91 fill default endpoint protocol")
			if err := fillDefaultEndpointProtocol(ctx, &tx); err != nil {
				return err
			}
			log.Info("Default endpoint protocol (migration #91) filled successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func fillDefaultEndpointProtocol(ctx context.Context, tx *bun.Tx) error {
	endpointsWithoutProto := make([]*V91Endpoint, 0)
	err := tx.NewSelect().Model(&endpointsWithoutProto).Relation("Cluster").Where("protocol is null").Scan(ctx)
	if err != nil {
		log.Errorf("Error selecting endpoints with null protocol:\n %v", err)
		return err
	}

	for _, endpoint := range endpointsWithoutProto {
		endpoint.Protocol = resolveEndpointProtocol(endpoint)
		if _, err := tx.NewUpdate().Model(endpoint).WherePK().Exec(ctx); err != nil {
			log.Errorf("Error updating endpoint protocol:\n %v", err)
			return err
		}
	}
	return nil
}

func resolveEndpointProtocol(endpoint *V91Endpoint) string {
	if endpoint.Cluster.TLSId == 0 {
		return "http"
	} else {
		return "https"
	}
}

type V91Endpoint struct {
	bun.BaseModel            `bun:"endpoints"`
	Id                       int32       `bun:",pk" json:"id"`
	Address                  string      `bun:",notnull"`
	Port                     int32       `bun:",notnull"`
	Protocol                 string      `bun:",notnull"`
	ClusterId                int32       `bun:"clusterid,notnull"`
	Cluster                  *v91Cluster `bun:"rel:belongs-to,join:clusterid=id"`
	DeploymentVersion        string      `bun:"deployment_version,notnull"`
	InitialDeploymentVersion string      `bun:"initialdeploymentversion,notnull"`
	Hostname                 string      `bun:"hostname"`
	OrderId                  int32       `bun:"order_id"`
}

type v91Cluster struct {
	bun.BaseModel `bun:"clusters"`
	Id            int32  `bun:",pk" json:"id"`
	Name          string `bun:",notnull" json:"name"`
	LbPolicy      string `bun:"lbpolicy,notnull" json:"lbPolicy"`
	DiscoveryType string `bun:"type,notnull" json:"type"`
	Version       int32  `bun:",notnull"`
	HttpVersion   *int32 `bun:"http_version,notnull,default:1" json:"httpVersion"`
	EnableH2      bool   `bun:"enableh2,default:false" json:"enableH2"`
	TLSId         int32  `bun:"tls_id,notnull"`
}
