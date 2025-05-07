package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemRepo_DeleteTlsConfigById(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	tlsId := createAndSaveTlsConfig(t, repo, "test")
	assert.NotNil(t, getTlsConfigById(t, repo, tlsId))

	deleteTestTlsConfigById(t, repo, tlsId)
	assert.Nil(t, getTlsConfigById(t, repo, tlsId))
}

func TestInMemRepo_FindAllTlsConfigs(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	createAndSaveTlsConfig(t, repo, "first")
	createAndSaveTlsConfig(t, repo, "second")
	configs, err := repo.WithRTxVal(func(dao Repository) (interface{}, error) {
		configs, err := dao.FindAllTlsConfigs()
		assert.Nil(t, err)
		return configs, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(configs.([]*domain.TlsConfig)))
	configNames := make([]string, 2)
	for i, config := range configs.([]*domain.TlsConfig) {
		configNames[i] = config.Name
	}
	assert.ElementsMatch(t, configNames, []string{"first", "second"})
}

func TestInMemRepo_FindAllTlsConfigsByNodeGroup(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	saveTlsConfig(t, repo, createTlsConfig("first", []*domain.NodeGroup{
		{
			Name: "first-ng",
		},
	}))
	saveTlsConfig(t, repo, createTlsConfig("second", []*domain.NodeGroup{
		{
			Name: "first-ng",
		},
	}))
	saveTlsConfig(t, repo, createTlsConfig("third", []*domain.NodeGroup{
		{
			Name: "second-ng",
		},
	}))

	configs, err := repo.WithRTxVal(func(dao Repository) (interface{}, error) {
		configs, err := dao.FindAllTlsConfigsByNodeGroup("first-ng")
		assert.Nil(t, err)
		return configs, nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(configs.([]*domain.TlsConfig)))
	configNames := make([]string, 2)
	for i, config := range configs.([]*domain.TlsConfig) {
		configNames[i] = config.Name
	}
	assert.ElementsMatch(t, configNames, []string{"first", "second"})
}

func TestInMemRepo_TlsConfigNodeGroupsOverride(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	_, err := repo.WithWTx(func(repo Repository) error {
		if err := repo.SaveNodeGroup(&domain.NodeGroup{
			Name: "first-ng",
		}); err != nil {
			return err
		}
		return repo.SaveNodeGroup(&domain.NodeGroup{
			Name: "second-ng",
		})
	})
	assert.Nil(t, err)

	// TC 1: create with 1 node group
	saveTlsConfig(t, repo, createTlsConfig("first", []*domain.NodeGroup{
		{
			Name: "first-ng",
		},
	}))
	tlsConfigs, err := repo.FindAllTlsConfigsByNodeGroup("first-ng")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigs))
	tlsConfigs, err = repo.FindAllTlsConfigsByNodeGroup("second-ng")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tlsConfigs))

	actualTls, err := repo.FindTlsConfigByName("first")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualTls.NodeGroups))

	// TC 2: add another node group
	saveTlsConfig(t, repo, createTlsConfig("first", []*domain.NodeGroup{
		{
			Name: "first-ng",
		},
		{
			Name: "second-ng",
		},
	}))
	tlsConfigs, err = repo.FindAllTlsConfigsByNodeGroup("first-ng")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigs))
	tlsConfigs, err = repo.FindAllTlsConfigsByNodeGroup("second-ng")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigs))

	actualTls, err = repo.FindTlsConfigByName("first")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(actualTls.NodeGroups))

	// TC 3: remove first node group
	saveTlsConfig(t, repo, createTlsConfig("first", []*domain.NodeGroup{
		{
			Name: "second-ng",
		},
	}))
	tlsConfigs, err = repo.FindAllTlsConfigsByNodeGroup("first-ng")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tlsConfigs))
	tlsConfigs, err = repo.FindAllTlsConfigsByNodeGroup("second-ng")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigs))

	actualTls, err = repo.FindTlsConfigByName("first")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(actualTls.NodeGroups))
}

func TestInMemRepo_FindTlsConfigById(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	id := createAndSaveTlsConfig(t, repo, "test")
	config, err := repo.WithRTxVal(func(dao Repository) (interface{}, error) {
		config, err := dao.FindTlsConfigById(id)
		assert.Nil(t, err)
		return config, err
	})
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test", config.(*domain.TlsConfig).Name)
}

func TestInMemRepo_FindTlsConfigByName(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	createAndSaveTlsConfig(t, repo, "test")
	config, err := repo.WithRTxVal(func(dao Repository) (interface{}, error) {
		config, err := dao.FindTlsConfigByName("test")
		assert.Nil(t, err)
		return config, err
	})
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test", config.(*domain.TlsConfig).Name)
}

