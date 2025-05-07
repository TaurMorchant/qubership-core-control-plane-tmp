package dao

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	testDVersion      = "v1"
	numberOfRoutes    = 20
	numberOfClusters  = 5
	numberOfEndpoints = 10
)

type readWriteFunction func(*testing.T, *sync.WaitGroup, int, *InMemRepo)

type DoNothingGenerator struct{}

func (g *DoNothingGenerator) Generate(uniqEntity domain.Unique) error {
	return nil
}

type GeneratorMock struct {
	counter int32
}

func (g *GeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}

func TestInMemDao_WithWTxCreate(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	cluster := createCluster()

	payload := func(dao Repository) error {
		return dao.SaveCluster(cluster)
	}

	changes, err := testable.WithWTx(payload)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 1, len(changes))
	assert.Equal(t, domain.ClusterTable, changes[0].Table)
	assert.True(t, changes[0].Created())
	assert.Equal(t, cluster, changes[0].After)
}

func TestInMemDao_WithWtxChangesOrdering(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	routeConfig := createRouteConfig()
	virtualHost := createVirtualHost()
	virtualHost.RouteConfigurationId = routeConfig.Id
	route := createRoute()
	route.VirtualHostId = virtualHost.Id

	payload := func(dao Repository) error {
		if err := dao.SaveRouteConfig(routeConfig); err != nil {
			return err
		}
		if err := dao.SaveVirtualHost(virtualHost); err != nil {
			return err
		}
		if err := dao.SaveRoute(route); err != nil {
			return err
		}
		return nil
	}

	changes, err := testable.WithWTx(payload)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 3, len(changes))
	assert.Equal(t, domain.RouteConfigurationTable, changes[0].Table)
	assert.Equal(t, domain.VirtualHostTable, changes[1].Table)
	assert.Equal(t, domain.RouteTable, changes[2].Table)
}

func Test_SingleGoroutineReadWriteTest(t *testing.T) {
	numberOfGoroutines := 1
	testReadWriteInternal(t, numberOfGoroutines, readInSeparateGoroutine)
	testReadWriteInternal(t, numberOfGoroutines, saveInSeparateGoroutine)
}

func Test_TwoGoroutinesReadWriteTest(t *testing.T) {
	numberOfGoroutines := 2
	testReadWriteInternal(t, numberOfGoroutines, readInSeparateGoroutine)
	testReadWriteInternal(t, numberOfGoroutines, saveInSeparateGoroutine)
}

func Test_TenGoroutinesReadWriteTest(t *testing.T) {
	numberOfGoroutines := 10
	testReadWriteInternal(t, numberOfGoroutines, readInSeparateGoroutine)
	testReadWriteInternal(t, numberOfGoroutines, saveInSeparateGoroutine)
}

func Test_TwentyGoroutinesReadWriteTest(t *testing.T) {
	numberOfGoroutines := 20
	testReadWriteInternal(t, numberOfGoroutines, readInSeparateGoroutine)
	testReadWriteInternal(t, numberOfGoroutines, saveInSeparateGoroutine)
}

func Test_SingleGoroutineConcurrentWriteTest(t *testing.T) {
	mockDao := getInMemRepo()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	saveOrUpdateDeploymentInTransaction(t, wg, mockDao)
	readAndCheckDeploymentInTransaction(t, mockDao, domain.ActiveStage, "")
	wg.Wait()
}

func Test_TwoGoroutineConcurrentWriteTest(t *testing.T) {
	mockDao := getInMemRepo()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go saveOrUpdateDeploymentInTransaction(t, wg, mockDao)
	go saveOrUpdateDeploymentInTransaction(t, wg, mockDao)
	wg.Wait()
	readAndCheckDeploymentInTransaction(t, mockDao, domain.LegacyStage, domain.ActiveStage)
}

func Test_GoroutineWriteAndReadTest(t *testing.T) {
	mockDao := getInMemRepo()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go saveOrUpdateDeploymentInTransactionWithStage(t, wg, mockDao, domain.LegacyStage)
	time.Sleep(100 * time.Millisecond)
	go saveOrUpdateDeploymentInTransactionWithStage(t, wg, mockDao, domain.CandidateStage)
	wg.Wait()
	readAndCheckDeploymentInTransaction(t, mockDao, domain.CandidateStage, domain.ActiveStage, domain.LegacyStage, domain.ActiveStage)
	dV, err := mockDao.FindAllDeploymentVersions()
	assert.Empty(t, err)
	assert.NotEmpty(t, dV)
	assert.Equal(t, 1, len(dV))
}

