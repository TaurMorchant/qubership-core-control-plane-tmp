package routeconfig

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleRouteTransformationRule(t *testing.T) {
	simpleRouteTransformationRule := NewSimpleRouteTransformationRule()

	applicableRouteGroup := &RouteGroup{
		routesVersions: []*domain.DeploymentVersion{
			{},
		},
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Stage: domain.ActiveStage,
				},
				ClusterName: "ClusterName1",
			},
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Version: "v2",
					Stage:   domain.LegacyStage,
				},
				DirectResponseCode: uint32(0),
				ClusterName:        "ClusterName2",
			},
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Version: "v3",
					Stage:   domain.LegacyStage,
				},
				DirectResponseCode: uint32(1),
			},
		},
	}

	isApplicable := simpleRouteTransformationRule.IsApplicable(applicableRouteGroup)
	assert.True(t, isApplicable)

	notApplicableRouteGroup := &RouteGroup{
		routesVersions: []*domain.DeploymentVersion{
			{},
			{},
		},
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Stage: domain.LegacyStage,
				},
			},
		},
	}
	isApplicable = simpleRouteTransformationRule.IsApplicable(notApplicableRouteGroup)
	assert.False(t, isApplicable)

	routes := simpleRouteTransformationRule.Apply(applicableRouteGroup)
	assert.Equal(t, len(applicableRouteGroup.routes), len(routes))
	assert.Equal(t, applicableRouteGroup.routes[0], routes[0])
	assert.Equal(t, applicableRouteGroup.routes[1], routes[1])
	assert.Equal(t, applicableRouteGroup.routes[2], routes[2])
}

func TestNoActiveRouteTransformer(t *testing.T) {
	noActiveRouteTransformer := NewNoActiveRouteTransformer()

	applicableRouteGroup := &RouteGroup{
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Version: "v2",
					Stage:   domain.LegacyStage,
				},
				DirectResponseCode: uint32(0),
				ClusterName:        "ClusterName2",
			},
		},
	}
	headerMatchers := &domain.HeaderMatcher{Name: "x-version", ExactMatch: applicableRouteGroup.routes[0].DeploymentVersionVal.Version}

	isApplicable := noActiveRouteTransformer.IsApplicable(applicableRouteGroup)
	assert.True(t, isApplicable)

	notApplicableRouteGroup := &RouteGroup{
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Stage: domain.ActiveStage,
				},
			},
		},
	}

	isApplicable = noActiveRouteTransformer.IsApplicable(notApplicableRouteGroup)
	assert.False(t, isApplicable)

	routes := noActiveRouteTransformer.Apply(applicableRouteGroup)
	assert.Equal(t, 1, len(routes))
	assert.Equal(t, headerMatchers, routes[0].HeaderMatchers[0])
}

func TestGenericVersionedRouteTransformRule(t *testing.T) {
	genericVersionedRouteTransformRule := NewGenericVersionedRouteTransformRule()

	applicableRouteGroup := &RouteGroup{
		routesVersions: []*domain.DeploymentVersion{
			{},
			{},
		},
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Stage: domain.ActiveStage,
				},
				ClusterName: "ClusterName1",
			},
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Version: "v2",
					Stage:   domain.LegacyStage,
				},
				DirectResponseCode: uint32(0),
				ClusterName:        "ClusterName2",
			},
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Version: "v3",
					Stage:   domain.LegacyStage,
				},
				DirectResponseCode: uint32(1),
			},
		},
	}

	isApplicable := genericVersionedRouteTransformRule.IsApplicable(applicableRouteGroup)
	assert.True(t, isApplicable)

	notApplicableRouteGroup := &RouteGroup{
		routesVersions: []*domain.DeploymentVersion{
			{},
			{},
		},
		routes: []*domain.Route{
			{
				DeploymentVersionVal: &domain.DeploymentVersion{
					Stage: domain.LegacyStage,
				},
			},
		},
	}
	isApplicable = genericVersionedRouteTransformRule.IsApplicable(notApplicableRouteGroup)
	assert.False(t, isApplicable)

	headerMatchers1 := &domain.HeaderMatcher{Name: "x-version", ExactMatch: applicableRouteGroup.routes[1].DeploymentVersionVal.Version}
	headerMatchers2 := &domain.HeaderMatcher{Name: "x-version", ExactMatch: applicableRouteGroup.routes[2].DeploymentVersionVal.Version}

	routes := genericVersionedRouteTransformRule.Apply(applicableRouteGroup)
	assert.Equal(t, len(applicableRouteGroup.routes), len(routes))
	assert.Equal(t, applicableRouteGroup.routes[0], routes[0])
	assert.Equal(t, headerMatchers1, routes[1].HeaderMatchers[0])
	assert.Equal(t, headerMatchers2, routes[2].HeaderMatchers[0])
}

func TestRouteTransformer_Transform(t *testing.T) {
	routes := []*domain.Route{
		{
			Id: int32(1),
		},
	}
	mockTransformationRule := &TransformationRuleStub{
		routes:       routes,
		isApplicable: true,
	}
	routeTransformer := NewRouteTransformer(mockTransformationRule)

	routeGroups := []*RouteGroup{
		{
			allVersions: []*domain.DeploymentVersion{
				{
					Version: "v1",
				},
			},
		},
		{
			allVersions: []*domain.DeploymentVersion{
				{
					Version: "skip",
				},
			},
		},
	}

	resultRoutes := routeTransformer.Transform(routeGroups)
	assert.Equal(t, 1, len(resultRoutes))
	assert.Equal(t, routes, resultRoutes)
}

type TransformationRuleStub struct {
	routes       []*domain.Route
	isApplicable bool
}

func (t *TransformationRuleStub) IsApplicable(routeGroup *RouteGroup) bool {
	if routeGroup.allVersions[0].Version == "skip" {
		return false
	}
	return true
}

func (t *TransformationRuleStub) Apply(routeGroup *RouteGroup) []*domain.Route {
	return t.routes
}
