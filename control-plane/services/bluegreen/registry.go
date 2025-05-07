package bluegreen

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dao"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/bus"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/event/events"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/restcontrollers/dto"
	cfgres "github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/configresources"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/services/entity"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/util/msaddr"
)

var log = util.NewLoggerWrap("blue-green")

type VersionsRegistry[R dto.ServicesVersionPayload] interface {
	GetMicroservicesByVersionAsMap(ctx context.Context, repo dao.Repository, version *domain.DeploymentVersion) (map[domain.MicroserviceKey]*domain.MicroserviceVersion, error)

	DeleteVersions(ctx context.Context, repo dao.Repository, versions ...*domain.DeploymentVersion) error
	GetMicroserviceCurrentVersion(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace, initialVersion string) ([]dto.VersionInRegistry, error)
	IsMicroservicePresentInVersion(ctx context.Context, repo dao.Repository, microserviceKey domain.MicroserviceKey, version *domain.DeploymentVersion) (bool, error)
}

type versionsRegistry[R dto.ServicesVersionPayload] struct {
	dao           dao.Dao
	entityService *entity.Service
	bus           bus.BusPublisher
}

func NewVersionsRegistry(dao dao.Dao, entityService *entity.Service, busPublisher bus.BusPublisher) *versionsRegistry[dto.ServicesVersionPayload] {
	return &versionsRegistry[dto.ServicesVersionPayload]{dao: dao, entityService: entityService, bus: busPublisher}
}

func (r *versionsRegistry[R]) GetConfigRes() cfgres.ConfigRes[dto.ServicesVersionPayload] {
	return cfgres.ConfigRes[dto.ServicesVersionPayload]{
		Key: cfgres.ResourceKey{
			APIVersion: "nc.core.mesh/v3",
			Kind:       "ServicesVersion",
		},
		Applier: r,
	}
}

func (r *versionsRegistry[R]) Validate(_ context.Context, res dto.ServicesVersionPayload) (bool, string) {
	return res.Validate()
}

func (r *versionsRegistry[R]) IsOverriddenByCR(_ context.Context, res dto.ServicesVersionPayload) bool {
	return res.Overridden
}

func (r *versionsRegistry[R]) Apply(ctx context.Context, res dto.ServicesVersionPayload) (any, error) {
	isCreated := false
	isDeleted := false
	changes, err := r.dao.WithWTx(func(repo dao.Repository) error {
		// check if we will need to notify blue-green versions watchers
		var err error
		isCreated, err = r.checkIfVersionIsBeingCreated(ctx, repo, res)
		if err != nil {
			return err
		}

		isDeleted, err = r.processMicroserviceVersionsPayload(ctx, repo, res)
		if err != nil {
			return log.ErrorC(ctx, err, "Failed to register microservice versions configuration due to error")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if isCreated || isDeleted {
		if err := r.notifyBgVersionsWatchers(ctx, isCreated, changes); err != nil {
			return nil, err
		}
	}
	logger.InfoC(ctx, "Sending event to update nodes")
	err = r.bus.Publish(bus.TopicChanges, events.NewChangeEvent(changes))
	if err != nil {
		logger.ErrorC(ctx, "Can't publish reload event, cause: %v", err)
		return nil, err
	}
	return map[string]string{"message": "microservices version applied successfully"}, nil
}

func (r *versionsRegistry[R]) GetMicroservicesByVersionAsMap(ctx context.Context, repo dao.Repository, version *domain.DeploymentVersion) (map[domain.MicroserviceKey]*domain.MicroserviceVersion, error) {
	microservices, err := repo.FindMicroserviceVersionsByVersion(version)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "GetMicroservicesByVersionAsMap failed to dao.FindMicroserviceVersionsByVersion %v", *version)
	}
	// collect to map, so later we can check existence and remove elements easily
	microservicesAsMap := util.SliceToMap[*domain.MicroserviceVersion, domain.MicroserviceKey, *domain.MicroserviceVersion](
		microservices,
		func(element *domain.MicroserviceVersion) domain.MicroserviceKey {
			return element.GetMicroserviceKey()
		},
		func(element *domain.MicroserviceVersion) *domain.MicroserviceVersion {
			return element
		},
	)
	return microservicesAsMap, nil
}

func (r *versionsRegistry[R]) DeleteVersions(ctx context.Context, repo dao.Repository, versions ...*domain.DeploymentVersion) error {
	for _, version := range versions {
		msVersions, err := repo.FindMicroserviceVersionsByVersion(version)
		if err != nil {
			return log.ErrorC(ctx, err, "Could not find ms versions for %+v during candidates deletion", version)
		}
		for _, msVersion := range msVersions {
			err = repo.DeleteMicroserviceVersion(msVersion.Name, msaddr.Namespace{Namespace: msVersion.Namespace}, msVersion.InitialDeploymentVersion)
			if err != nil {
				return log.ErrorC(ctx, err, "Could not delete microservice candidate version %+v using DAO", *msVersion)
			}
		}
	}
	return nil
}

func (r *versionsRegistry[R]) GetMicroserviceCurrentVersion(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace, initialVersion string) ([]dto.VersionInRegistry, error) {
	msVersion, err := repo.FindMicroserviceVersionByNameAndInitialVersion(serviceName, namespace, initialVersion)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not find current microservice %s (%+v) version using DAO", serviceName, namespace)
	}
	if msVersion == nil {
		return nil, nil
	}
	dVersion, err := repo.FindDeploymentVersion(msVersion.DeploymentVersion)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not get deployment version %s using DAO", msVersion.DeploymentVersion)
	}
	return r.convertToDto(ctx, repo, []*domain.MicroserviceVersion{msVersion}, dVersion)
}

