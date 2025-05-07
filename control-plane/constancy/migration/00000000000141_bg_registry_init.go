package migration

import (
	"context"
	"database/sql"
	"github.com/netcracker/qubership-core-control-plane/util/msaddr"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	number := 141
	name := "blue-green versions registry initial migration"
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Infof("#%d %s", number, name)

			if err := initBlueGreenVersionsRegistry(ctx, &tx); err != nil {
				return err
			}

			log.Info("%s (migration #%d) applied successfully", name, number)
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func initBlueGreenVersionsRegistry(ctx context.Context, db *bun.Tx) error {
	clusters := make([]*V141Cluster, 0)
	if err := db.NewSelect().
		Model(&clusters).
		Relation("Endpoints").
		Scan(ctx); err != nil {
		log.Errorf("Error selecting all clusters: %v", err)
		return err
	}

	for _, cluster := range clusters {
		serviceName := v141ExtractFamilyName(cluster.Name)
		namespace := v141ExtractNamespace(cluster.Name).GetNamespace()
		for _, endpoint := range cluster.Endpoints {
			msVersion := &V141MicroserviceVersion{
				Name:                     serviceName,
				Namespace:                namespace,
				DeploymentVersion:        endpoint.DeploymentVersion,
				InitialDeploymentVersion: endpoint.InitialDeploymentVersion,
			}
			if err := insertMicroserviceVersionIfNotExist(ctx, db, msVersion); err != nil {
				log.ErrorC(ctx, "Failed to insert microservice version during migration:\n %v", err)
				return err
			}
		}
	}
	return nil
}

func insertMicroserviceVersionIfNotExist(ctx context.Context, db *bun.Tx, msVersion *V141MicroserviceVersion) error {
	log.DebugC(ctx, "Checking if %+v does not exist and needs to be inserted", *msVersion)
	exists, err := db.NewSelect().Model(msVersion).WherePK().Exists(ctx)
	if err != nil {
		log.ErrorC(ctx, "Failed to check microservice version %+v existence during migration:\n %v", *msVersion, err)
		return err
	}
	if !exists {
		log.DebugC(ctx, "Microservice version %+v does not exist and needs to be inserted", *msVersion)
		if _, err := db.NewInsert().Model(msVersion).Exec(ctx); err != nil {
			log.ErrorC(ctx, "Failed to insert microservice version %+v during migration:\n %v", *msVersion, err)
			return err
		}
		log.InfoC(ctx, "Inserted microservice version %+v during migration", *msVersion)
	}
	return nil
}

type V141Cluster struct {
	bun.BaseModel `bun:"table:clusters"`

	Id            int32  `bun:",pk" json:"id"`
	Name          string `bun:",notnull" json:"name"`
	LbPolicy      string `bun:"lbpolicy,notnull" json:"lbPolicy"`
	DiscoveryType string `bun:"column:type,notnull" json:"type"`
	Version       int32  `bun:",notnull"`
	HttpVersion   *int32 `bun:"http_version,nullzero,notnull,default:1" json:"httpVersion"`
	EnableH2      bool   `bun:"enableh2" json:"enableH2"`
	TLSId         int32  `bun:"tls_id,nullzero,notnull"`
	//TLS            *TlsConfig      `bun:"rel:belongs-to,join:tls_id=id" json:"tlsConfigName"`
	//NodeGroups     []*NodeGroup    `bun:"m2m:clusters_node_groups,join:Cluster=NodeGroup,notnull"`
	Endpoints []*V141Endpoint `bun:"rel:has-many,join:id=clusterid"`
}

type V141Endpoint struct {
	bun.BaseModel `bun:"table:endpoints"`

	Id                       int32        `bun:",pk"`
	Address                  string       `bun:",notnull"`
	Port                     int32        `bun:",notnull"`
	Protocol                 string       `bun:",scanonly"`
	ClusterId                int32        `bun:"clusterid,nullzero,notnull"`
	Cluster                  *V141Cluster `bun:"rel:belongs-to,join:clusterid=id"`
	DeploymentVersion        string       `bun:"deployment_version,nullzero,notnull"`
	InitialDeploymentVersion string       `bun:"initialdeploymentversion,notnull"`
	//DeploymentVersionVal     *DeploymentVersion `bun:"rel:belongs-to,join:deployment_version=version"`
	//HashPolicies             []*HashPolicy      `bun:"rel:has-many,join:id=endpointid"`
	Hostname          string `bun:"hostname"`
	OrderId           int32  `bun:"order_id"`
	StatefulSessionId int32  `bun:"statefulsessionid,nullzero,notnull" json:"statefulSessionId"`
	//StatefulSession          *StatefulSession   `bun:"rel:belongs-to,join:statefulsessionid=id" json:"statefulSession"`
}

type V141MicroserviceVersion struct {
	bun.BaseModel `bun:"table:microservice_versions"`

	Name                     string `bun:"name,pk"`
	Namespace                string `bun:"namespace,pk"`
	DeploymentVersion        string `bun:"deployment_version,nullzero,notnull"`
	InitialDeploymentVersion string `bun:"initial_version,notnull,pk"`
}

func v141ExtractFamilyName(clusterKey string) string {
	if idx := strings.Index(clusterKey, "||"); idx != -1 {
		return clusterKey[:idx]
	}
	return clusterKey
}

func v141ExtractNamespace(clusterKey string) *msaddr.Namespace {
	namespacedName := v141ExtractNamespacedName(clusterKey)
	if dotIdx := strings.Index(namespacedName, "."); dotIdx != -1 {
		namespace := namespacedName[dotIdx+1:]
		return &msaddr.Namespace{Namespace: namespace}
	}
	return &msaddr.Namespace{}
}

func v141ExtractNamespacedName(clusterKey string) string {
	keyParts := strings.Split(clusterKey, "||")
	if len(keyParts) > 1 {
		return keyParts[1]
	}
	return clusterKey
}