func Test_ConcurrentDeploymentUpdate(t *testing.T) {
	mockDao := getInMemRepo()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go saveOrUpdateDeploymentInTransactionWithStage(t, wg, mockDao, domain.ActiveStage)
	go saveOrUpdateDeploymentInTransactionWithStage(t, wg, mockDao, domain.LegacyStage)
	time.Sleep(100 * time.Millisecond)
	go saveOrUpdateDeploymentInTransactionWithStage(t, wg, mockDao, domain.CandidateStage)
	wg.Wait()
	readAndCheckDeploymentInTransaction(t, mockDao, domain.CandidateStage, domain.ActiveStage, domain.LegacyStage)
	dV, err := mockDao.FindAllDeploymentVersions()
	assert.Empty(t, err)
	assert.NotEmpty(t, dV)
	assert.Equal(t, 1, len(dV))
}

func Test_ConcurrentDeploymentUpdateAndRead(t *testing.T) {
	mockDao := getInMemRepo()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	signal := make(chan bool)
	saveOrUpdateDeploymentInTransaction(t, wg, mockDao)
	go readDeploymentInTransactionWithStage(t, wg, mockDao, signal)
	go saveDeploymentInTransactionWithStage(t, wg, mockDao, signal)
	wg.Wait()
	readAndCheckDeploymentInTransaction(t, mockDao, domain.LegacyStage, domain.CandidateStage, domain.ActiveStage)
	dV, err := mockDao.FindAllDeploymentVersions()
	assert.Empty(t, err)
	assert.NotEmpty(t, dV)
	assert.Equal(t, 1, len(dV))
}

// PRIVATE
func createRoute() *domain.Route {
	routeUuid := uuid.New().String()
	return &domain.Route{
		Uuid:                     routeUuid,
		RouteKey:                 "||route",
		Prefix:                   "/api/v1",
		ClusterName:              "cluster",
		Version:                  1,
		DeploymentVersion:        "v1",
		InitialDeploymentVersion: "v1",
	}
}

func createVirtualHost() *domain.VirtualHost {
	return &domain.VirtualHost{
		Name:    "virtual-host",
		Version: 1,
	}
}

func createRouteConfig() *domain.RouteConfiguration {
	return &domain.RouteConfiguration{
		Name:        "route-config",
		Version:     1,
		NodeGroupId: "test-nodeGroup",
	}
}

func createCluster() *domain.Cluster {
	var httpVersion int32 = 1
	return &domain.Cluster{
		Name:          "cluster",
		LbPolicy:      "LEAST_REQUEST",
		DiscoveryType: "STRICT_DNS",
		Version:       1,
		HttpVersion:   &httpVersion,
	}
}

func testReadWriteInternal(t *testing.T, numberOfGoroutines int, f readWriteFunction) {
	mockDao := getInMemRepo()
	testMultiGoroutineReadWrite(t, mockDao, numberOfGoroutines, f)
	checkNumberOfEntities(t, mockDao, numberOfGoroutines)
	checkData(t, mockDao, numberOfGoroutines)
}

func testMultiGoroutineReadWrite(t *testing.T, mockDao *InMemRepo, numberOfGoroutines int, f readWriteFunction) {
	signal := make(chan bool)
	go loopReadInTransaction(t, mockDao, signal, 0)
	<-signal
	wg := &sync.WaitGroup{}
	for i := 1; i <= numberOfGoroutines; i++ {
		wg.Add(1)
		f(t, wg, i, mockDao)
	}
	signal <- true
	wg.Wait()
}

func getInMemRepo() *InMemRepo {
	return &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
}

func readInSeparateGoroutine(t *testing.T, wg *sync.WaitGroup, i int, mockDao *InMemRepo) {
	saveInTransaction(t, wg, i, mockDao)
	go loopReadInTransaction(t, mockDao, nil, i)
}

func saveInSeparateGoroutine(t *testing.T, wg *sync.WaitGroup, i int, mockDao *InMemRepo) {
	go saveInTransaction(t, wg, i, mockDao)
}

// read and test data
func loopReadInTransaction(t *testing.T, storage *InMemRepo, signal chan bool, iterNumber int) {
	signal <- true
	err := storage.WithRTx(func(dao Repository) error {
		for {
			checkNumberOfEntities(t, dao, iterNumber)
			checkData(t, dao, iterNumber)
			select {
			case <-signal:
				return nil
			case <-time.After(50 * time.Millisecond):

			}
		}
	})
	assert.Empty(t, err)
}

func readAndCheckDeploymentInTransaction(t *testing.T, storage *InMemRepo, expectedStage string, oldStages ...string) {
	err := storage.WithRTx(func(dao Repository) error {
		readAndCheckDeploymentVersion(t, dao, expectedStage, oldStages...)
		return nil
	})
	assert.Empty(t, err)
}