func (r *versionsRegistry[R]) GetVersionsForMicroservice(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace) ([]dto.VersionInRegistry, error) {
	msVersions, err := repo.FindMicroserviceVersionsByNameAndNamespace(serviceName, namespace)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not get microservice versions service name %s and namespace %v using DAO", serviceName, namespace)
	}
	return r.convertToDto(ctx, repo, msVersions)
}

func (r *versionsRegistry[R]) GetMicroservicesForVersion(ctx context.Context, repo dao.Repository, version *domain.DeploymentVersion) ([]dto.VersionInRegistry, error) {
	msVersions, err := repo.FindMicroserviceVersionsByVersion(version)
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not get microservices for version %s", version.Version)
	}
	if version.Stage == "" {
		version, err = repo.FindDeploymentVersion(version.Version)
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Could not get deployment version %s using DAO", version.Version)
		}
	}
	return r.convertToDto(ctx, repo, msVersions, version)
}

func (r *versionsRegistry[R]) IsMicroservicePresentInVersion(ctx context.Context, repo dao.Repository, microserviceKey domain.MicroserviceKey, version *domain.DeploymentVersion) (bool, error) {
	microserviceVersions, err := repo.FindMicroserviceVersionsByNameAndNamespace(microserviceKey.Name, microserviceKey.GetNamespace())
	if err != nil {
		return false, log.ErrorC(ctx, err, "Could not find ms versions by name and namespace using DAO")
	}
	for _, msVersion := range microserviceVersions {
		if msVersion.DeploymentVersion == version.Version {
			return true, nil
		}
	}
	return false, nil
}

func (r *versionsRegistry[R]) GetAll(ctx context.Context, repo dao.Repository) ([]dto.VersionInRegistry, error) {
	msVersions, err := repo.FindAllMicroserviceVersions()
	if err != nil {
		return nil, log.ErrorC(ctx, err, "Could not load all MicroserviceVersions using DAO")
	}
	return r.convertToDto(ctx, repo, msVersions)
}

