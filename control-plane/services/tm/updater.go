package tm

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	envoy "github.com/netcracker/qubership-core-control-plane/control-plane/v2/envoy/cache"
	"sort"
	"time"
)

type TenantNamespaceUpdater struct {
	storage      dao.Dao
	cacheManager *envoy.UpdateManager
}

func NewTenantNamespaceUpdater(storage dao.Dao, cacheManager *envoy.UpdateManager) *TenantNamespaceUpdater {
	return &TenantNamespaceUpdater{
		storage:      storage,
		cacheManager: cacheManager,
	}
}

func (updater *TenantNamespaceUpdater) UpdateAllListeners(rawTenantsToUpdate map[string]interface{}) error {
	logger.InfoC(ctx, "Starting to update listeners")
	tenantsToUpdate := updater.convertTenantsToUpdate(rawTenantsToUpdate)
	err := updater.updateAllListenersWithTenants(tenantsToUpdate)
	if err != nil {
		logger.ErrorC(ctx, "Can't update listeners \n%v", err)
		return err
	}
	logger.InfoC(ctx, "Listeners are updated")
	return nil
}

func (updater *TenantNamespaceUpdater) convertTenantsToUpdate(tenantsToUpdate map[string]interface{}) map[string]string {
	convertedMap := make(map[string]string)
	for key, value := range tenantsToUpdate {
		convertedMap[key] = value.(string)
	}
	return convertedMap
}

func (updater *TenantNamespaceUpdater) updateAllListenersWithTenants(tenantsToUpdate map[string]string) error {
	logger.InfoC(ctx, "Get all listeners")
	listeners, err := updater.storage.FindAllListeners()
	if err != nil {
		logger.ErrorC(ctx, "Can't get listeners %v", err)
		return err
	}
	if len(tenantsToUpdate) == 0 {
		logger.InfoC(ctx, "List of tenants is empty")
		updater.UpdateListeners(listeners, "")
		return nil
	}
	logger.InfoC(ctx, "List of the tenants to update is %v", tenantsToUpdate)
	namespaceMapping := updater.convertToNamespaceMappings(tenantsToUpdate)
	logger.InfoC(ctx, "List of the namespace mappings is %s", namespaceMapping)
	return updater.UpdateListeners(listeners, namespaceMapping)
}

// Function converts map to string in format 'tenantsNS = {["t_id"] = "ns", ["t_id2"] = "ns2"}'
// Map contains information about tenants and namespaces
func (updater *TenantNamespaceUpdater) convertToNamespaceMappings(tenantsToUpdate map[string]string) string {
	keys := make([]string, 0)
	for tenantId, _ := range tenantsToUpdate {
		keys = append(keys, tenantId)
	}
	sort.Strings(keys)
	namespaceMapping := "tenantsNS = {"
	for _, tenantId := range keys {
		tenantNamespace := tenantsToUpdate[tenantId]
		namespaceMapping += "[\"" + tenantId + "\"]"
		namespaceMapping += " = "
		namespaceMapping += "\"" + tenantNamespace + "\""
		namespaceMapping += ", "
	}
	namespaceMapping = namespaceMapping[:len(namespaceMapping)-2]
	namespaceMapping += "}"
	return namespaceMapping
}

func (updater *TenantNamespaceUpdater) UpdateListeners(listeners []*domain.Listener, namespaceMapping string) error {
	logger.InfoC(ctx, "Starting to update listeners")
	for _, listener := range listeners {
		logger.InfoC(ctx, "Updating listener %v", listener)
		version := time.Now().UnixNano()
		err := updater.cacheManager.UpdateListener(listener.NodeGroupId, version, listener, namespaceMapping)
		if err != nil {
			logger.ErrorC(ctx, "Can't update listener %v", err)
			return err
		}
		logger.InfoC(ctx, "Listener is updated successfully %v", listener)
	}
	logger.InfoC(ctx, "All listeners are updated successfully")
	return nil
}
