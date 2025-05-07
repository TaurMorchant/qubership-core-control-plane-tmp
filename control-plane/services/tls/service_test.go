package tls

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
)

func TestSaveTlsConfig_CreateNew_BadConfigs(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, inMemDao := getService()
	const testSni = "test-sni"

	//Bad CA
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badCaCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse certificate")
	configs, err := inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad first CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFirstCaCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM format")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad middle CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badMiddleCaCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM format")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad last CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badLastCaCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to decode PEM format")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted start CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedStartCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted end CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedEndCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted start first CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedStartFirstCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted end first CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedEndFirstCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted start second CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedStartSecondCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad formatted end second CA
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: badFormattedEndSecondCert,
			SNI:       testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "wrong cert format: certificate must start with")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad client cert
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  goodCaCert,
			ClientCert: badClientCert,
			PrivateKey: goodPrivateKey,
			SNI:        testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse certificate")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Bad private key
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  goodCaCert,
			ClientCert: goodClientCert,
			PrivateKey: badPrivateKey,
			SNI:        testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to parse private key")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))

	//Client cert does not match with key
	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  goodCaCert,
			ClientCert: invalidClientCertForMatch,
			PrivateKey: prKeyForMatch,
			SNI:        testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "private key does not match with provided client certificate")
	configs, err = inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 0, len(configs))
}

func TestSaveTlsConfig_CreateNew_OnlyTrustedCa(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, inMemDao := getService()
	const testSni = "test-sni"
	const testCa = goodCaCerts
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:   true,
			Insecure:  true,
			TrustedCA: testCa,
			SNI:       testSni,
		},
	})
	require.Nil(t, err)
	tlsConfig, err := inMemDao.FindTlsConfigByName(tlsConfigName)
	require.Nil(t, err)
	require.Equal(t, tlsConfigName, tlsConfig.Name)
	require.True(t, tlsConfig.Enabled)
	require.True(t, tlsConfig.Insecure)
	require.Equal(t, testSni, tlsConfig.SNI)
	require.Equal(t, testCa, tlsConfig.TrustedCA)

	configs, err := inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 1, len(configs))
}

func TestSaveTlsConfig_CreateNew(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, inMemDao := getService()
	const testSni = "test-sni"
	const testCa = goodCaCert
	const testClientCert = goodClientCert
	const testPrivateKey = goodPrivateKey
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  testCa,
			ClientCert: testClientCert,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.Nil(t, err)
	tlsConfig, err := inMemDao.FindTlsConfigByName(tlsConfigName)
	require.Nil(t, err)
	require.Equal(t, tlsConfigName, tlsConfig.Name)
	require.True(t, tlsConfig.Enabled)
	require.True(t, tlsConfig.Insecure)
	require.Equal(t, testSni, tlsConfig.SNI)
	require.Equal(t, testCa, tlsConfig.TrustedCA)
	require.Equal(t, testClientCert, tlsConfig.ClientCert)
	require.Equal(t, testPrivateKey, tlsConfig.PrivateKey)

	configs, err := inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 1, len(configs))
}

func TestDeleteTlsConfig(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, inMemDao := getService()
	const testSni = "test-sni"
	const testCa = goodCaCerts
	const testClientCert = goodClientCert
	const testPrivateKey = goodPrivateKey
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  testCa,
			ClientCert: testClientCert,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.Nil(t, err)

	service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls:  nil,
	})

	tlsConfig, err := inMemDao.FindTlsConfigByName(tlsConfigName)
	require.Nil(t, err)
	require.Nil(t, tlsConfig)
}

