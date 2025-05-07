// Package tm provides modules for work with tenant-manager service.
// This package is part of tenant-based routing functional.
package tm

import (
	"context"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/clustering"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/tlsmode"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"time"
)

const (
	WebsocketPath  = "/api/v4/tenant-manager/watch"
	WebsocketTopic = "/channels/tenants"
)

var (
	logger logging.Logger
	ctx    = context.Background()
)

func init() {
	logger = logging.GetLogger("tenant-based-routing")
}

type Watcher struct {
	watcherClient       *TenantWatcherClient
	socketClient        *SocketClient
	active              bool
	internalGatewayAddr string
}

func NewWatcher(storage dao.Dao, updateManager *cache.UpdateManager) *Watcher {
	watcherClient.updateTenantNamespace = NewTenantNamespaceUpdater(storage, updateManager)
	tmApiUrl := tlsmode.UrlFromProperty(tlsmode.Websocket, "tenant.manager.api.url", domain.InternalGateway)
	return &Watcher{
		watcherClient:       watcherClient,
		socketClient:        NewSocketClient(),
		internalGatewayAddr: tmApiUrl,
	}
}

func (w *Watcher) UpAndStartWatchSocket(_ clustering.NodeInfo, state clustering.Role) error {
	switch state {
	case clustering.Master:
		w.active = true
		go func() {
			for w.active {
				sub, err := w.socketClient.ConnectAndSubscribe(w.internalGatewayAddr+WebsocketPath, WebsocketTopic)
				if err != nil {
					logger.Errorf("Can't connect to tenant manager socket %v", err)
					time.Sleep(30 * time.Second)
					continue
				}
				logger.Infof("Starting tenant manager watcher")

				w.watcherClient.Watch(sub)

				logger.Infof("Stopping tenant manager watcher. Watcher will be restarted")
			}
		}()
	case clustering.Slave:
		w.watcherClient.StopWatching()
		w.active = false
	case clustering.Phantom:
	default:
		logger.Errorf("Unexpected role %v", state)
	}

	return nil
}
