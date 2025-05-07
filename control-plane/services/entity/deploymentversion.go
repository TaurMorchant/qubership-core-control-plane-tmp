package entity

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"time"
)

func (srv *Service) getDefaultDeploymentVersion(dao dao.Repository) (*domain.DeploymentVersion, error) {
	return srv.getOrCreateDeploymentVersionInternal(dao, srv.defaultVersion, domain.ActiveStage)
}

func (srv *Service) GetDefaultVersion() string {
	return srv.defaultVersion
}

func (srv *Service) GetActiveDeploymentVersion(dao dao.Repository) (*domain.DeploymentVersion, error) {
	res, err := dao.FindDeploymentVersionsByStage(domain.ActiveStage)
	if err != nil {
		logger.Errorf("Error while loading ACTIVE deployment version from DAO: %v", err)
		return nil, err
	}
	if res == nil || len(res) == 0 {
		return srv.getDefaultDeploymentVersion(dao)
	} else {
		return res[0], nil
	}
}

func (srv *Service) GetOrCreateDeploymentVersion(dao dao.Repository, version string) (*domain.DeploymentVersion, error) {
	if version == "" {
		dv, err := srv.GetActiveDeploymentVersion(dao)
		if err != nil {
			logger.Errorf("Failed to load ACTIVE deployment version using DAO: %v", err)
		}
		return dv, err
	}
	return srv.getOrCreateDeploymentVersionInternal(dao, version, domain.CandidateStage)
}

func (srv *Service) SaveDeploymentVersion(dao dao.Repository, version *domain.DeploymentVersion) error {
	dVersion, err := dao.FindDeploymentVersion(version.Version)
	if err != nil {
		logger.Errorf("Failed to load deployment version %s using DAO: %v", version, err)
		return err
	}
	if dVersion != nil {
		err := dao.DeleteDeploymentVersion(dVersion)
		if err != nil {
			logger.Errorf("Failed to delete deployment version %s using DAO: %v", dVersion, err)
			return err
		}
	}
	version.UpdatedWhen = time.Now()
	return dao.SaveDeploymentVersion(version)
}

func (srv *Service) getOrCreateDeploymentVersionInternal(dao dao.Repository, version, stage string) (*domain.DeploymentVersion, error) {
	dv, err := dao.FindDeploymentVersion(version)
	if err != nil {
		logger.Errorf("Failed to load deployment version %v from DAO: %v", version, err)
		return nil, err
	}
	if dv == nil {
		dv = &domain.DeploymentVersion{
			Version:     version,
			Stage:       stage,
			CreatedWhen: time.Now(),
			UpdatedWhen: time.Now(),
		}
		if err := dao.SaveDeploymentVersion(dv); err != nil {
			logger.Errorf("Failed to save new deployment version %v in stage %v using DAO: %v", version, stage, err)
			return nil, err
		}
	}
	return dv, nil
}