func TestInMemRepo_SaveTlsConfig(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	nodeGroups := []*domain.NodeGroup{
		{
			Name: "nodeGroup1",
		},
		{
			Name: "nodeGroup2",
		},
	}
	saveNodeGroup(t, repo, nodeGroups[0])
	saveNodeGroup(t, repo, nodeGroups[1])

	tlsConfig := createTlsConfig("test", nodeGroups)
	tlsConfig.Enabled = true
	tlsConfig.Insecure = true
	tlsConfig.SNI = "sni"
	tlsConfig.TrustedCA = "test-ca"
	tlsConfig.ClientCert = "test-client-cert"
	tlsConfig.PrivateKey = "private-key"
	id := saveTlsConfig(t, repo, tlsConfig)

	config := getTlsConfigById(t, repo, id)
	assert.NotNil(t, config)
	assert.Equal(t, "test", config.Name)
	assert.True(t, config.Enabled)
	assert.True(t, config.Insecure)
	assert.Equal(t, "sni", config.SNI)
	assert.Equal(t, "test-ca", config.TrustedCA)
	assert.Equal(t, "test-client-cert", config.ClientCert)
	assert.Equal(t, "private-key", config.PrivateKey)

	tlsConfig.ClientCert = "test-client-cert-update"
	id = saveTlsConfig(t, repo, tlsConfig)
	assert.Equal(t, "test-client-cert-update", config.ClientCert)

	config = getTlsConfigById(t, repo, id)
	assert.Equal(t, "nodeGroup1", config.NodeGroups[0].Name)
	assert.Equal(t, "nodeGroup2", config.NodeGroups[1].Name)
}

func TestInMemRepo_TlsConfig_Save_Backup_Resotre(t *testing.T) {
	repo := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	nodeGroups := []*domain.NodeGroup{
		{
			Name: "node_group_SBR",
		},
	}
	saveNodeGroup(t, repo, nodeGroups[0])

	tlsConfig := createTlsConfig("tls_config_SBR", nodeGroups)
	tlsConfig.Enabled = true
	tlsConfig.Insecure = true
	tlsConfig.SNI = "sni"
	tlsConfig.TrustedCA = "test-ca"
	tlsConfig.ClientCert = "test-client-cert"
	tlsConfig.PrivateKey = "private-key"
	id := saveTlsConfig(t, repo, tlsConfig)

	config := getTlsConfigById(t, repo, id)
	assert.NotNil(t, config)
	assert.Equal(t, "tls_config_SBR", config.Name)
	assert.True(t, config.Enabled)
	assert.True(t, config.Insecure)
	assert.Equal(t, "sni", config.SNI)
	assert.Equal(t, "test-ca", config.TrustedCA)
	assert.Equal(t, "test-client-cert", config.ClientCert)
	assert.Equal(t, "private-key", config.PrivateKey)
	assert.Equal(t, "node_group_SBR", config.NodeGroups[0].Name)

	tlsConfigsByNodeGroups, err := repo.FindAllTlsConfigsByNodeGroup("node_group_SBR")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigsByNodeGroups))

	snapshot, err := repo.storage.Backup()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(snapshot.TlsConfigsNodeGroups))
	assert.Equal(t, "node_group_SBR", snapshot.TlsConfigsNodeGroups[0].NodeGroupName)
	assert.Equal(t, id, snapshot.TlsConfigsNodeGroups[0].TlsConfigId)

	repo.storage.Clear()
	configAfterClear := getTlsConfigById(t, repo, id)
	assert.Nil(t, configAfterClear)
	tlsConfigsByNodeGroupsAfterClear, err := repo.FindAllTlsConfigsByNodeGroup("node_group_SBR")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(tlsConfigsByNodeGroupsAfterClear))

	repo.storage.Restore(*snapshot)
	configAfterRestore := getTlsConfigById(t, repo, id)
	assert.NotNil(t, configAfterRestore)
	assert.Equal(t, "tls_config_SBR", configAfterRestore.Name)
	tlsConfigsByNodeGroupsAfterRestore, err := repo.FindAllTlsConfigsByNodeGroup("node_group_SBR")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(tlsConfigsByNodeGroupsAfterRestore))
	assert.Equal(t, id, tlsConfigsByNodeGroupsAfterRestore[0].Id)
}

func createAndSaveTlsConfig(t *testing.T, repo *InMemRepo, name string) int32 {
	return saveTlsConfig(t, repo, createTlsConfig(name, nil))
}

func saveTlsConfig(t *testing.T, repo *InMemRepo, tls *domain.TlsConfig) int32 {
	id, _, err := repo.WithWTxVal(func(dao Repository) (interface{}, error) {
		err := dao.SaveTlsConfig(tls)
		assert.Nil(t, err)
		return tls.Id, nil
	})
	assert.Nil(t, err)
	return id.(int32)
}

func saveNodeGroup(t *testing.T, repo *InMemRepo, nodeGroup *domain.NodeGroup) string {
	name, _, err := repo.WithWTxVal(func(dao Repository) (interface{}, error) {
		err := dao.SaveNodeGroup(nodeGroup)
		assert.Nil(t, err)
		return nodeGroup.Name, nil
	})
	assert.Nil(t, err)
	return name.(string)
}

func createTlsConfig(name string, nodeGroups []*domain.NodeGroup) *domain.TlsConfig {
	return &domain.TlsConfig{
		Name:       name,
		NodeGroups: nodeGroups,
		Enabled:    true,
		Insecure:   true,
		TrustedCA:  name + "test-TrustedCA",
		ClientCert: name + "test-ClientCert",
		PrivateKey: name + "test-PrivateKey",
		SNI:        name + "test-SNI",
	}
}

func getTlsConfigById(t *testing.T, repo *InMemRepo, id int32) *domain.TlsConfig {
	tls, err := repo.WithRTxVal(func(dao Repository) (interface{}, error) {
		tlsConfig, err := dao.FindTlsConfigById(id)
		assert.Nil(t, err)
		return tlsConfig, nil
	})
	assert.Nil(t, err)
	return tls.(*domain.TlsConfig)
}

func deleteTestTlsConfigById(t *testing.T, repo *InMemRepo, id int32) {
	_, err := repo.WithWTx(func(dao Repository) error {
		err := dao.DeleteTlsConfigById(id)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)
}
