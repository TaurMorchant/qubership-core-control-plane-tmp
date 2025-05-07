package version

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersionStateAppropriatelySorted(t *testing.T) {
	expectedSortResult := []*domain.DeploymentVersion{
		{
			Version: "v1",
		},
		{
			Version: "v2",
		},

		{
			Version: "v10",
		},
		{
			Version: "v11",
		},
		{
			Version: "v12",
		},
	}
	input := []*domain.DeploymentVersion{
		{Version: "v12"}, {Version: "v11"}, {Version: "v10"}, {Version: "v2"}, {Version: "v1"},
	}
	state := NewVersionState(input)
	for idx, val := range state.versions {
		assert.Equal(t, expectedSortResult[idx].Version, val.Version)
	}
}
