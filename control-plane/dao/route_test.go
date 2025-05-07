package dao

import (
	"github.com/google/uuid"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindRoutesByVirtualHostIdAndDeploymentVersion(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	firstVirtualHostId := int32(1)
	secondVirtualHostId := int32(2)

	firstDeploymentVersion := "v1"
	secondDeploymentVersion := "v2"
	nonExistingDeploymentVersion := "v3"

	firstRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     firstVirtualHostId,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: firstDeploymentVersion,
		ClusterName:       "clusterName1",
	}

	secondRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     secondVirtualHostId,
		RouteKey:          "/api/v2/test",
		DeploymentVersion: secondDeploymentVersion,
		ClusterName:       "clusterName2",
	}

	_, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveRoute(firstRoute))
		assert.Nil(t, dao.SaveRoute(secondRoute))
		return nil
	})

	foundRoutes, err := testable.FindRoutesByVirtualHostIdAndDeploymentVersion(firstVirtualHostId, firstDeploymentVersion)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByVirtualHostIdAndDeploymentVersion(firstVirtualHostId, nonExistingDeploymentVersion)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))
}

func TestFindRoutesByDeploymentVersions(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	dVersions := []*domain.DeploymentVersion{
		{Version: "v1"},
		{Version: "v2"},
		{Version: "v3"},
	}

	firstRoute := &domain.Route{
		Uuid:                 uuid.New().String(),
		VirtualHostId:        1,
		RouteKey:             "/api/v1/test",
		DeploymentVersion:    "v1",
		DeploymentVersionVal: dVersions[0],
		ClusterName:          "clusterName1",
	}

	secondRoute := &domain.Route{
		Uuid:                 uuid.New().String(),
		VirtualHostId:        2,
		RouteKey:             "/api/v2/test",
		DeploymentVersion:    "v2",
		DeploymentVersionVal: dVersions[1],
		ClusterName:          "clusterName2",
	}

	_, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveRoute(firstRoute))
		assert.Nil(t, dao.SaveRoute(secondRoute))
		return nil
	})

	foundRoutes, err := testable.FindRoutesByDeploymentVersions(dVersions[0])
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByDeploymentVersions(dVersions[0], dVersions[1], dVersions[2])
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])
	assert.Equal(t, secondRoute, foundRoutes[1])

	foundRoutes, err = testable.FindRoutesByDeploymentVersions(dVersions[2])
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))
}

func TestFindAndDeleteRouteByIdAndUUID(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	firstUUID := uuid.New().String()
	secondUUID := uuid.New().String()

	firstRoute := &domain.Route{
		Uuid:              firstUUID,
		VirtualHostId:     1,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: "v1",
		ClusterName:       "clusterName1",
	}

	secondRoute := &domain.Route{
		Uuid:              secondUUID,
		Id:                2,
		VirtualHostId:     2,
		RouteKey:          "/api/v2/test",
		DeploymentVersion: "v2",
		ClusterName:       "clusterName2",
	}

	_, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveRoute(firstRoute))
		assert.Nil(t, dao.SaveRoute(secondRoute))
		return nil
	})

	foundFirstRoute, err := testable.FindRouteById(1)
	assert.Equal(t, firstRoute, foundFirstRoute)

	foundSecondRoute, err := testable.FindRouteByUuid(secondUUID)
	assert.Equal(t, secondRoute, foundSecondRoute)

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteRouteById(1))
		return nil
	})

	foundFirstRoute, err = testable.FindRouteById(1)
	assert.Nil(t, err)
	assert.Nil(t, foundFirstRoute)

	foundSecondRoute, err = testable.FindRouteByUuid(secondUUID)
	assert.Nil(t, err)
	assert.Equal(t, secondRoute, foundSecondRoute)

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteRouteByUUID(secondUUID))
		return nil
	})

	foundSecondRoute, err = testable.FindRouteByUuid(secondUUID)
	assert.Nil(t, err)
	assert.Nil(t, foundSecondRoute)
}

func TestFindRoutesByNamespaceHeaderIsNot(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	firstRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: "v1",
		ClusterName:       "clusterName1",
	}

	secondRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1",
		DeploymentVersion: "v1",
		ClusterName:       "clusterName2",
		Autogenerated:     true,
	}

	routes := []*domain.Route{firstRoute, secondRoute}

	saveRoutesUsingDao(t, testable, routes)

	expectedHeaderMatcher := &domain.HeaderMatcher{
		Id:   1,
		Name: "namespace",
		RangeMatch: domain.RangeMatch{
			Start: domain.NewNullInt(0),
			End:   domain.NewNullInt(0),
		},
		RouteId: 1,
	}

	changes, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveHeaderMatcher(expectedHeaderMatcher))
		return nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, changes)

	foundRoutes, err := testable.FindRoutesByNamespaceHeaderIsNot(firstRoute.ClusterName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])
}

