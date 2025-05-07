package routeconfig

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouteGroup(t *testing.T) {
	deploymentVersions := []*domain.DeploymentVersion{
		{
			Version: "v1",
			Stage:   domain.ArchivedStage,
		},
		{
			Version: "v2",
			Stage:   domain.LegacyStage,
		},
		{
			Version: "v3",
			Stage:   domain.ActiveStage,
		},
		{
			Version: "v4",
			Stage:   domain.CandidateStage,
		},
	}

	routeGroup := NewRouteGroup(deploymentVersions)

	isMultiVersioned := routeGroup.IsMultiVersioned()
	assert.False(t, isMultiVersioned)

	archivedRoute := &domain.Route{
		DeploymentVersionVal: deploymentVersions[0],
	}
	routeGroup.AddRoute(archivedRoute)

	isMultiVersioned = routeGroup.IsMultiVersioned()
	assert.False(t, isMultiVersioned)

	hasActiveVersion := routeGroup.HasActiveVersion()
	assert.False(t, hasActiveVersion)

	legacyRoute := &domain.Route{
		DeploymentVersionVal: deploymentVersions[1],
	}
	routeGroup.AddRoute(legacyRoute)

	isMultiVersioned = routeGroup.IsMultiVersioned()
	assert.True(t, isMultiVersioned)

	hasActiveVersion = routeGroup.HasActiveVersion()
	assert.False(t, hasActiveVersion)

	initActiveRoute := &domain.Route{
		DeploymentVersionVal: deploymentVersions[2],
	}
	routeGroup.AddRoute(initActiveRoute)

	hasActiveVersion = routeGroup.HasActiveVersion()
	assert.True(t, hasActiveVersion)

	activeRoute := routeGroup.GetActiveVersionRoute()
	assert.Equal(t, initActiveRoute, activeRoute)

	noActiveRoutes := routeGroup.GetNoActiveVersionRoutes()
	assert.Equal(t, []*domain.Route{archivedRoute, legacyRoute}, noActiveRoutes)
}
