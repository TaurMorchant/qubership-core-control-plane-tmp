package bluegreen

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegistry_Filters(t *testing.T) {
	inMemDao, bgSrv := prepareTest("v2")
	v1 := domain.NewDeploymentVersion("v1", domain.LegacyStage)
	v2 := domain.NewDeploymentVersion("v2", domain.ActiveStage)
	saveDeploymentVersions(t, inMemDao, v1, v2)

	registry, ok := bgSrv.versionsRegistry.(*versionsRegistry[dto.ServicesVersionPayload])
	assert.True(t, ok)

	result, err := registry.GetAll(ctx, inMemDao)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v1",
		Stage:   domain.LegacyStage,
	}, result)
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v2",
		Stage:   domain.ActiveStage,
	}, result)

	_, err = inMemDao.WithWTx(func(repo dao.Repository) error {
		cluster1 := &domain.Cluster{Name: "test-cluster1||test-cluster1||8080"}
		cluster1Tls := &domain.Cluster{Name: "test-cluster1||test-cluster1||8443"}
		cluster2 := &domain.Cluster{Name: "test-cluster2||test-cluster2.test-ns||8080"}
		cluster2Tls := &domain.Cluster{Name: "test-cluster2||test-cluster2.test-ns||8443"}
		assert.Nil(t, repo.SaveCluster(cluster1))
		assert.Nil(t, repo.SaveCluster(cluster1Tls))
		assert.Nil(t, repo.SaveCluster(cluster2))
		assert.Nil(t, repo.SaveCluster(cluster2Tls))
		endpoint1v1 := &domain.Endpoint{
			Address:                  "test-cluster1-v1",
			Port:                     8080,
			Protocol:                 "http",
			ClusterId:                cluster1.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}
		endpoint1v1Tls := &domain.Endpoint{
			Address:                  "test-cluster1-v1",
			Port:                     8443,
			Protocol:                 "https",
			ClusterId:                cluster1.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}
		endpoint1v2 := &domain.Endpoint{
			Address:                  "test-cluster1-v2",
			Port:                     8080,
			Protocol:                 "http",
			ClusterId:                cluster1.Id,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v2",
		}
		endpoint1v2Tls := &domain.Endpoint{
			Address:                  "test-cluster1-v2",
			Port:                     8443,
			Protocol:                 "https",
			ClusterId:                cluster1.Id,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v2",
		}
		endpoint2v2 := &domain.Endpoint{
			Address:                  "test-cluster2",
			Port:                     8080,
			Protocol:                 "http",
			ClusterId:                cluster2.Id,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v1",
		}
		endpoint2v2Tls := &domain.Endpoint{
			Address:                  "test-cluster2",
			Port:                     8443,
			Protocol:                 "https",
			ClusterId:                cluster2Tls.Id,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v1",
		}
		assert.Nil(t, repo.SaveEndpoint(endpoint1v1))
		assert.Nil(t, repo.SaveEndpoint(endpoint1v1Tls))
		assert.Nil(t, repo.SaveEndpoint(endpoint1v2))
		assert.Nil(t, repo.SaveEndpoint(endpoint1v2Tls))
		assert.Nil(t, repo.SaveEndpoint(endpoint2v2))
		assert.Nil(t, repo.SaveEndpoint(endpoint2v2Tls))
		assert.Nil(t, repo.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     "test-cluster1",
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}))
		assert.Nil(t, repo.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     "test-cluster1",
			Namespace:                msaddr.LocalNamespace,
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v2",
		}))
		assert.Nil(t, repo.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     "test-cluster2",
			Namespace:                "test-ns",
			DeploymentVersion:        "v2",
			InitialDeploymentVersion: "v1",
		}))
		return nil
	})
	assert.Nil(t, err)

	// test GetAll
	result, err = registry.GetAll(ctx, inMemDao)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v1",
		Stage:   domain.LegacyStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster1", Namespace: msaddr.LocalNamespace, Endpoints: []string{"http://test-cluster1-v1:8080", "https://test-cluster1-v1:8443"}},
		},
	}, result)
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v2",
		Stage:   domain.ActiveStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster1", Namespace: msaddr.LocalNamespace, Endpoints: []string{"http://test-cluster1-v2:8080", "https://test-cluster1-v2:8443"}},
			{Cluster: "test-cluster2", Namespace: "test-ns", Endpoints: []string{"http://test-cluster2:8080", "https://test-cluster2:8443"}},
		},
	}, result)

	// test GetMicroserviceCurrentVersion
	result, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-cluster2", msaddr.Namespace{Namespace: "test-ns"}, "v1")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v2",
		Stage:   domain.ActiveStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster2", Namespace: "test-ns", Endpoints: []string{"http://test-cluster2:8080", "https://test-cluster2:8443"}},
		},
	}, result)

	// test GetVersionsForMicroservice
	result, err = registry.GetVersionsForMicroservice(ctx, inMemDao, "test-cluster1", msaddr.CurrentNamespace())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(result))
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v1",
		Stage:   domain.LegacyStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster1", Namespace: msaddr.LocalNamespace, Endpoints: []string{"http://test-cluster1-v1:8080", "https://test-cluster1-v1:8443"}},
		},
	}, result)
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v2",
		Stage:   domain.ActiveStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster1", Namespace: msaddr.LocalNamespace, Endpoints: []string{"http://test-cluster1-v2:8080", "https://test-cluster1-v2:8443"}},
		},
	}, result)

	// test GetMicroservicesForVersion
	result, err = registry.GetMicroservicesForVersion(ctx, inMemDao, &domain.DeploymentVersion{Version: "v2"})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result))
	verifyServicesVersion(t, dto.VersionInRegistry{
		Version: "v2",
		Stage:   domain.ActiveStage,
		Clusters: []dto.Microservice{
			{Cluster: "test-cluster1", Namespace: msaddr.LocalNamespace, Endpoints: []string{"http://test-cluster1-v2:8080", "https://test-cluster1-v2:8443"}},
			{Cluster: "test-cluster2", Namespace: "test-ns", Endpoints: []string{"http://test-cluster2:8080", "https://test-cluster2:8443"}},
		},
	}, result)

	// test Get As Map
	servicesMap, err := registry.GetMicroservicesByVersionAsMap(ctx, inMemDao, &domain.DeploymentVersion{Version: "v2"})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(servicesMap))
	isCluster1Present := false
	isCluster2Present := false
	for msKey, ms := range servicesMap {
		if msKey.Name == "test-cluster1" {
			assert.Equal(t, msaddr.LocalNamespace, msKey.Namespace)
			assert.Equal(t, "test-cluster1", ms.Name)
			assert.Equal(t, msaddr.LocalNamespace, ms.Namespace)
			assert.Equal(t, "v2", ms.DeploymentVersion)
			assert.Equal(t, "v2", ms.InitialDeploymentVersion)
			isCluster1Present = true
		} else if msKey.Name == "test-cluster2" {
			assert.Equal(t, "test-ns", msKey.Namespace)
			assert.Equal(t, "test-cluster2", ms.Name)
			assert.Equal(t, "test-ns", ms.Namespace)
			assert.Equal(t, "v2", ms.DeploymentVersion)
			assert.Equal(t, "v1", ms.InitialDeploymentVersion)
			isCluster2Present = true
		}
	}
	assert.True(t, isCluster1Present)
	assert.True(t, isCluster2Present)
}