func readAndCheckDeploymentVersion(t *testing.T, dao Repository, expectedStage string, oldStages ...string) {
	dV, err := dao.FindDeploymentVersionsByStage(expectedStage)
	assert.Nil(t, err)
	assert.NotEmpty(t, dV)
	assert.Equal(t, 1, len(dV))
	assert.Equal(t, expectedStage, dV[0].Stage)
	for _, oldStage := range oldStages {
		dV, err = dao.FindDeploymentVersionsByStage(oldStage)
		assert.Nil(t, err)
		assert.Empty(t, dV)
	}
}

func checkNumberOfEntities(t *testing.T, dao Repository, goroutine int) {
	routes, err := dao.FindAllRoutes()
	assert.Empty(t, err)
	assert.Equal(t, numberOfRoutes*goroutine, len(routes))
	clusters, err := dao.FindAllClusters()
	assert.Empty(t, err)
	assert.Equal(t, numberOfClusters*goroutine, len(clusters))
	endpoints, err := dao.FindAllEndpoints()
	assert.Empty(t, err)
	assert.Equal(t, numberOfClusters*numberOfEndpoints*goroutine, len(endpoints))
}

func checkData(t *testing.T, dao Repository, goroutine int) {
	checkRoutes(t, dao, goroutine)
	checkClustersWithEndpoints(t, dao, goroutine)
}

func checkClustersWithEndpoints(t *testing.T, dao Repository, goroutine int) {
	clusters, err := dao.FindAllClusters()
	assert.Empty(t, err)
	assert.Equal(t, goroutine*numberOfClusters, len(clusters))
	for _, cluster := range clusters {
		assert.NotEmpty(t, cluster.Id)
		assert.Equal(t, fmt.Sprintf("cluster-%d", cluster.Id), cluster.Name)
		checkEndpoints(t, dao, cluster.Id)
	}
}

func checkEndpoints(t *testing.T, dao Repository, clusterId int32) {
	endpoints, err := dao.FindEndpointsByClusterId(clusterId)
	assert.Empty(t, err)
	assert.NotEmpty(t, endpoints)
	assert.Equal(t, numberOfEndpoints, len(endpoints))
	endpointAddress := fmt.Sprintf("http://some-address-v%d", clusterId)
	for _, endpoint := range endpoints {
		assert.NotEmpty(t, endpoint.Id)
		assert.NotEmpty(t, endpoint.DeploymentVersion)
		checkDeploymentVersion(t, dao, endpoint.DeploymentVersion)
		assert.Equal(t, endpointAddress, endpoint.Address)
	}
}

func checkRoutes(t *testing.T, dao Repository, goroutine int) {
	routes, err := dao.FindAllRoutes()
	assert.Empty(t, err)
	assert.Equal(t, numberOfRoutes*goroutine, len(routes))
	for _, route := range routes {
		assert.NotEmpty(t, route.Id)
		checkDeploymentVersion(t, dao, route.DeploymentVersion)
		assert.Equal(t, fmt.Sprintf("route-key|r-%d", route.Id), route.RouteKey)
		assert.Equal(t, fmt.Sprintf("clustername|r-%d", route.Id), route.ClusterName)
		assert.NotEmpty(t, route.Uuid)
	}
}

func checkDeploymentVersion(t *testing.T, dao Repository, version string) {
	dVersion, err := dao.FindDeploymentVersion(version)
	assert.Empty(t, err)
	assert.NotEmpty(t, dVersion)
	assert.Equal(t, version, dVersion.Version)
}

//save data

func saveOrUpdateDeploymentInTransactionWithStage(t *testing.T, wg *sync.WaitGroup, storage *InMemRepo, stage string) {
	defer wg.Done()
	_, err := storage.WithWTx(func(dao Repository) error {
		//simulate work
		time.Sleep(200 * time.Millisecond)
		saveOrUpdateDeploymentVersion(t, dao, stage)
		return nil
	})
	assert.Nil(t, err)
}

func saveOrUpdateDeploymentInTransaction(t *testing.T, wg *sync.WaitGroup, storage *InMemRepo) {
	defer wg.Done()
	_, err := storage.WithWTx(func(dao Repository) error {
		saveOrUpdateDeploymentVersion(t, dao, domain.LegacyStage)
		return nil
	})
	assert.Nil(t, err)
}

