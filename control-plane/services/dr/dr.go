package dr

import (
	"encoding/json"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	"github.com/netcracker/qubership-core-control-plane/constancy"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/db"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/websocket"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"sync"
)

const NotifyChannelName = "notify_changes"

var log = logging.GetLogger("dr_service")

type Service struct {
	MasterInitializer      clustering.MasterNodeInitializer
	DBProvider             db.DBProvider
	ConstantStorage        constancy.Storage
	Dao                    dao.Dao
	Bus                    bus.BusPublisher
	ActiveActiveController *websocket.ActiveActiveController
	VersionController      *websocket.VersionController

	mutex      *sync.Mutex
	dbListener db.PersistentStorageListener
}

func (srv *Service) Start() error {
	log.Info("Starting DR service")
	srv.mutex = &sync.Mutex{}
	dbListener, err := srv.DBProvider.Listen(NotifyChannelName, srv.initStorageAfterConnect, srv.processNotification)
	if err != nil {
		log.Errorf("Error in initializing DR db listener: %v", err)
		return err
	}
	srv.dbListener = dbListener
	return nil
}

func (srv *Service) Close() {
	log.Info("Closing db listener...")
	srv.dbListener.Close()
}

func (srv *Service) initStorageAfterConnect() {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()
	log.Info("Initializing storage from DB from scratch")
	if err := srv.MasterInitializer.InitMaster(); err != nil {
		log.Panicf("Could not initialize storage due to error: %v", err)
	}
}

func (srv *Service) processNotification(payload string) {
	srv.mutex.Lock()
	defer srv.mutex.Unlock()

	log.Debugf("Got db notification on channel %s: %s", NotifyChannelName, payload)
	var event map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Panicf("Could not unmarshall db event json: %v", err)
	}
	srv.processChange(event)
}

func (srv *Service) processChange(event map[string]interface{}) {
	table, ok := event["entity"].(string)
	if !ok {
		log.Warnf("Received unexpected event: no \"entity\" field")
		return
	}
	mapper, ok := mappers[table]
	if !ok {
		log.Warnf("Received event with unsupported entity: %v", table)
		return
	}

	if event["operation"].(string) == "DELETE" {
		changes, err := srv.Dao.WithWTx(func(repo dao.Repository) error {
			return mapper.deleteFunc(event, repo)
		})
		if err != nil {
			log.Panicf("Could not delete entity %v from inMemory storage: %v", table, err)
		}
		go srv.postProcessChanges(table, changes)
	} else {
		entity, err := mapper.loadFromDBFunc(event, srv.ConstantStorage)
		if err != nil {
			log.Panicf("Could not load entity %s from constant storage: %v", table, err)
		}
		log.Infof("Loaded %s from constant storage: %v", table, entity)
		if entity != nil { // if entity == nil, then it was probably deleted by another event
			changes, err := srv.Dao.WithWTx(func(repo dao.Repository) error {
				return repo.SaveEntity(table, entity)
			})
			if err != nil {
				log.Panicf("Could not save entity %s to inMemory storage: %v", table, err)
			}
			log.Debugf("Successfully saved entity %+v to inMemory storage", entity)
			go srv.postProcessChanges(table, changes)

			if table == domain.EnvoyConfigVersionTable {
				srv.publishEvent(entity.(*domain.EnvoyConfigVersion))
			}
		}
	}
}

func (srv *Service) postProcessChanges(table string, changes []memdb.Change) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered panic in post-processing changes for entity %s %+v:\n %v", table, changes, r)
		}
	}()

	log.Infof("Post-processing changes for entity %s: %+v", table, changes)
	if table == domain.DeploymentVersionTable { // notify Deployment Versions watchers
		if err := srv.VersionController.NotifyWatchers(changes); err != nil {
			log.Errorf("Could not notify Deployment Versions watchers:\n %v", err)
		}
	}
}

func (srv *Service) publishEvent(entity *domain.EnvoyConfigVersion) {
	log.Infof("Publishing event %+v on topic %s", entity, bus.TopicPartialReapply)
	err := srv.Bus.Publish(bus.TopicPartialReapply,
		&events.PartialReloadEvent{EnvoyVersions: []*domain.EnvoyConfigVersion{entity}})
	if err != nil {
		log.Panicf("Could not publish event %v on topic %v:\n %v", bus.TopicPartialReapply, entity, err)
	}
}