func verifyServicesVersion(t *testing.T, expectedVersion dto.VersionInRegistry, result []dto.VersionInRegistry) {
	for _, actualVersion := range result {
		if actualVersion.Version == expectedVersion.Version {
			assert.Equal(t, expectedVersion.Stage, actualVersion.Stage)
			assert.Equal(t, len(expectedVersion.Clusters), len(actualVersion.Clusters))
			for _, expectedCluster := range expectedVersion.Clusters {
				actualCluster := getMicroservice(t, actualVersion.Clusters, expectedCluster.Cluster, expectedCluster.Namespace)
				assert.Equal(t, len(expectedCluster.Endpoints), len(actualCluster.Endpoints))
				assert.True(t, util.SliceContains(actualCluster.Endpoints, expectedCluster.Endpoints...))
			}
			return
		}
	}
	t.Fail()
}

func getMicroservice(t *testing.T, slice []dto.Microservice, serviceName, namespace string) dto.Microservice {
	for _, microservice := range slice {
		if serviceName == microservice.Cluster && namespace == microservice.Namespace {
			return microservice
		}
	}
	t.Fail()
	return dto.Microservice{}
}

func TestRegistry_GetConfigRes(t *testing.T) {
	_, bgSrv := prepareTest("v1")

	registry, ok := bgSrv.versionsRegistry.(*versionsRegistry[dto.ServicesVersionPayload])
	assert.True(t, ok)

	configRes := registry.GetConfigRes()
	assert.True(t, registry == configRes.Applier)
}

