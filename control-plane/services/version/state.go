package version

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"sort"
)

var logger logging.Logger

func init() {
	logger = logging.GetLogger("version/state")
}

type VersionsState struct {
	versions         []*domain.DeploymentVersion
	versionToPromote *domain.DeploymentVersion
}

func NewVersionState(versions []*domain.DeploymentVersion) *VersionsState {
	sort.Slice(versions, func(i, j int) bool {
		iVersion, err := versions[i].NumericVersion()
		if err != nil {
			logger.Errorf("Cannot determine version for %v, zero version will be used in version sorting", versions[i])
		}
		jVersion, err := versions[j].NumericVersion()
		if err != nil {
			logger.Errorf("Cannot determine version for %v, zero version will be used in version sorting", versions[j])
		}
		return iVersion < jVersion
	})
	version := &VersionsState{versionToPromote: domain.NewDeploymentVersion("", "")}
	version.versions = version.cloneSliceWithEntity(versions)
	return version
}

func (version *VersionsState) cloneSliceWithEntity(sliceToClone []*domain.DeploymentVersion) []*domain.DeploymentVersion {
	versions := make([]*domain.DeploymentVersion, len(sliceToClone))
	for index, versionToClone := range sliceToClone {
		versions[index] = versionToClone.Clone()
	}
	return versions
}

func (version *VersionsState) SetVersionToPromote(versionToPromote *domain.DeploymentVersion) {
	version.versionToPromote = versionToPromote
}

func (version *VersionsState) GetActive() *domain.DeploymentVersion {
	var activeVersion *domain.DeploymentVersion
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.ActiveStage {
			activeVersion = dVersion
			break
		}
	}
	return activeVersion
}

func (version *VersionsState) GetLegacy() *domain.DeploymentVersion {
	var legacyVersion *domain.DeploymentVersion
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.LegacyStage {
			legacyVersion = dVersion
			break
		}
	}
	if legacyVersion == nil {
		return nil
	}
	return legacyVersion
}

func (version *VersionsState) GetCandidates() []*domain.DeploymentVersion {
	var candidateVersions []*domain.DeploymentVersion
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.CandidateStage && version.versionToPromote.Version != dVersion.Version {
			candidateVersions = append(candidateVersions, dVersion)
		}
	}
	return candidateVersions
}

func (version *VersionsState) GetOldestArchivedVersion() *domain.DeploymentVersion {
	var oldestArchivedVersion *domain.DeploymentVersion
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.ArchivedStage {
			oldestArchivedVersion = dVersion
		}
	}
	return oldestArchivedVersion
}

func (version *VersionsState) GetVersions() []*domain.DeploymentVersion {
	return version.versions
}

func (version *VersionsState) GetHistorySize() int {
	historySize := 0
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.ArchivedStage {
			historySize++
		}
	}
	return historySize
}

func (version *VersionsState) GetArchivedVersionsToDelete(numberToDelete int) []*domain.DeploymentVersion {
	var oldestArchivedVersion []*domain.DeploymentVersion
	for _, dVersion := range version.versions {
		if dVersion.Stage == domain.ArchivedStage {
			oldestArchivedVersion = append(oldestArchivedVersion, dVersion)
			numberToDelete--
			if numberToDelete == 0 {
				break
			}
		}
	}
	return oldestArchivedVersion
}
