package dao

import (
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var (
	logger logging.Logger
)

func init() {
	logger = logging.GetLogger("tlsConfig")
}

func (d *InMemRepo) SaveTlsConfig(tls *domain.TlsConfig) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()

	tlsConfigFromDb, err := d.FindTlsConfigByName(tls.Name)
	if err != nil {
		return err
	}

	if tls.Id == 0 {
		if tlsConfigFromDb == nil {
			if err := d.idGenerator.Generate(tls); err != nil {
				return err
			}
		} else {
			tls.Id = tlsConfigFromDb.Id
		}
	}

	err = d.storage.Save(txCtx.tx, domain.TlsConfigTable, tls)
	if err != nil {
		return err
	}

	// update nodeGroups many-to-many table if needed
	// deleting outdated tlsConfig-nodeGroup bindings
	if tlsConfigFromDb != nil {
		for _, group := range tlsConfigFromDb.NodeGroups {
			if !contains(tls.NodeGroups, group) { // if already exists
				logger.Info("Deleting TlsConfig %s to NodeGroup %s binding", tls.Name, group)
				if err := d.DeleteTlsConfigByIdAndNodeGroupName(&domain.TlsConfigsNodeGroups{
					TlsConfigId:   tlsConfigFromDb.Id,
					NodeGroupName: group.Name,
				}); err != nil {
					logger.Error("Failed to delete TlsConfig %s to NodeGroup %s binding:\n %v", tls.Name, group, err)
					return err
				}
			}
		}
	}
	// saving actual tlsConfig-nodeGroup bindings if required
	for _, group := range tls.NodeGroups {
		if tlsConfigFromDb != nil && contains(tlsConfigFromDb.NodeGroups, group) { // if already exists
			logger.Info("TlsConfig %s is already bound to node group %s", tls.Name, group)
			continue
		}
		logger.Info("Binding TlsConfig %s to node group %s", tls.Name, group)
		err = d.storage.Save(txCtx.tx, domain.TlsConfigsNodeGroupsTable, &domain.TlsConfigsNodeGroups{TlsConfigId: tls.Id, NodeGroupName: group.Name})
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *InMemRepo) FindTlsConfigByName(name string) (*domain.TlsConfig, error) {
	tlsConfig, err := FindFirstByIndex[domain.TlsConfig](d, domain.TlsConfigTable, "name", name)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		logger.Debugf("Found TlsConfig with Id %d", tlsConfig.Id)
		tlsConfigsNodeGroups, err := FindByIndex[domain.TlsConfigsNodeGroups](d, domain.TlsConfigsNodeGroupsTable, "tlsConfigId", tlsConfig.Id)
		if err != nil {
			return nil, err
		} else {
			tlsConfig.NodeGroups = make([]*domain.NodeGroup, len(tlsConfigsNodeGroups))
			for i, tlsConfigNodeGroup := range tlsConfigsNodeGroups {
				tlsConfig.NodeGroups[i], err = d.FindNodeGroupByName(tlsConfigNodeGroup.NodeGroupName)
				if err != nil {
					return nil, err
				}
			}
			return tlsConfig, nil
		}
	}
	return nil, nil
}

func (d *InMemRepo) FindTlsConfigById(id int32) (*domain.TlsConfig, error) {
	return FindById[domain.TlsConfig](d, domain.TlsConfigTable, id)
}

func (d *InMemRepo) FindAllTlsConfigsByNodeGroup(nodeGroup string) ([]*domain.TlsConfig, error) {
	txCtx := d.getTxCtx(false)
	defer txCtx.closeIfLocal()
	result, err := d.storage.FindByIndex(txCtx.tx, domain.TlsConfigsNodeGroupsTable, "nodeGroupName", nodeGroup)
	if err != nil {
		return nil, errors.WrapPrefix(err, "can not get tlsConfig node groups", 1)
	}
	tlsConfigNodeGroups := result.([]*domain.TlsConfigsNodeGroups)
	tlsConfigs := make([]*domain.TlsConfig, len(tlsConfigNodeGroups))
	for i, tlsConfigNodeGroup := range tlsConfigNodeGroups {
		tlsConfig, err := d.FindTlsConfigById(tlsConfigNodeGroup.TlsConfigId)
		if err != nil {
			return nil, err
		}
		tlsConfigs[i] = tlsConfig
	}
	return tlsConfigs, nil
}

func (d *InMemRepo) FindAllTlsConfigs() ([]*domain.TlsConfig, error) {
	return FindAll[domain.TlsConfig](d, domain.TlsConfigTable)
}

func (d *InMemRepo) DeleteTlsConfigById(id int32) error {
	return d.DeleteById(domain.TlsConfigTable, id)
}

func contains(nodeGroups []*domain.NodeGroup, nodeGroup *domain.NodeGroup) bool {
	for _, group := range nodeGroups {
		if group.Name == nodeGroup.Name {
			logger.Debugf("TlsConfig contains nodeGroup %s", nodeGroup.Name)
			return true
		}
	}
	return false
}

func (d *InMemRepo) DeleteTlsConfigByIdAndNodeGroupName(relation *domain.TlsConfigsNodeGroups) error {
	return d.Delete(domain.TlsConfigsNodeGroupsTable, relation)
}
