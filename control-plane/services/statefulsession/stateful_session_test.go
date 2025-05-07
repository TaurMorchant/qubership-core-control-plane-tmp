package statefulsession

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/stretchr/testify/assert"
	"sync"
	"sync/atomic"
	"testing"
)

func TestServiceImpl_FindAll(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	spec := &dto.StatefulSession{
		Version:   "",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "",
		Gateways:  []string{"private-gateway-service"},
		Port:      nil,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := srv.FindAll(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.Version)
	assert.Equal(t, "test-cluster", session.Cluster)
	assert.Equal(t, "default", session.Namespace)
	assert.Nil(t, session.Port)
	assert.Empty(t, session.Hostname)
	assert.Nil(t, session.Route)
	assert.NotNil(t, session.Enabled)
	assert.True(t, *session.Enabled)
	assert.NotNil(t, session.Cookie)
	assert.Equal(t, "sticky-cookie-v1", session.Cookie.Name)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestServiceImpl_FindAll_PerEndpoint(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		cluster := &domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		}
		err := repo.SaveCluster(cluster)
		assert.Nil(t, err)

		err = repo.SaveEndpoint(&domain.Endpoint{
			Address:                  "test-cluster-v1",
			Port:                     8080,
			ClusterId:                cluster.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		})
		assert.Nil(t, err)
		return nil
	})

	port := 8080
	spec := &dto.StatefulSession{
		Version:   "v1",
		Namespace: "",
		Cluster:   "test-cluster",
		//Hostname:  "test-cluster-v1",
		Gateways: []string{"private-gateway-service"},
		Port:     &port,
		Enabled:  nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := srv.FindAll(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.Version)
	assert.Equal(t, "test-cluster", session.Cluster)
	assert.Equal(t, "default", session.Namespace)
	assert.NotNil(t, session.Port)
	assert.Equal(t, 8080, *session.Port)
	assert.Equal(t, "test-cluster-v1", session.Hostname)
	assert.Nil(t, session.Route)
	assert.NotNil(t, session.Enabled)
	assert.True(t, *session.Enabled)
	assert.NotNil(t, session.Cookie)
	assert.Equal(t, "sticky-cookie-v1", session.Cookie.Name)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestServiceImpl_ApplyStatefulSessionPerCluster(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	spec := &dto.StatefulSession{
		Version:   "",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "",
		Gateways:  []string{"private-gateway-service"},
		Port:      nil,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestServiceImpl_ApplyStatefulSessionPerCluster_Upd(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		err := repo.SaveCluster(&domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		})
		assert.Nil(t, err)
		err = repo.SaveStatefulSessionConfig(&domain.StatefulSession{
			CookieName:               "sticky-cookie-old",
			CookiePath:               "/old",
			Enabled:                  true,
			ClusterName:              "test-cluster",
			Namespace:                "default",
			Gateways:                 []string{"private-gateway-service"},
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		})
		assert.Nil(t, err)
		return nil
	})

	spec := &dto.StatefulSession{
		Version:   "",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "",
		Gateways:  []string{"private-gateway-service"},
		Port:      nil,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "", session.CookiePath)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

func TestServiceImpl_ApplyStatefulSessionPerEndpointByVersion(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		cluster := &domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		}
		err := repo.SaveCluster(cluster)
		assert.Nil(t, err)

		err = repo.SaveEndpoint(&domain.Endpoint{
			Address:                  "test-cluster-v1",
			Port:                     8080,
			ClusterId:                cluster.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		})
		assert.Nil(t, err)
		return nil
	})

	port := 8080
	spec := &dto.StatefulSession{
		Version:   "v1",
		Namespace: "",
		Cluster:   "test-cluster",
		//Hostname:  "test-cluster-v1",
		Gateways: []string{"private-gateway-service"},
		Port:     &port,
		Enabled:  nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "", session.CookiePath)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])

	endpoint, err := genericDao.FindEndpointByStatefulSession(session.Id)
	assert.Nil(t, err)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "test-cluster-v1", endpoint.Address)
}

func TestServiceImpl_ApplyStatefulSessionPerEndpointByVersion_Upd(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		cluster := &domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		}
		err := repo.SaveCluster(cluster)
		assert.Nil(t, err)

		ses := &domain.StatefulSession{
			CookieName:               "sticky-cookie-old",
			CookiePath:               "/old",
			Enabled:                  true,
			ClusterName:              "test-cluster",
			Namespace:                "default",
			Gateways:                 []string{"private-gateway-service"},
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}
		err = repo.SaveStatefulSessionConfig(ses)
		assert.Nil(t, err)

		err = repo.SaveEndpoint(&domain.Endpoint{
			Address:                  "test-cluster-v1",
			Port:                     8080,
			ClusterId:                cluster.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
			StatefulSessionId:        ses.Id,
		})
		assert.Nil(t, err)
		return nil
	})

	port := 8080
	spec := &dto.StatefulSession{
		Version:   "v1",
		Namespace: "",
		Cluster:   "test-cluster",
		//Hostname:  "test-cluster-v1",
		Gateways: []string{"private-gateway-service"},
		Port:     &port,
		Enabled:  nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "", session.CookiePath)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])

	endpoint, err := genericDao.FindEndpointByStatefulSession(session.Id)
	assert.Nil(t, err)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "test-cluster-v1", endpoint.Address)
}

