package tls

import (
	"context"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/cert"
	"github.com/netcracker/qubership-core-control-plane/dao"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/event/bus"
	"github.com/netcracker/qubership-core-control-plane/event/events"
	"github.com/netcracker/qubership-core-control-plane/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/ui"
	"github.com/netcracker/qubership-core-control-plane/util"
	tlsUtil "github.com/netcracker/qubership-core-control-plane/util/tls"
	"strings"

	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var logger logging.Logger

func init() {
	logger = logging.GetLogger("tls-service")
}

type Service struct {
	dao                dao.Dao
	bus                bus.BusPublisher
	certificateManager *cert.CertificateManager
}

type void struct{}

var member void

func NewTlsService(dao dao.Dao, bus bus.BusPublisher, certificateManager *cert.CertificateManager) *Service {
	return &Service{
		dao:                dao,
		bus:                bus,
		certificateManager: certificateManager,
	}
}

func (s *Service) deleteTlsConfig(ctx context.Context, tls *dto.TlsConfig) error {
	logger.InfoC(ctx, "Deleting tls configuration by name %s", tls.Name)
	nodeGroupsToUpdate := make([]string, 0)
	changes, err := s.dao.WithWTx(func(repo dao.Repository) error {
		existing, err := repo.FindTlsConfigByName(tls.Name)
		if err != nil {
			logger.ErrorC(ctx, "Failed to load TLS config by name %s to delete:\n %v", tls.Name, err)
			return err
		}
		if existing == nil {
			logger.InfoC(ctx, "There is no tls configuration for name %s, nothing to delete", tls.Name)
			return nil
		}
		for _, nodeGroup := range existing.NodeGroups {
			err = repo.DeleteTlsConfigByIdAndNodeGroupName(&domain.TlsConfigsNodeGroups{TlsConfigId: existing.Id, NodeGroupName: nodeGroup.Name})
			if err != nil {
				logger.ErrorC(ctx, "Failed to delete TLS config %s nodeGroup %s reference:\n %v", tls.Name, nodeGroup.Name, err)
				return err
			}
			nodeGroupsToUpdate = append(nodeGroupsToUpdate, nodeGroup.Name)
		}

		if err = repo.DeleteTlsConfigById(existing.Id); err != nil {
			logger.ErrorC(ctx, "Failed to delete TLS config %s:\n %v", tls.Name, err)
			return err
		}

		return s.updateEnvoyConfigVersionByNodeGroups(ctx, repo, nodeGroupsToUpdate)
	})
	if err != nil {
		return err
	}

	for _, nodeGroup := range nodeGroupsToUpdate {
		event := events.NewChangeEventByNodeGroup(nodeGroup, changes)
		logger.InfoC(ctx, "Publishing changes for node group %s: %s", nodeGroup, event.ToString())
		err = s.bus.Publish(bus.TopicChanges, event)
		if err != nil {
			logger.Errorf("can not publish changes for TLS with nodeGroupId=%s, %v", nodeGroup, err)
			return err
		}
	}

	logger.InfoC(ctx, "Done deleting tls configuration %s", tls.Name)
	return nil
}

func (s *Service) SaveTlsConfig(ctx context.Context, tls *dto.TlsConfig) error {
	logger.InfoC(ctx, "Applying tls configuration %+v", *tls)

	if tls.Tls == nil {
		return s.deleteTlsConfig(ctx, tls)
	}

	if tls.Tls.Enabled && !tls.Tls.Insecure {
		if tls.Tls.TrustedCA == "" {
			return errors.New("TlsDef must be insecure or must contain non-empty trustedCA")
		}
		if (tls.Tls.ClientCert != "" && tls.Tls.PrivateKey == "") || (tls.Tls.PrivateKey != "" && tls.Tls.ClientCert == "") {
			return errors.New("TlsDef for mTLS must contain both non-empty clientCert and privateKey")
		}
	}

	tls.Tls.TrustedCA = strings.TrimSpace(tls.Tls.TrustedCA)
	tls.Tls.ClientCert = strings.TrimSpace(tls.Tls.ClientCert)
	tls.Tls.PrivateKey = strings.TrimSpace(tls.Tls.PrivateKey)

	if err := tlsUtil.GetTlsWithCertificatesValidation(tls.Tls.TrustedCA, tls.Tls.ClientCert, tls.Tls.PrivateKey); err != nil {
		return err
	}

	nodeGroupsToUpdate := make([]string, 0)
	nodeGroupsToUpdate = append(nodeGroupsToUpdate, tls.TrustedForGateways...)
	if len(nodeGroupsToUpdate) > 0 {
		logger.InfoC(ctx, "Tls config %s is trusted for gateways: %+q", tls.Name, nodeGroupsToUpdate)
	}
	changes, err := s.dao.WithWTx(func(dao dao.Repository) error {
		err := s.createNodeGroupsIfNotExists(ctx, dao, tls.TrustedForGateways)
		if err != nil {
			return err
		}

		savedTlsConfig, err := s.createOrSaveTlsConfig(ctx, dao, tls)
		if err != nil {
			return err
		}

		nodeGroupsForTlsConfig, err := s.getAllNodeGroupsForTlsConfig(ctx, dao, savedTlsConfig)
		if err != nil {
			return err
		}
		nodeGroupsToUpdate = append(nodeGroupsToUpdate, nodeGroupsForTlsConfig...)

		return s.updateEnvoyConfigVersionByNodeGroups(ctx, dao, nodeGroupsToUpdate)
	})

	for _, nodeGroup := range nodeGroupsToUpdate {
		event := events.NewChangeEventByNodeGroup(nodeGroup, changes)
		logger.InfoC(ctx, "Publishing changes for node group %s: %s", nodeGroup, event.ToString())
		err = s.bus.Publish(bus.TopicChanges, event)
		if err != nil {
			logger.Errorf("can not publish changes for TLS with nodeGroupId=%s, %v", nodeGroup, err)
			return err
		}
	}

	updateCertificateMetrics(s)
	logger.InfoC(ctx, "Done registering tls configuration %+v", *tls)
	return err
}

func (s *Service) createNodeGroupsIfNotExists(ctx context.Context, dao dao.Repository, gateways []string) error {
	for _, nodeGroupId := range gateways {
		nodeGroup, err := dao.FindNodeGroupByName(nodeGroupId)
		if err != nil {
			logger.ErrorC(ctx, "Problem during getting node group with id %v: %v", nodeGroupId, err)
			return err
		}
		if nodeGroup == nil {
			newNodeGroup := domain.NodeGroup{Name: nodeGroupId}
			err := dao.SaveNodeGroup(&newNodeGroup)
			if err != nil {
				logger.ErrorC(ctx, "Cannot create Node Group with id %v: %v", nodeGroupId, err)
				return err
			}
		}
	}
	return nil
}

func (s *Service) GetGlobalTlsConfigs(cluster *domain.Cluster, affectedNodeGroups ...string) ([]*domain.TlsConfig, error) {
	tlsConfigs := make(map[*domain.TlsConfig]void)
	nodeGroups, err := s.dao.FindNodeGroupsByCluster(cluster)
	if err != nil {
		return nil, err
	}
	for _, nodeGroup := range nodeGroups {
		if len(affectedNodeGroups) != 0 && !util.SliceContains(affectedNodeGroups, nodeGroup.Name) {
			logger.Infof("Cluster %s node group %s is ignored by current update action", cluster.Name, nodeGroup.Name)
			continue
		}
		tlsConfigsByNodeGroup, err := s.dao.FindAllTlsConfigsByNodeGroup(nodeGroup.Name)
		if err != nil {
			return nil, err
		}
		for _, tc := range tlsConfigsByNodeGroup {
			logger.Info("Found gateway-wide TLS for cluster %s and node group %s: %+v", cluster.Name, nodeGroup, *tc)
			tlsConfigs[tc] = member
		}
	}
	tlsConfigsSlice := make([]*domain.TlsConfig, 0, len(tlsConfigs))

	for config := range tlsConfigs {
		tlsConfigsSlice = append(tlsConfigsSlice, config)
	}
	return tlsConfigsSlice, nil
}

func (s *Service) GetTlsDefResource() cfgres.Resource {
	return &tlsDefResource{service: s}
}

func (s *Service) createOrSaveTlsConfig(ctx context.Context, dao dao.Repository, tls *dto.TlsConfig) (*domain.TlsConfig, error) {
	foundTlsConfig, err := dao.FindTlsConfigByName(tls.Name)
	if err != nil {
		return nil, err
	}
	tlsConfigToSave := dto.ConvertTLSToDomain(tls)
	if foundTlsConfig != nil {
		tlsConfigToSave.Id = foundTlsConfig.Id
	}
	logger.InfoC(ctx, "Built domain tls config: %+v", *tlsConfigToSave)
	err = dao.SaveTlsConfig(tlsConfigToSave)
	if err != nil {
		return nil, err
	}
	return tlsConfigToSave, nil
}

func (s *Service) getAllNodeGroupsForTlsConfig(ctx context.Context, dao dao.Repository, tlsConfig *domain.TlsConfig) ([]string, error) {
	nodeGroups := []string{}
	// find related clusters and their node groups to update envoy cache
	clusters, err := dao.FindAllClusters()
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		if cluster.TLSId == tlsConfig.Id {
			logger.InfoC(ctx, "Tls config %s is bound to cluster: %s", tlsConfig.Name, cluster.Name)
			ngs, err := dao.FindNodeGroupsByCluster(cluster)
			if err != nil {
				return nil, err
			}
			for _, ng := range ngs {
				nodeGroups = append(nodeGroups, ng.Name)
			}
		}
	}

	return nodeGroups, nil
}

