package itmod

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

var db *memdb.MemDB

func TestMemStorageOneIndexMemoryCorrelation(t *testing.T) {
	runtime.MemProfileRate = 0
	db, _ = memdb.NewMemDB(oneIndexSchema)
	runtime.MemProfileRate = 1
	route := domain.Route{
		Id:                       5,
		Uuid:                     uuid.New().String(),
		VirtualHostId:            50,
		RouteKey:                 "4564s5d4s6d4s65d4s65d4s65d4s6d4s56d4sd7s89d7s65d4wad1465df4w6a5d4wad",
		Prefix:                   "/api/v1/route/to/eden",
		ClusterName:              "limb||limb||80",
		HostAutoRewrite:          domain.NullBool{},
		PrefixRewrite:            "/api/v1/route/to/valhalla",
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	}
	tx := db.Txn(true)
	_ = tx.Insert(domain.RouteTable, route)
	tx.Commit()
}

func TestMemStorageMultiIndexMemoryCorrelation(t *testing.T) {
	runtime.MemProfileRate = 0
	db, _ = memdb.NewMemDB(multiIndexSchema)
	runtime.MemProfileRate = 1
	route := domain.Route{
		Id:                       5,
		Uuid:                     uuid.New().String(),
		VirtualHostId:            50,
		RouteKey:                 "4564s5d4s6d4s65d4s65d4s65d4s6d4s56d4sd7s89d7s65d4wad1465df4w6a5d4wad",
		Prefix:                   "/api/v1/route/to/eden",
		ClusterName:              "limb||limb||80",
		HostAutoRewrite:          domain.NullBool{},
		PrefixRewrite:            "/api/v1/route/to/valhalla",
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	}
	tx := db.Txn(true)
	_ = tx.Insert(domain.RouteTable, route)
	tx.Commit()
	tx = db.Txn(false)
	newRoute, _ := tx.First(domain.RouteTable, "id", route.Id)
	assert.Equal(t, route, newRoute)
}

// To catch memory profile run follow command
// go test -v control-plane/itmod -run ^\QTestMemStorageCloneObject\E$ -memprofile mem.out
func TestMemStorageCloneObject(t *testing.T) {
	runtime.MemProfileRate = 0
	db, _ = memdb.NewMemDB(oneIndexSchema)
	runtime.MemProfileRate = 1
	route := &domain.Route{
		Id:                       5,
		Uuid:                     uuid.New().String(),
		VirtualHostId:            50,
		RouteKey:                 "4564s5d4s6d4s65d4s65d4s65d4s6d4s56d4sd7s89d7s65d4wad1465df4w6a5d4wad",
		Prefix:                   "/api/v1/route/to/eden",
		ClusterName:              "limb||limb||80",
		HostAutoRewrite:          domain.NullBool{},
		PrefixRewrite:            "/api/v1/route/to/valhalla",
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	}
	tx := db.Txn(true)
	_ = tx.Insert(domain.RouteTable, route)
	tx.Commit()
	tx = db.Txn(false)
	foundEntity, _ := tx.First(domain.RouteTable, "id", route.Id)
	newRoute := foundEntity.(*domain.Route)
	assert.Equal(t, route, newRoute)
	clone := newRoute.Clone()
	clone.DeploymentVersion = "v2"
	tx = db.Txn(true)
	_ = tx.Insert(domain.RouteTable, clone)
	foundEntity2, _ := tx.First(domain.RouteTable, "id", route.Id)
	newRoute2 := foundEntity2.(*domain.Route)
	assert.NotEqual(t, route, newRoute2)
	runtime.GC()
}

var oneIndexSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		domain.RouteTable: {
			Name: domain.RouteTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.IntFieldIndex{Field: "Id"},
				},
			},
		},
	},
}
var multiIndexSchema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		domain.RouteTable: {
			Name: domain.RouteTable,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:    "id",
					Unique:  true,
					Indexer: &memdb.IntFieldIndex{Field: "Id"},
				},
				"uuid": {
					Name:    "uuid",
					Unique:  true,
					Indexer: &memdb.StringFieldIndex{Field: "Uuid"},
				},
				"virtualHostId": {
					Name:    "virtualHostId",
					Unique:  false,
					Indexer: &memdb.IntFieldIndex{Field: "VirtualHostId"},
				},
				"vHostIdAndRouteKey": {
					Name:   "vHostIdAndRouteKey",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.IntFieldIndex{Field: "VirtualHostId"},
							&memdb.StringFieldIndex{Field: "RouteKey"},
						},
						AllowMissing: false,
					},
				},
				"dVersion": {
					Name:    "dVersion",
					Unique:  false,
					Indexer: &memdb.StringFieldIndex{Field: "DeploymentVersion"},
				},
				"clusterName": {
					Name:         "clusterName",
					Unique:       false,
					AllowMissing: true,
					Indexer:      &memdb.StringFieldIndex{Field: "ClusterName"},
				},
				"clusterNameAndDVersion": {
					Name:   "clusterNameAndDVersion",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ClusterName"},
							&memdb.StringFieldIndex{Field: "DeploymentVersion"},
						},
						AllowMissing: true,
					},
				},
				"autoGenAndDVersion": {
					Name:   "autoGenAndDVersion",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.BoolFieldIndex{Field: "Autogenerated"},
							&memdb.StringFieldIndex{Field: "DeploymentVersion"},
						},
						AllowMissing: false,
					},
				},
				"deploymentVersionAndRouteKey": {
					Name:   "deploymentVersionAndRouteKey",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "RouteKey"},
							&memdb.StringFieldIndex{Field: "DeploymentVersion"},
						},
						AllowMissing: false,
					},
				},
				"vHostIdAndDeploymentVersion": {
					Name:   "vHostIdAndDeploymentVersion",
					Unique: false,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.IntFieldIndex{Field: "VirtualHostId"},
							&memdb.StringFieldIndex{Field: "DeploymentVersion"},
						},
						AllowMissing: false,
					},
				},
			},
		},
	},
}