func TestServiceImpl_ApplyStatefulSessionPerEndpointByHostname(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		cluster := &domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		}
		err := repo.SaveCluster(cluster)
		assert.Nil(t, err)

		err = repo.SaveEndpoint(&domain.Endpoint{
			Address:                  "test-cluster-v1",
			Port:                     8080,
			ClusterId:                cluster.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		})
		assert.Nil(t, err)
		return nil
	})

	port := 8080
	spec := &dto.StatefulSession{
		//Version:   "v1",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "test-cluster-v1",
		Gateways:  []string{"private-gateway-service"},
		Port:      &port,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "", session.CookiePath)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])

	endpoint, err := genericDao.FindEndpointByStatefulSession(session.Id)
	assert.Nil(t, err)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "test-cluster-v1", endpoint.Address)
}

func TestServiceImpl_ApplyStatefulSessionPerEndpointByHostname_Upd(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	_, _ = genericDao.WithWTx(func(repo dao.Repository) error {
		cluster := &domain.Cluster{
			Name:     "test-cluster||test-cluster||8080",
			LbPolicy: "RING_HASH",
			Version:  1,
		}
		err := repo.SaveCluster(cluster)
		assert.Nil(t, err)

		ses := &domain.StatefulSession{
			CookieName:               "sticky-cookie-old",
			CookiePath:               "/old",
			Enabled:                  true,
			ClusterName:              "test-cluster",
			Namespace:                "default",
			Gateways:                 []string{"private-gateway-service"},
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
		}
		err = repo.SaveStatefulSessionConfig(ses)
		assert.Nil(t, err)

		err = repo.SaveEndpoint(&domain.Endpoint{
			Address:                  "test-cluster-v1",
			Port:                     8080,
			ClusterId:                cluster.Id,
			DeploymentVersion:        "v1",
			InitialDeploymentVersion: "v1",
			StatefulSessionId:        ses.Id,
		})
		assert.Nil(t, err)
		return nil
	})

	port := 8080
	spec := &dto.StatefulSession{
		//Version:   "v1",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "test-cluster-v1",
		Gateways:  []string{"private-gateway-service"},
		Port:      &port,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, "", session.CookiePath)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])

	endpoint, err := genericDao.FindEndpointByStatefulSession(session.Id)
	assert.Nil(t, err)
	assert.NotNil(t, endpoint)
	assert.Equal(t, "test-cluster-v1", endpoint.Address)
}

func TestServiceImpl_ApplyStatefulSession(t *testing.T) {
	genericDao := dao.NewInMemDao(ram.NewStorage(), &idGeneratorMock{}, nil)
	busPublisher := &busMock{calls: map[string]int{}}
	srv := NewService(genericDao, entity.NewService("v1"), busPublisher)

	spec := &dto.StatefulSession{
		Version:   "",
		Namespace: "",
		Cluster:   "test-cluster",
		Hostname:  "",
		Gateways:  []string{"private-gateway-service"},
		Port:      nil,
		Enabled:   nil,
		Cookie: &dto.Cookie{
			Name: "sticky-cookie",
		},
	}
	err := srv.ApplyStatefulSession(context.Background(), spec)
	assert.Nil(t, err)
	assert.Greater(t, busPublisher.GetCalls(bus.TopicMultipleChanges), 0)

	sessions, err := genericDao.FindAllStatefulSessionConfigs()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sessions))
	session := sessions[0]
	assert.Equal(t, "v1", session.DeploymentVersion)
	assert.Equal(t, "v1", session.InitialDeploymentVersion)
	assert.Equal(t, "test-cluster", session.ClusterName)
	assert.Equal(t, "default", session.Namespace)
	assert.True(t, session.Enabled)
	assert.Equal(t, "sticky-cookie-v1", session.CookieName)
	assert.Equal(t, 1, len(session.Gateways))
	assert.Equal(t, "private-gateway-service", session.Gateways[0])
}

type busMock struct {
	mutex sync.Mutex
	calls map[string]int
}

func (m *busMock) GetCalls(topic string) int {
	return m.calls[topic]
}

func (m *busMock) Publish(topic string, data interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.calls[topic]++
	return nil
}

func (m *busMock) Shutdown() {}

type idGeneratorMock struct {
	counter int32
}

func (g *idGeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&g.counter, 1))
	}
	return nil
}