func (r *versionsRegistry[R]) checkIfVersionIsBeingCreated(ctx context.Context, repo dao.Repository, payload dto.ServicesVersionPayload) (bool, error) {
	// we will need to notify blue-green versions watchers via websocket in case deployment version does not exist yet, so check:
	// 1) blue-green version explicitly specified in request (means it is not rolling update)
	if payload.Version == "" {
		return false, nil
	}
	// 2) this is CREATE request & this blue-green version is being registered for the first time
	if payload.Exists == nil || *payload.Exists {
		if version, err := repo.FindDeploymentVersion(payload.Version); err != nil {
			return false, log.ErrorC(ctx, err, "Error while checking deployment version existence via DAO\n: %v")
		} else if version != nil {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func (r *versionsRegistry[R]) notifyBgVersionsWatchers(ctx context.Context, isCreated bool, changes []memdb.Change) error {
	var versionChange memdb.Change
	for _, change := range changes {
		if change.Table != domain.DeploymentVersionTable {
			continue
		}
		if isCreated {
			if change.Created() {
				if _, ok := change.After.(*domain.DeploymentVersion); ok {
					versionChange = change
					break
				}
			}
		} else {
			if change.Deleted() {
				if _, ok := change.Before.(*domain.DeploymentVersion); ok {
					versionChange = change
					break
				}
			}
		}
	}
	if versionChange.Before == nil && versionChange.After == nil {
		log.WarnC(ctx, "BG registry did not find expected deployment version change for some reason, so notification will not be sent")
		return nil
	}
	changeEvent := &events.ChangeEvent{
		Changes: map[string][]memdb.Change{
			domain.DeploymentVersionTable: {versionChange},
		},
	}

	if err := r.bus.Publish(bus.TopicBgRegistry, changeEvent); err != nil {
		_ = log.ErrorC(ctx, err, "Could not notify %s about new bg version:\n %v", bus.TopicBgRegistry, err)
		return errors.New("BG registry have been updated, but watchers notification failed with error: " + err.Error())
	}
	return nil
}

func (r *versionsRegistry[R]) processMicroserviceVersionsPayload(ctx context.Context, repo dao.Repository, payload dto.ServicesVersionPayload) (bool, error) {
	version, err := r.entityService.GetOrCreateDeploymentVersion(repo, payload.Version)
	if err != nil {
		return false, log.ErrorC(ctx, err, "Could not get or create deployment version %s", payload.Version)
	}

	namespace := msaddr.Namespace{Namespace: payload.Namespace}

	isCreateRequest := true
	if payload.Exists != nil && !*payload.Exists {
		isCreateRequest = false
	}

	for _, serviceName := range payload.Services {
		if err := r.applyServiceFromPayload(ctx, repo, isCreateRequest, serviceName, namespace, version); err != nil {
			return false, err
		}
	}

	if !isCreateRequest && version.Stage == domain.CandidateStage {
		isDeleted, err := r.deleteDeploymentVersionIfNecessary(ctx, repo, version)
		if err != nil {
			return false, log.ErrorC(ctx, err, "Error in CANDIDATE deployment version cleanup after microservices version deletion")
		}
		return isDeleted, nil
	}
	log.InfoC(ctx, "MicroserviceVersions applied successfully")
	return false, nil
}

func (r *versionsRegistry[R]) applyServiceFromPayload(ctx context.Context, repo dao.Repository, isCreate bool,
	serviceName string, namespace msaddr.Namespace, version *domain.DeploymentVersion) error {
	if isCreate {
		if err := repo.SaveMicroserviceVersion(&domain.MicroserviceVersion{
			Name:                     serviceName,
			Namespace:                namespace.GetNamespace(),
			DeploymentVersion:        version.Version,
			InitialDeploymentVersion: version.Version,
		}); err != nil {
			return log.ErrorC(ctx, err, "Could not save microservice %s (namespace=%s) version %s using DAO", serviceName, namespace.GetNamespace(), version.Version)
		}
	} else {
		if err := repo.DeleteMicroserviceVersion(serviceName, namespace, version.Version); err != nil {
			return log.ErrorC(ctx, err, "Could not delete microservice %s (namespace=%+v) version %s using DAO", serviceName, namespace, version.Version)
		}
	}
	return nil
}

func (r *versionsRegistry[R]) deleteDeploymentVersionIfNecessary(ctx context.Context, repo dao.Repository, version *domain.DeploymentVersion) (bool, error) {
	existingMicroservices, err := repo.FindMicroserviceVersionsByVersion(version)
	if err != nil {
		return false, log.ErrorC(ctx, err, "Failed to check other microservices existence for the version %v", *version)
	}
	if len(existingMicroservices) > 0 {
		return false, nil
	}

	existingEndpoints, err := repo.FindEndpointsByDeploymentVersion(version.Version)
	if err != nil {
		return false, log.ErrorC(ctx, err, "Failed to check endpoints existence for the version %v", *version)
	}
	if len(existingEndpoints) > 0 {
		return false, nil
	}

	return true, repo.DeleteDeploymentVersion(version)
}

func (r *versionsRegistry[R]) getEndpointsForMicroserviceVersion(ctx context.Context, repo dao.Repository, msVersion *domain.MicroserviceVersion) ([]string, error) {
	clusters, err := repo.FindClustersByFamilyNameAndNamespace(msVersion.Name, msaddr.Namespace{Namespace: msVersion.Namespace})
	if err != nil {
		return nil, log.ErrorC(ctx, err, "BG registry could not load clusters of %s in namespace %s using DAO", msVersion.Name, msVersion.Namespace)
	}
	result := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		endpoints, err := repo.FindEndpointsByClusterIdAndDeploymentVersion(cluster.Id, &domain.DeploymentVersion{Version: msVersion.DeploymentVersion})
		if err != nil {
			return nil, log.ErrorC(ctx, err, "BG registry could not load endpoints of cluster %s using DAO", cluster.Name)
		}
		for _, endpoint := range endpoints {
			result = append(result, r.buildEndpointAddr(endpoint))
		}
	}
	return result, nil
}

func (r *versionsRegistry[R]) hasEndpointsOfVersion(ctx context.Context, repo dao.Repository, serviceName string, namespace msaddr.Namespace, deploymentVersion *domain.DeploymentVersion) (bool, error) {
	clusters, err := repo.FindClustersByFamilyNameAndNamespace(serviceName, namespace)
	if err != nil {
		return false, log.ErrorC(ctx, err, "BG registry could not load clusters of %s in namespace %s using DAO", serviceName, namespace.Namespace)
	}
	for _, cluster := range clusters {
		endpoints, err := repo.FindEndpointsByClusterIdAndDeploymentVersion(cluster.Id, deploymentVersion)
		if err != nil {
			return false, log.ErrorC(ctx, err, "BG registry could not load endpoints of cluster %s using DAO", cluster.Name)
		}
		for _, endpoint := range endpoints {
			if endpoint.DeploymentVersion == deploymentVersion.Version {
				return true, nil
			}
		}
	}
	return false, nil
}

func (r *versionsRegistry[R]) convertToDto(ctx context.Context, repo dao.Repository, microservices []*domain.MicroserviceVersion, versions ...*domain.DeploymentVersion) ([]dto.VersionInRegistry, error) {
	if len(versions) == 0 { // if no versions provided, we need to load all of them and pick the ones with microservices
		var err error
		versions, err = repo.FindAllDeploymentVersions()
		if err != nil {
			return nil, log.ErrorC(ctx, err, "Could not load all deployment versions using DAO")
		}
	}
	result := make([]dto.VersionInRegistry, 0, len(versions))
	for _, version := range versions {
		versionDto := dto.VersionInRegistry{
			Version:  version.Version,
			Stage:    version.Stage,
			Clusters: make([]dto.Microservice, 0, len(microservices)),
		}
		for _, ms := range microservices {
			if ms.DeploymentVersion == version.Version {
				endpoints, err := r.getEndpointsForMicroserviceVersion(ctx, repo, ms)
				if err != nil {
					return nil, err
				}
				msDto := dto.Microservice{
					Cluster:   ms.Name,
					Namespace: ms.Namespace,
					Endpoints: endpoints,
				}
				versionDto.Clusters = append(versionDto.Clusters, msDto)
			}
		}
		result = append(result, versionDto)
	}
	return result, nil
}

func (r *versionsRegistry[R]) buildEndpointAddr(endpoint *domain.Endpoint) string {
	if endpoint.Protocol == "" {
		return fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port)
	}
	return fmt.Sprintf("%s://%s:%d", endpoint.Protocol, endpoint.Address, endpoint.Port)
}