func saveDeploymentInTransactionWithStage(t *testing.T, wg *sync.WaitGroup, storage *InMemRepo, signal chan bool) {
	defer wg.Done()
	_, err := storage.WithWTx(func(dao Repository) error {
		saveOrUpdateDeploymentVersion(t, dao, domain.LegacyStage)
		//write is completed
		signal <- true
		//waiting for reading
		<-signal
		saveOrUpdateDeploymentVersion(t, dao, domain.CandidateStage)
		//write is completed
		signal <- true
		//waiting for reading
		<-signal
		saveOrUpdateDeploymentVersion(t, dao, domain.LegacyStage)
		return nil
	})
	//transaction completed
	signal <- true
	assert.Nil(t, err)
}

func readDeploymentInTransactionWithStage(t *testing.T, wg *sync.WaitGroup, storage *InMemRepo, signal chan bool) {
	defer wg.Done()
	//waiting for writing
	<-signal
	err := storage.WithRTx(func(dao Repository) error {
		readAndCheckDeploymentVersion(t, dao, domain.ActiveStage, domain.CandidateStage, domain.LegacyStage)
		//reading is completed
		signal <- true
		//waiting for writing
		<-signal
		readAndCheckDeploymentVersion(t, dao, domain.ActiveStage, domain.CandidateStage, domain.LegacyStage)
		//reading is completed
		signal <- true
		//waiting for writing
		<-signal
		readAndCheckDeploymentVersion(t, dao, domain.ActiveStage, domain.CandidateStage, domain.LegacyStage)
		return nil
	})
	assert.Nil(t, err)
}

func saveOrUpdateDeploymentVersion(t *testing.T, dao Repository, newStage string) {
	dV, err := dao.FindDeploymentVersion(testDVersion)
	assert.Nil(t, err)
	if dV != nil {
		newDv := dV.Clone()
		newDv.Stage = newStage
		assert.Nil(t, dao.SaveDeploymentVersion(newDv))
	} else {
		saveDeploymentVersion(t, dao, domain.ActiveStage)
	}
}

func saveDeploymentVersion(t *testing.T, dao Repository, newStage string) {
	dV := domain.NewDeploymentVersion(testDVersion, newStage)
	assert.Nil(t, dao.SaveDeploymentVersion(dV))
}

func saveInTransaction(t *testing.T, wg *sync.WaitGroup, goroutine int, storage *InMemRepo) {
	defer wg.Done()
	_, err := storage.WithWTx(func(dao Repository) error {
		deploymentVersion := domain.NewDeploymentVersion(fmt.Sprintf("v%d", goroutine), domain.ActiveStage)
		assert.Nil(t, dao.SaveDeploymentVersion(deploymentVersion))
		saveRoutes(t, goroutine, dao, deploymentVersion.Version)
		saveClustersAndEndpoints(t, goroutine, dao, deploymentVersion.Version)
		return nil
	})
	assert.Nil(t, err)
}

func saveClustersAndEndpoints(t *testing.T, goroutine int, dao Repository, deploymentVersion string) {
	for i := 1; i <= numberOfClusters; i++ {
		clusterId := int32(numberOfClusters*goroutine + i)
		testCluster := domain.NewCluster(fmt.Sprintf("cluster-%d", clusterId), false)
		testCluster.SetId(clusterId)
		assert.Nil(t, dao.SaveCluster(testCluster))
		saveEndpoints(t, goroutine, dao, clusterId, deploymentVersion)
	}
}

func saveEndpoints(t *testing.T, goroutine int, dao Repository, clusterId int32, deploymentVersion string) {
	for j := 1; j <= numberOfEndpoints; j++ {
		endpointId := int32(numberOfEndpoints*int(clusterId)*goroutine + j)
		testEndpoint := domain.NewEndpoint(fmt.Sprintf("http://some-address-v%d", clusterId), 8080, deploymentVersion, deploymentVersion, clusterId)
		testEndpoint.SetId(endpointId)
		assert.Nil(t, dao.SaveEndpoint(testEndpoint))
	}
}

func saveRoutes(t *testing.T, goroutine int, dao Repository, deploymentVersion string) {
	for i := 1; i <= numberOfRoutes; i++ {
		saveRoute(t, goroutine, i, dao, deploymentVersion)
	}
}

func saveRoute(t *testing.T, goroutine, i int, dao Repository, deploymentVersion string) {
	id := numberOfRoutes*goroutine + i
	testRoute := &domain.Route{
		Id:                int32(id),
		RouteKey:          fmt.Sprintf("route-key|r-%d", id),
		ClusterName:       fmt.Sprintf("clustername|r-%d", id),
		DeploymentVersion: deploymentVersion,
		Uuid:              uuid.New().String(),
	}
	assert.Nil(t, dao.SaveRoute(testRoute))
}
