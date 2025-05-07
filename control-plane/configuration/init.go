package config

import (
	"github.com/netcracker/qubership-core-control-plane/composite"
	"github.com/netcracker/qubership-core-control-plane/constancy"
	"github.com/netcracker/qubership-core-control-plane/dr"
	"github.com/netcracker/qubership-core-control-plane/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/services/entity"
	"github.com/netcracker/qubership-core-control-plane/services/tls"
	"sync"
	"time"
)

type CommonMasterNodeInitializer struct {
	PersistentStorage      constancy.Storage
	RamStorageConfigurator *InMemoryStorageConfigurator
	EnvoyConfigUpdater     *cache.UpdateManager
	EntityService          *entity.Service
	CompositeService       *composite.Service
	mutex                  *sync.Mutex
	secured                bool
}

func NewCommonMasterNodeInitializer(persistentStorage constancy.Storage, ramStorageConfigurator *InMemoryStorageConfigurator, envoyConfigUpdater *cache.UpdateManager, entityService *entity.Service, compositeService *composite.Service, secured bool) *CommonMasterNodeInitializer {
	return &CommonMasterNodeInitializer{PersistentStorage: persistentStorage, RamStorageConfigurator: ramStorageConfigurator,
		EnvoyConfigUpdater: envoyConfigUpdater, EntityService: entityService, CompositeService: compositeService, mutex: &sync.Mutex{}, secured: secured}
}

func (i CommonMasterNodeInitializer) InitMaster() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	logger.Debug("init master started")

	// 1 Load from db
	if err := i.RamStorageConfigurator.SyncInMemoryWithPersistentStorage(); err != nil {
		logger.Error("init master finished with error: %v", err)
		return err
	}

	// 2 load OOB routes
	if dr.GetMode() == dr.Active {
		commonConfiguration := NewCommonConfiguration(i.RamStorageConfigurator.GetDao(), i.EntityService, i.secured)
		err := commonConfiguration.CreateCommonConfiguration()
		if err != nil {
			logger.Error("init master failed on CreateCommonConfiguration: %v", err)
			return err
		}
		logger.Debug("CreateCommonConfiguration completed")
	}
	// 3 build envoy config
	i.EnvoyConfigUpdater.InitConfigWithRetry()
	tls.TriggerCertificateMetricsUpdate()
	logger.Debug("InitConfigWithRetry completed")

	if i.CompositeService.Mode() == composite.SatelliteMode {
		return i.CompositeService.InitSatellite(5 * time.Minute)
	}

	return nil
}