func TestSaveTlsConfig_UpdateSame(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, inMemDao := getService()
	const testSni = "test-sni"
	const testCa = goodCaCert
	const testClientCert = goodClientCert
	const testPrivateKey = goodPrivateKey
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   true,
			TrustedCA:  testCa,
			ClientCert: testClientCert,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.Nil(t, err)
	tlsConfig, err := inMemDao.FindTlsConfigByName(tlsConfigName)
	require.Nil(t, err)
	require.True(t, tlsConfig.Enabled)

	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    false,
			Insecure:   true,
			TrustedCA:  testCa,
			ClientCert: testClientCert,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.Nil(t, err)
	tlsConfig, err = inMemDao.FindTlsConfigByName(tlsConfigName)
	require.Nil(t, err)
	require.False(t, tlsConfig.Enabled)

	configs, err := inMemDao.FindAllTlsConfigs()
	require.Nil(t, err)
	require.Equal(t, 1, len(configs))
}

func TestSaveTlsConfig_GetGlobalTlsConfigs(t *testing.T) {
	tlsConfigName := "test-tls-config"
	clusterName := "first-cluster"
	nodeGroupName := "first-ng"
	service, inMemDao := getService()

	_, err := inMemDao.WithWTx(func(dao dao.Repository) error {
		err := dao.SaveNodeGroup(domain.NewNodeGroup(nodeGroupName))
		if err != nil {
			return err
		}

		cluster := domain.NewCluster(clusterName, true)
		err = dao.SaveCluster(cluster)
		if err != nil {
			return err
		}

		err = dao.SaveClustersNodeGroup(domain.NewClusterNodeGroups(cluster.Id, nodeGroupName))
		if err != nil {
			return err
		}

		return nil
	})
	require.Nil(t, err)

	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name:               tlsConfigName,
		TrustedForGateways: []string{nodeGroupName, "second"},
		Tls: &dto.Tls{
			Enabled:  true,
			Insecure: true,
		},
	})
	require.Nil(t, err)

	cluster, err := inMemDao.FindClusterByName(clusterName)
	require.Nil(t, err)
	tlsConfigs, err := service.GetGlobalTlsConfigs(cluster)
	require.Nil(t, err)
	require.Equal(t, tlsConfigName, tlsConfigs[0].Name)
}

func TestValidateAndSaveTlsConfig(t *testing.T) {
	tlsConfigName := "test-tls-config"
	service, _ := getService()
	const testSni = "test-sni"
	const testCa = "test-ca"
	const testClientCert = "test-client-cert"
	const testPrivateKey = "test-private-key"
	const emptyValue = ""
	err := service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   false,
			TrustedCA:  emptyValue,
			ClientCert: testClientCert,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "TlsDef must be insecure or must contain non-empty trustedCA")

	err = service.SaveTlsConfig(nil, &dto.TlsConfig{
		Name: tlsConfigName,
		Tls: &dto.Tls{
			Enabled:    true,
			Insecure:   false,
			TrustedCA:  testCa,
			ClientCert: emptyValue,
			PrivateKey: testPrivateKey,
			SNI:        testSni,
		},
	})
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "TlsDef for mTLS must contain both non-empty clientCert and privateKey")
}

func NewService(dao dao.Dao, bus bus.BusPublisher) *Service {
	return &Service{
		dao: dao,
		bus: bus,
	}
}

func getService() (*Service, *dao.InMemDao) {
	inMemStorage := ram.NewStorage()
	internalBus := bus.GetInternalBusInstance()
	eventBus := bus.NewEventBusAggregator(inMemStorage, internalBus, internalBus, nil, nil)
	genericDao := dao.NewInMemDao(inMemStorage, &idGeneratorMock{}, []func([]memdb.Change) error{flushChanges})

	service := NewService(genericDao, eventBus)
	_, _ = genericDao.WithWTx(func(dao dao.Repository) error {
		_ = dao.SaveDeploymentVersion(&domain.DeploymentVersion{
			Version: "v1",
			Stage:   domain.ActiveStage,
		})
		return nil
	})
	return service, genericDao
}

func flushChanges(_ []memdb.Change) error {
	return nil
}

type idGeneratorMock struct {
	seq int32
}

func (generator *idGeneratorMock) Generate(uniqEntity domain.Unique) error {
	if uniqEntity.GetId() == 0 {
		uniqEntity.SetId(atomic.AddInt32(&generator.seq, 1))
	}
	return nil
}
