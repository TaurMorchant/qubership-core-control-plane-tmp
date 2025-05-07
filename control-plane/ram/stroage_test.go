package ram

import (
	"github.com/hashicorp/go-uuid"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStorage(t *testing.T) {
	storage := NewStorage()
	uuid1, _ := uuid.GenerateUUID()
	uuid2, _ := uuid.GenerateUUID()
	routes := []*domain.Route{
		{
			Id:                1,
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			Prefix:            "/api/v1",
			ClusterName:       "cluster",
			DeploymentVersion: "v1",
		},
		{
			Id:                2,
			Uuid:              uuid2,
			VirtualHostId:     2,
			RouteKey:          "/api/v1",
			Prefix:            "/api/v1",
			ClusterName:       "cluster",
			DeploymentVersion: "v1",
		},
	}
	tx := storage.WriteTx()
	err := storage.Save(tx, domain.RouteTable, routes)
	if err != nil {
		t.Fatal(err)
	}
	tx.Commit()

	foundRoute, err := storage.FindFirstByIndex(storage.ReadTx(), domain.RouteTable, "vHostIdAndRouteKey", int32(1), "/api/v1")
	if err != nil {
		t.Fatal(err)
	}
	expectedRoute := routes[0]
	assert.Equal(t, expectedRoute, foundRoute)
}