func TestFindRoutesByClusterName(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	nonExistingClusterName := "clusterName3"

	firstRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: "v1",
		ClusterName:       "clusterName1",
	}

	secondRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1",
		DeploymentVersion: "v1",
		ClusterName:       "clusterName2",
		Autogenerated:     true,
	}

	routes := []*domain.Route{firstRoute, secondRoute}

	saveRoutesUsingDao(t, testable, routes)

	foundRoutes, err := testable.FindRoutesByClusterName(firstRoute.ClusterName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByClusterName(secondRoute.ClusterName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, secondRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByClusterName(nonExistingClusterName)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))
}

func TestFindRoutesByDeploymentVersionAndRouteKey(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	deploymentVersion := "v1"
	nonExistingDeploymentVersion := "v2"

	firstRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: deploymentVersion,
		ClusterName:       "clusterName",
	}

	secondRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1",
		DeploymentVersion: deploymentVersion,
		ClusterName:       "clusterName",
		Autogenerated:     true,
	}

	routes := []*domain.Route{firstRoute, secondRoute}

	saveRoutesUsingDao(t, testable, routes)

	foundRoutes, err := testable.FindRoutesByDeploymentVersionAndRouteKey(deploymentVersion, firstRoute.RouteKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByDeploymentVersionAndRouteKey(deploymentVersion, secondRoute.RouteKey)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, secondRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByDeploymentVersionAndRouteKey(nonExistingDeploymentVersion, firstRoute.RouteKey)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))
}

func TestFindRoutesByAutoGeneratedAndDeploymentVersion(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	deploymentVersion := "v1"
	nonExistingDeploymentVersion := "v2"

	firstRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1/test",
		DeploymentVersion: deploymentVersion,
		ClusterName:       "clusterName",
	}

	secondRoute := &domain.Route{
		Uuid:              uuid.New().String(),
		VirtualHostId:     1,
		RouteKey:          "/api/v1",
		DeploymentVersion: deploymentVersion,
		ClusterName:       "clusterName",
		Autogenerated:     true,
	}

	routes := []*domain.Route{firstRoute, secondRoute}

	saveRoutesUsingDao(t, testable, routes)

	foundRoutes, err := testable.FindRoutesByAutoGeneratedAndDeploymentVersion(false, deploymentVersion)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, firstRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByAutoGeneratedAndDeploymentVersion(true, deploymentVersion)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, secondRoute, foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByAutoGeneratedAndDeploymentVersion(false, nonExistingDeploymentVersion)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))
}

func TestSaveRoutes(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()

	routes := []*domain.Route{
		{
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v1",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid2,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			DeploymentVersion: "v2",
			ClusterName:       "clusterName",
			Autogenerated:     true,
		},
		{
			Uuid:              uuid3,
			VirtualHostId:     2,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v3",
			ClusterName:       "clusterName2",
		},
	}
	foundRoutes, err := testable.FindAllRoutes()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(foundRoutes))

	saveRoutesUsingDao(t, testable, routes)

	foundRoutes, err = testable.FindAllRoutes()
	assert.Nil(t, err)
	assert.Equal(t, len(routes), len(foundRoutes))
}

func TestDeleteHeaderMatcher(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	firstHeaderMatcher := &domain.HeaderMatcher{
		Id:   1,
		Name: "TestHeader1",
		RangeMatch: domain.RangeMatch{
			Start: domain.NewNullInt(0),
			End:   domain.NewNullInt(0),
		},
		RouteId: 1,
	}

	secondHeaderMatcher := &domain.HeaderMatcher{
		Id:   2,
		Name: "TestHeader2",
		RangeMatch: domain.RangeMatch{
			Start: domain.NewNullInt(0),
			End:   domain.NewNullInt(0),
		},
		RouteId: 1,
	}

	changes, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveHeaderMatcher(firstHeaderMatcher))
		assert.Nil(t, dao.SaveHeaderMatcher(secondHeaderMatcher))
		return nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, changes)

	actualHeaderMatcher, err := testable.FindHeaderMatcherByRouteId(1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualHeaderMatcher))

	_, err = testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.DeleteHeaderMatcher(firstHeaderMatcher))
		return nil
	})
	assert.Nil(t, err)

	actualHeaderMatcher, err = testable.FindHeaderMatcherByRouteId(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualHeaderMatcher))

	_, err = testable.WithWTx(func(dao Repository) error {
		deletedCount, _ := dao.DeleteHeaderMatchersByRouteId(secondHeaderMatcher.RouteId)
		assert.Equal(t, 1, deletedCount)
		return nil
	})
	assert.Nil(t, err)

	actualHeaderMatcher, err = testable.FindHeaderMatcherByRouteId(1)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(actualHeaderMatcher))
}

