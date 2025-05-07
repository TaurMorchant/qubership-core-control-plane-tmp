package dao

import "github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"

func (d *InMemRepo) SaveTcpKeepalive(entity *domain.TcpKeepalive) error {
	txCtx := d.getTxCtx(true)
	defer txCtx.closeIfLocal()
	if err := d.idGenerator.Generate(entity); err != nil {
		return err
	}
	err := d.storage.Save(txCtx.tx, domain.TcpKeepaliveTable, entity)
	if err != nil {
		return err
	}
	return nil
}

func (d *InMemRepo) FindAllTcpKeepalives() ([]*domain.TcpKeepalive, error) {
	return FindAll[domain.TcpKeepalive](d, domain.TcpKeepaliveTable)
}

func (d *InMemRepo) FindTcpKeepaliveById(id int32) (*domain.TcpKeepalive, error) {
	return FindById[domain.TcpKeepalive](d, domain.TcpKeepaliveTable, id)
}

func (d *InMemRepo) DeleteTcpKeepaliveById(id int32) error {
	return d.DeleteById(domain.TcpKeepaliveTable, id)
}
