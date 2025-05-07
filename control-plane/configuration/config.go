package config

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/constancy"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"sync"
)

type InMemoryStorageConfigurator struct {
	constantStorage constancy.Storage
	storage         ram.RamStorage
	loader          *ram.StorageLoader
	flusher         *constancy.Flusher
	dao             *dao.InMemDao
	onceDao         sync.Once
}

func NewInMemoryStorageConfigurator(constantStorage constancy.Storage, batchTm constancy.BatchTransactionManager) *InMemoryStorageConfigurator {
	inMemStorage := ram.NewStorage()
	loader := &ram.StorageLoader{PersistentStorage: constantStorage}
	podStateManager := &constancy.PodStateManagerImpl{Storage: constantStorage}
	flusher := &constancy.Flusher{BatchTm: batchTm, PodStateManager: podStateManager}
	return &InMemoryStorageConfigurator{
		storage:         inMemStorage,
		constantStorage: constantStorage,
		loader:          loader,
		flusher:         flusher,
	}
}

func (c *InMemoryStorageConfigurator) GetDao() *dao.InMemDao {
	c.onceDao.Do(func() {
		c.dao = dao.NewInMemDao(c.storage, c.constantStorage, c.defaultBeforeCommitCallbacks())
	})
	return c.dao
}

func (c *InMemoryStorageConfigurator) SyncInMemoryWithPersistentStorage() error {
	if err := c.loader.ClearAndLoad(c.storage); err != nil {
		return err
	}
	return nil
}

func (c *InMemoryStorageConfigurator) defaultBeforeCommitCallbacks() []func([]memdb.Change) error {
	return []func([]memdb.Change) error{c.flushChangesToPersistenceStorage}
}

func (c *InMemoryStorageConfigurator) flushChangesToPersistenceStorage(changes []memdb.Change) error {
	return c.flusher.Flush(changes)
}