func TestInMemDao_FindRoutes(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()

	routes := []*domain.Route{
		{
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v1",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid2,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			DeploymentVersion: "v2",
			ClusterName:       "clusterName",
			Autogenerated:     true,
		},
		{
			Uuid:              uuid3,
			VirtualHostId:     2,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v3",
			ClusterName:       "clusterName2",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, route := range routes {
			assert.Nil(t, dao.SaveRoute(route))
		}
		return nil
	})
	assert.Nil(t, err)

	foundRoutes, err := testable.FindRoutesByVirtualHostIdAndRouteKey(int32(1), "/api/v1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, routes[1], foundRoutes[0])

	foundRoutes, err = testable.FindRoutesByDeploymentVersionIn("v2", "v3")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundRoutes))

	foundRoutes, err = testable.FindRoutesByClusterNameAndDeploymentVersion("clusterName", "v2")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, foundRoutes[0], routes[1])

	foundRoutes, err = testable.FindRoutesByVirtualHostId(2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))
	assert.Equal(t, routes[2], foundRoutes[0])
}

func TestInMemDao_DeleteRoutesByAutoGeneratedAndDeploymentVersion(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()

	routes := []*domain.Route{
		{
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v1",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid2,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			DeploymentVersion: "v2",
			ClusterName:       "clusterName",
			Autogenerated:     true,
		},
		{
			Uuid:              uuid3,
			VirtualHostId:     2,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v3",
			ClusterName:       "clusterName2",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, route := range routes {
			assert.Nil(t, dao.SaveRoute(route))
		}
		return nil
	})
	assert.Nil(t, err)

	deletedRoutes, err := testable.DeleteRoutesByAutoGeneratedAndDeploymentVersion(true, "v2")
	assert.Nil(t, err)
	assert.Equal(t, 1, deletedRoutes)
}

func TestInMemDao_FindHeaderMatcherByRouteId(t *testing.T) {
	// TODO implement me
}

func TestInMemDao_FindRoutesByDeploymentVersionStageIn(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	uuid1 := uuid.New().String()
	uuid2 := uuid.New().String()
	uuid3 := uuid.New().String()

	routes := []*domain.Route{
		{
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v1",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid2,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			DeploymentVersion: "v2",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid3,
			VirtualHostId:     2,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v3",
			ClusterName:       "clusterName2",
		},
	}
	dVersions := []*domain.DeploymentVersion{
		{
			Version: "v1",
			Stage:   "LEGACY",
		},
		{
			Version: "v2",
			Stage:   "ACTIVE",
		},
		{
			Version: "v3",
			Stage:   "CANDIDATE",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, route := range routes {
			assert.Nil(t, dao.SaveRoute(route))
		}
		for _, dVersion := range dVersions {
			assert.Nil(t, dao.SaveDeploymentVersion(dVersion))
		}
		return nil
	})
	assert.Nil(t, err)

	actualRoutes, err := testable.FindRoutesByDeploymentVersionStageIn("LEGACY", "CANDIDATE")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualRoutes))
}

func TestInMemRepo_FindRoutesByUUIDPrefix(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	uuid1 := "11111111-1e8b-4470-be04-4f6535d6d500"
	uuid2 := "11112222-1e8b-4470-be04-4f6535d6d500"

	routes := []*domain.Route{
		{
			Uuid:              uuid1,
			VirtualHostId:     1,
			RouteKey:          "/api/v1/test",
			DeploymentVersion: "v1",
			ClusterName:       "clusterName",
		},
		{
			Uuid:              uuid2,
			VirtualHostId:     1,
			RouteKey:          "/api/v1",
			DeploymentVersion: "v2",
			ClusterName:       "clusterName",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, route := range routes {
			assert.Nil(t, dao.SaveRoute(route))
		}
		return nil
	})
	assert.Nil(t, err)

	foundRoutes, err := testable.FindRoutesByUUIDPrefix("11111111")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(foundRoutes))

	foundRoutes, err = testable.FindRoutesByUUIDPrefix("1111")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(foundRoutes))
}

func TestInMemRepo_SaveAndGetHeaderMatcherWithRangeMatch(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	expectedHeaderMatcher := &domain.HeaderMatcher{
		Id:   1,
		Name: "TestHeader",
		RangeMatch: domain.RangeMatch{
			Start: domain.NullInt{},
			End:   domain.NullInt{},
		},
		RouteId: 1,
	}

	changes, err := testable.WithWTx(func(dao Repository) error {
		assert.Nil(t, dao.SaveHeaderMatcher(expectedHeaderMatcher))
		return nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, changes)
	actualHeaderMatcher, err := testable.FindHeaderMatcherByRouteId(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualHeaderMatcher))
	assert.False(t, actualHeaderMatcher[0].RangeMatch.Start.Valid)
	assert.False(t, actualHeaderMatcher[0].RangeMatch.End.Valid)
	assert.Equal(t, int64(0), actualHeaderMatcher[0].RangeMatch.Start.Int64)
	assert.Equal(t, int64(0), actualHeaderMatcher[0].RangeMatch.End.Int64)
}

func saveRoutesUsingDao(t *testing.T, testable Dao, routes []*domain.Route) {
	_, err := testable.WithWTx(func(repo Repository) error {
		for _, route := range routes {
			assert.Nil(t, repo.SaveRoute(route))
		}
		return nil
	})
	assert.Nil(t, err)
}
