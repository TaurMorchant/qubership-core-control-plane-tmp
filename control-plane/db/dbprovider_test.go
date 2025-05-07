package db

import (
	"context"
	_ "github.com/netcracker/qubership-core-control-plane/serviceregistrar"
	dbaasbase "github.com/netcracker/qubership-core-lib-go-dbaas-base-client/v3"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Test error cases
// Required to suppress coverage requirement

func TestNewDBProvider(t *testing.T) {
	os.Setenv("microservice.namespace", "test")
	configloader.Init(configloader.EnvPropertySource())
	dbaasPool := dbaasbase.NewDbaaSPool()
	provider, err := NewDBProvider(dbaasPool)
	assert.Nil(t, err)
	assert.NotNil(t, provider)
}

func TestGetBunDB(t *testing.T) {
	os.Setenv("microservice.namespace", "test")
	configloader.Init(configloader.EnvPropertySource())
	dbaasPool := dbaasbase.NewDbaaSPool()
	provider, err := NewDBProvider(dbaasPool)
	assert.Nil(t, err)

	_, err = provider.GetDB(context.Background())
	assert.NotNil(t, err)
}

func TestGetConn(t *testing.T) {
	os.Setenv("microservice.namespace", "test")
	configloader.Init(configloader.EnvPropertySource())
	dbaasPool := dbaasbase.NewDbaaSPool()
	provider, err := NewDBProvider(dbaasPool)
	assert.Nil(t, err)

	_, err = provider.GetConn(context.Background())
	assert.NotNil(t, err)
}
