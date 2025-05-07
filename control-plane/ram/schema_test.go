package ram

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScheme(t *testing.T) {
	if err := schema.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemeEnvoyConfigVersion(t *testing.T) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		t.Fatal(err)
	}
	envConfVersions := []*domain.EnvoyConfigVersion{
		{"NodeGroupA", "Listener", 4654164674186145},
		{"NodeGroupA", "Cluster", 4651461456143546},
		{"NodeGroupB", "Listener", 231465798465498},
	}
	tx := db.Txn(true)
	for _, envConfVersion := range envConfVersions {
		if err := tx.Insert(domain.EnvoyConfigVersionTable, envConfVersion); err != nil {
			t.Fatal(err)
		}
	}
	tx.Commit()

	tx = db.Txn(false)
	defer tx.Abort()
	for _, envConfVersion := range envConfVersions {
		if actual, err := tx.First(domain.EnvoyConfigVersionTable, "id", envConfVersion.NodeGroup, envConfVersion.EntityType); err == nil {
			assert.Equal(t, envConfVersion, actual)
		} else {
			t.Fatal(err)
		}
	}
}

func TestSchemeNodeGroupIndexes(t *testing.T) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		t.Fatal(err)
	}
	nodeGroups := []*domain.NodeGroup{
		domain.NewNodeGroup("NodeGroupA"),
		domain.NewNodeGroup("NodeGroupB"),
		domain.NewNodeGroup("NodeGroupC"),
		domain.NewNodeGroup("NodeGroupD"),
	}
	tx := db.Txn(true)

	for _, nodeGroup := range nodeGroups {
		if err := tx.Insert(domain.NodeGroupTable, nodeGroup); err != nil {
			t.Fatal(err)
		}
	}
	tx.Commit()

	tx = db.Txn(false)
	for _, nodeGroup := range nodeGroups {
		if actual, err := tx.First(domain.NodeGroupTable, "id", nodeGroup.Name); err == nil {
			assert.Equal(t, nodeGroup, actual)
		} else {
			t.Fatal(err)
		}
	}
	tx.Abort()
}

func TestSchemeClusterNodeGroupIndexes(t *testing.T) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		t.Fatal(err)
	}
	clusterNodeGroups := []*domain.ClustersNodeGroup{
		domain.NewClusterNodeGroups(1, "NodeGroupA"),
		domain.NewClusterNodeGroups(2, "NodeGroupA"),
		domain.NewClusterNodeGroups(3, "NodeGroupA"),
		domain.NewClusterNodeGroups(3, "NodeGroupB"),
	}
	tx := db.Txn(true)

	for _, clusterNodeGroup := range clusterNodeGroups {
		if err := tx.Insert(domain.ClusterNodeGroupTable, clusterNodeGroup); err != nil {
			t.Fatal(err)
		}
	}
	tx.Commit()

	tx = db.Txn(false)
	for _, clusterNodeGroup := range clusterNodeGroups {
		if actual, err := tx.First(domain.ClusterNodeGroupTable, "id", clusterNodeGroup.ClustersId, clusterNodeGroup.NodegroupsName); err == nil {
			assert.Equal(t, clusterNodeGroup, actual)
		} else {
			t.Fatal(err)
		}
	}

	AssertFindResultSize(t, 2, tx, domain.ClusterNodeGroupTable, "clustersId", int32(3))
	AssertFindResultSize(t, 3, tx, domain.ClusterNodeGroupTable, "nodegroupsName", "NodeGroupA")

	tx.Abort()
}

func AssertFindResultSize(t *testing.T, expectedSize int, tx *memdb.Txn, table string, index string, args ...interface{}) {
	if itr, err := tx.Get(table, index, args...); err == nil {
		entities := make([]interface{}, 0)
		for entity := itr.Next(); entity != nil; entity = itr.Next() {
			entities = append(entities, entity)
		}
		assert.Equal(t, expectedSize, len(entities))
	} else {
		t.Fatal(err)
	}
}
