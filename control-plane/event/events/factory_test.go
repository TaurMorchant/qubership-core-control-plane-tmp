package events

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestChangesToMap(t *testing.T) {
	changes := []memdb.Change{
		{
			Table: "deployment_versions",
			Before: domain.DeploymentVersion{
				Version:     "v1",
				Stage:       "ACTIVE",
				CreatedWhen: time.Now(),
				UpdatedWhen: time.Now(),
			},
			After: domain.DeploymentVersion{
				Version:     "v1",
				Stage:       "LEGACY",
				CreatedWhen: time.Now(),
				UpdatedWhen: time.Now(),
			},
		},
		{
			Table: "deployment_versions",
			Before: domain.DeploymentVersion{
				Version:     "v2",
				Stage:       "CANDIDATE",
				CreatedWhen: time.Now(),
				UpdatedWhen: time.Now(),
			},
			After: domain.DeploymentVersion{
				Version:     "v2",
				Stage:       "ACTIVE",
				CreatedWhen: time.Now(),
				UpdatedWhen: time.Now(),
			},
		},
	}

	expectedMap := map[string][]memdb.Change{
		"deployment_versions": changes,
	}

	nodeGrName := "node-gr-name"
	result := NewChangeEventByNodeGroup(nodeGrName, changes)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMap, result.Changes)
	assert.Equal(t, nodeGrName, result.NodeGroup)

	result2 := NewChangeEvent(changes)
	assert.NotNil(t, result2)
	assert.Equal(t, expectedMap, result2.Changes)
	assert.Empty(t, result2.NodeGroup)
}