func (s *Service) updateEnvoyConfigVersionByNodeGroups(ctx context.Context, dao dao.Repository, nodeGroupsToUpdate []string) error {
	for _, nodeGroupId := range nodeGroupsToUpdate {
		logger.InfoC(ctx, "Updating version for node group %s", nodeGroupId)
		if err := dao.SaveEnvoyConfigVersion(domain.NewEnvoyConfigVersion(nodeGroupId, domain.ClusterTable)); err != nil {
			logger.ErrorC(ctx, "add TLS config failed due to error in envoy config version saving for cluster: %v", err)
			return err
		}
	}

	return nil
}

func (s *Service) ValidateCertificates() (*ui.CertificateDetailsResponse, error) {
	tlsConfigs, err := s.dao.FindAllTlsConfigs()
	if err != nil {
		return nil, fmt.Errorf("can't get tls configs: %+v", err)
	}
	response := &ui.CertificateDetailsResponse{
		TlsDefDetails: []*ui.TlsDefDetails{},
	}
	for _, tlsConfig := range tlsConfigs {
		validationResult, err := s.certificateManager.VerifyCert(tlsConfig.TrustedCA)
		if err == nil {
			tlsDefDetails := &ui.TlsDefDetails{
				Name: tlsConfig.Name,
			}
			response.TlsDefDetails = append(response.TlsDefDetails, tlsDefDetails)
			for _, nodeGroup := range tlsConfig.NodeGroups {
				usedIn := &ui.CertificateUsedIn{
					Clusters: make([]*ui.CertificateUsedInCluster, 0),
					Gateway:  nodeGroup.Name,
				}
				for _, cluster := range nodeGroup.Clusters {
					usedInCluster := &ui.CertificateUsedInCluster{
						Name:      cluster.Name,
						Endpoints: make([]string, 0),
					}
					for _, endpoint := range cluster.Endpoints {
						usedInCluster.Endpoints = append(usedInCluster.Endpoints, endpoint.String())
					}
					usedIn.Clusters = append(usedIn.Clusters, usedInCluster)
				}
				tlsDefDetails.UsedIn = append(tlsDefDetails.UsedIn, usedIn)
			}
			for _, certValidationresult := range validationResult.CertificateValidationResult {
				cerificateDetails := &ui.CertificateDetails{
					Reason:                      certValidationresult.Reason,
					Valid:                       certValidationresult.Valid,
					ValidFrom:                   certValidationresult.CertificateInfo.ValidFrom,
					ValidTill:                   certValidationresult.CertificateInfo.ValidTill,
					DaysTillExpiry:              certValidationresult.CertificateInfo.DaysTillExpiry,
					SANs:                        certValidationresult.CertificateInfo.SANs,
					CertificateCommonName:       certValidationresult.CertificateInfo.Name,
					CertificateId:               certValidationresult.CertificateInfo.Id,
					IssuerCertificateCommonName: certValidationresult.CertificateInfo.IssuerName,
					IssuerCertificateId:         certValidationresult.CertificateInfo.IssuerId,
				}
				tlsDefDetails.Certificates = append(tlsDefDetails.Certificates, cerificateDetails)
			}
		} else {
			tlsDef := &ui.TlsDefDetails{
				Name:         tlsConfig.Name,
				Certificates: make([]*ui.CertificateDetails, 0),
			}
			cerificateDdetails := &ui.CertificateDetails{
				Reason: fmt.Sprintf("can't validate cert: %+v", err),
				Valid:  false,
			}
			tlsDef.Certificates = append(tlsDef.Certificates, cerificateDdetails)
			response.TlsDefDetails = append(response.TlsDefDetails, tlsDef)
		}
	}
	return response, nil
}
