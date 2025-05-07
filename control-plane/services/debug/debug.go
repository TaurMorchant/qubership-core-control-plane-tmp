package debug

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/composite"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/data"
)

type Service struct {
	storage          *dao.InMemDao
	compositeService *composite.Service
}

func NewService(storage *dao.InMemDao, compositeService *composite.Service) *Service {
	return &Service{
		storage:          storage,
		compositeService: compositeService,
	}
}

func (s *Service) DumpDataSnapshot() (*data.Snapshot, error) {
	return s.storage.Backup()
}

func (s *Service) ValidateConfig() (*StatusConfig, error) {
	problem, err := ValidateConfig(s.storage, s.compositeService)
	if err != nil {
		logger.Errorf("Failed to Validate Config: %v", err)
		return nil, err
	}
	return problem, nil
}
