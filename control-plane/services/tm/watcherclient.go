package tm

import (
	"encoding/json"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/tm/entity"
	go_stomp_websocket "github.com/netcracker/qubership-core-lib-go-stomp-websocket/v3"
)

type TenantWatcherClient struct {
	cache                 *cache.CacheClient
	updateTenantNamespace *TenantNamespaceUpdater
	cancel                chan bool
	active                bool
}

var watcherClient *TenantWatcherClient

func init() {
	watcherClient = &TenantWatcherClient{
		cache:  cache.NewCacheClient(),
		cancel: make(chan bool),
	}
}

func (client *TenantWatcherClient) Watch(sub *go_stomp_websocket.Subscription) {
	logger.InfoC(ctx, "Starting to receive messages from tenant-manager")
	client.active = true
	client.handleFrames(sub)
	logger.InfoC(ctx, "Reading from socket is stopped")
}

func (client *TenantWatcherClient) StopWatching() {
	if client.active {
		logger.Infof("Stopping tenant watching")
		client.cancel <- true
	}
}

func (client *TenantWatcherClient) handleFrames(sub *go_stomp_websocket.Subscription) {
	for {
		logger.InfoC(ctx, "Waiting for frame from tenant-manager")
		select {
		case frame, ok := <-sub.FrameCh: // receive frame
			if ok && len(frame.Body) > 0 {
				logger.InfoC(ctx, "Message from tenant-manager is: \n %v", frame.Body)
				err := client.processFrame(frame)
				if err != nil {
					logger.ErrorC(ctx, "Can't process frame %v", err)
				}
			}
			if !ok {
				logger.ErrorC(ctx, "Error during reading from channel. Connection to tenant-manager socket is closed")
				return
			}
		case <-client.cancel:
			logger.Debug("Tenant watching has been stopped")
			client.active = false
			return
		}
	}
}

func (client *TenantWatcherClient) processFrame(frame *go_stomp_websocket.Frame) error {
	var tenants = new(entity.WatchApiTenant)
	err := json.Unmarshal([]byte(frame.Body), tenants)
	if err != nil {
		logger.ErrorC(ctx, "Can't unmarshal json %v", err)
		return err
	}

	isUpdated := false
	for _, tenant := range tenants.Tenants {
		logger.InfoC(ctx, "Received tenant with id %s", tenant.ExternalId)
		if client.tenantsNamespacesUpdate(tenant) {
			logger.InfoC(ctx, "Tenant status is changed to %s", tenant.Status)
			isUpdated = true
		}
	}
	if isUpdated {
		logger.InfoC(ctx, "One of the tenant is changed its status. Updating listeners")
		tenantsToUpdate := client.cache.GetAll()
		err := client.updateTenantNamespace.UpdateAllListeners(tenantsToUpdate)
		if err != nil {
			logger.InfoC(ctx, "Can't update listeners with tenants %v \n %v", tenantsToUpdate, err)
			return err
		}
	}
	logger.InfoC(ctx, "Status of the tenants are not changed. Listeners updating is skipped")
	return nil
}

func (client *TenantWatcherClient) tenantsNamespacesUpdate(tenant *entity.Tenant) bool {
	if tenant.ExternalId == "" {
		logger.InfoC(ctx, "Tenant with id: %s is not activated", tenant.ObjectId)
		return false
	}
	if tenant.Status == "" {
		return false
	}
	if tenant.Status == entity.TenantActive {
		if tenant.Namespace != "" {
			logger.InfoC(ctx, "Put tenant %s to namespaces cache", tenant.ExternalId)
			client.cache.Set(tenant.ExternalId, tenant.Namespace)
			return true
		}
	} else {
		_, isPresented := client.cache.Get(tenant.ExternalId)
		if isPresented {
			client.cache.Delete(tenant.ExternalId)
			logger.InfoC(ctx, "Delete tenant %s from namespaces cache", tenant.ExternalId)
			return true
		}
	}
	return false
}