func TestRegistry_GetMicroserviceCurrentVersion(t *testing.T) {
	inMemDao, bgSrv := prepareTest("v1")
	v1 := domain.NewDeploymentVersion("v1", domain.ActiveStage)
	saveDeploymentVersions(t, inMemDao, v1)

	mockListener := NewMockListener(t)

	registry, ok := bgSrv.versionsRegistry.(*versionsRegistry[dto.ServicesVersionPayload])
	assert.True(t, ok)

	actualVersions, err := registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms1", msaddr.CurrentNamespace(), "v1")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms1", msaddr.CurrentNamespace(), "v2")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	// apply v1 services
	_, err = registry.Apply(ctx, dto.ServicesVersionPayload{
		Services: []string{"test-ms1"},
		Version:  "v1",
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mockListener.Events))

	verifyMicroserviceVersion(t, registry, "test-ms1", "v1", domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})

	// apply v2 services
	_, err = registry.Apply(ctx, dto.ServicesVersionPayload{
		Services: []string{"test-ms1", "test-ms2"},
		Version:  "v2",
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mockListener.Events))
	mockListener.VerifyLastEvent(t, true, "v2")

	// verify test-ms1 still exists in v1
	verifyMicroserviceVersion(t, registry, "test-ms1", "v1", domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})

	// verify test-ms1 also exists in v2
	verifyMicroserviceVersion(t, registry, "test-ms1", "v2", domain.DeploymentVersion{Version: "v2", Stage: domain.CandidateStage})

	// verify test-ms2 presents only in v2
	verifyMicroserviceVersion(t, registry, "test-ms2", "v2", domain.DeploymentVersion{Version: "v2", Stage: domain.CandidateStage})
	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms2", msaddr.CurrentNamespace(), "v1")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	// delete test-ms1 from v2
	falseVal := false
	_, err = registry.Apply(ctx, dto.ServicesVersionPayload{
		Services: []string{"test-ms1"},
		Version:  "v2",
		Exists:   &falseVal,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mockListener.Events))

	// verify test-ms1 only exists in v1
	verifyMicroserviceVersion(t, registry, "test-ms1", "v1", domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms1", msaddr.CurrentNamespace(), "v2")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	// verify test-ms2 presents only in v2
	verifyMicroserviceVersion(t, registry, "test-ms2", "v2", domain.DeploymentVersion{Version: "v2", Stage: domain.CandidateStage})
	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms2", msaddr.CurrentNamespace(), "v1")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	// delete test-ms2 from v2
	_, err = registry.Apply(ctx, dto.ServicesVersionPayload{
		Services: []string{"test-ms2"},
		Version:  "v2",
		Exists:   &falseVal,
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mockListener.Events))
	mockListener.VerifyLastEvent(t, false, "v2")

	// verify test-ms1 only exists in v1
	verifyMicroserviceVersion(t, registry, "test-ms1", "v1", domain.DeploymentVersion{Version: "v1", Stage: domain.ActiveStage})
	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms1", msaddr.CurrentNamespace(), "v2")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)

	// verify there is no test-ms2 in v2
	actualVersions, err = registry.GetMicroserviceCurrentVersion(ctx, inMemDao, "test-ms2", msaddr.CurrentNamespace(), "v2")
	assert.Nil(t, err)
	assert.Empty(t, actualVersions)
}

type MockListener struct {
	Events []*events.ChangeEvent
}

func NewMockListener(t *testing.T) *MockListener {
	mockListener := MockListener{Events: make([]*events.ChangeEvent, 0, 4)}
	bus.GetInternalBusInstance().Subscribe(bus.TopicBgRegistry, func(data interface{}) {
		event, ok := data.(*events.ChangeEvent)
		assert.True(t, ok)
		mockListener.Events = append(mockListener.Events, event)
	})
	return &mockListener
}

func (m *MockListener) VerifyLastEvent(t *testing.T, isCreate bool, version string) {
	assert.NotEmpty(t, m.Events)
	event := m.Events[len(m.Events)-1]
	assert.NotNil(t, event)
	versionChanges, ok := event.Changes[domain.DeploymentVersionTable]
	assert.True(t, ok)
	assert.Equal(t, 1, len(versionChanges))
	assert.Equal(t, isCreate, versionChanges[0].Created())
	if isCreate {
		assert.Nil(t, versionChanges[0].Before)
		dVersion, ok := versionChanges[0].After.(*domain.DeploymentVersion)
		assert.True(t, ok)
		assert.Equal(t, version, dVersion.Version)
	} else {
		assert.Nil(t, versionChanges[0].After)
		dVersion, ok := versionChanges[0].Before.(*domain.DeploymentVersion)
		assert.True(t, ok)
		assert.Equal(t, version, dVersion.Version)
	}
}

func verifyMicroserviceVersion(t *testing.T, registry *versionsRegistry[dto.ServicesVersionPayload], serviceName, initialVersion string, version domain.DeploymentVersion) {
	actualVersions, err := registry.GetMicroserviceCurrentVersion(ctx, registry.dao, serviceName, msaddr.CurrentNamespace(), initialVersion)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualVersions))
	assert.Equal(t, version.Version, actualVersions[0].Version)
	assert.Equal(t, version.Stage, actualVersions[0].Stage)
	clusters := actualVersions[0].Clusters
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, serviceName, clusters[0].Cluster)
	assert.Equal(t, msaddr.LocalNamespace, clusters[0].Namespace)
	assert.Equal(t, 0, len(clusters[0].Endpoints))
}
