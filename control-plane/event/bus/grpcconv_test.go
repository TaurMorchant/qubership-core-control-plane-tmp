package bus

import (
	"bytes"
	"encoding/gob"
	"github.com/hashicorp/go-memdb"
	uuid3 "github.com/hashicorp/go-uuid"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	nodeGroup = "private-gateway-service"
)

func TestTLSConfigSerializable(t *testing.T) {
	storage := ram.NewStorage()
	tls := []*domain.TlsConfig{
		{
			Id:   1,
			Name: "A",
		},
		{
			Id:   2,
			Name: "B",
		},
	}
	tx := storage.WriteTx()
	err := storage.Save(tx, domain.TlsConfigTable, tls)
	if err != nil {
		t.Fatal(err)
	}
	tx.Commit()

	changeEvent := events.NewChangeEventByNodeGroup(nodeGroup, tx.Changes())
	builtEvent, err := buildProtobufEvent(changeEvent)
	assert.Nil(t, err)
	assert.NotNil(t, builtEvent)
	assert.NotNil(t, builtEvent.Data)
	assert.True(t, len(builtEvent.Data.Value) > 0)

	readChangeEvent, err := readChangeEventFromProtoEvent(builtEvent)

	assert.Nil(t, err)
	assert.Equal(t, changeEvent, readChangeEvent)
}

func TestTLSConfigSerializable_InReloadEvent(t *testing.T) {
	storage := ram.NewStorage()
	tls := []*domain.TlsConfig{
		{
			Id:   1,
			Name: "A",
		},
		{
			Id:   2,
			Name: "B",
		},
	}
	tx := storage.WriteTx()
	err := storage.Save(tx, domain.TlsConfigTable, tls)
	if err != nil {
		t.Fatal(err)
	}
	tx.Commit()

	changeEvent := events.NewReloadEvent(tx.Changes())
	builtEvent, err := buildProtobufEvent(changeEvent)
	assert.Nil(t, err)
	assert.NotNil(t, builtEvent)
	assert.NotNil(t, builtEvent.Data)
	assert.True(t, len(builtEvent.Data.Value) > 0)

	readChangeEvent, err := readReloadEventFromProtoEvent(builtEvent)

	assert.Nil(t, err)
	assert.Equal(t, changeEvent, readChangeEvent)
}

func TestReadProtobufEvent(t *testing.T) {
	changeEvent := &events.ReloadEvent{
		Changes: map[string][]memdb.Change{
			"changes-1": {memdb.Change{Table: "table-1"}},
		},
	}
	event, err := buildProtobufEvent(changeEvent)
	assert.Nil(t, err)

	afterRead, err := readProtobufEvent(event)
	assert.Nil(t, err)
	assert.NotNil(t, afterRead)
	assert.Equal(t, changeEvent, afterRead)
}

func TestRouteSerializable(t *testing.T) {
	storage := ram.NewStorage()
	uuid, _ := uuid3.GenerateUUID()
	routes := []*domain.Route{
		{
			Id:                1,
			Uuid:              uuid,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			Prefix:            "/api/v1",
			ClusterName:       "cluster",
			DeploymentVersion: "v1",
		},
	}
	retry := &domain.RetryPolicy{
		Id:    1,
		Route: routes[0],
	}
	routes[0].RetryPolicy = retry

	tx := storage.WriteTx()
	err := storage.Save(tx, domain.RouteTable, routes)
	if err != nil {
		t.Fatal(err)
	}
	tx.Commit()

	changeEvent := events.NewChangeEventByNodeGroup(nodeGroup, tx.Changes())
	builtEvent, err := buildProtobufEvent(changeEvent)
	assert.Nil(t, err)
	assert.NotNil(t, builtEvent)
	assert.NotNil(t, builtEvent.Data)
	assert.True(t, len(builtEvent.Data.Value) > 0)

	readChangeEvent, err := readChangeEventFromProtoEvent(builtEvent)

	assert.Nil(t, err)
	assert.Equal(t, changeEvent, readChangeEvent)
}

func Test_EncodingAndDecodingZeroValueOfCookieTTLInHashPolicy(t *testing.T) {
	var hashPolicy domain.HashPolicy
	hashPolicy.CookieName = "name"
	expectedValue := domain.NewNullInt(int64(0))
	hashPolicy.CookieTTL = expectedValue

	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(hashPolicy)
	assert.Nil(t, err)

	bytesBuffer := bytes.NewBuffer(b.Bytes())
	decoder := gob.NewDecoder(bytesBuffer)
	restoredHashPolicy := &domain.HashPolicy{}
	err = decoder.Decode(restoredHashPolicy)
	assert.Nil(t, err)

	assert.Equal(t, "name", restoredHashPolicy.CookieName)
	assert.NotNil(t, restoredHashPolicy.CookieTTL)
	assert.True(t, hashPolicy.Equals(restoredHashPolicy))
}
