package entity

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/cookie"
	"strings"
)

func (srv *Service) PutStatefulSession(dao dao.Repository, newSession *domain.StatefulSession) error {
	if newSession.Enabled {
		if newSession.CookieName == "" {
			newSession.CookieName = cookie.NameGenerator.GenerateCookieName("sticky")
		} else {
			suffix := fmt.Sprintf("-%s", newSession.InitialDeploymentVersion)
			if !strings.HasSuffix(newSession.CookieName, suffix) {
				newSession.CookieName += suffix
			}
		}
	}
	if err := dao.SaveStatefulSessionConfig(newSession); err != nil {
		logger.Errorf("Failed to save stateful session using DAO:\n %v", err)
		return err
	}
	return nil
}
