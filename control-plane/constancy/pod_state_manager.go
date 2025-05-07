package constancy

import (
	"errors"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
)

type PodStateManager interface {
	IsCurrentPodDefinedAsMaster() (bool, error)
}

type PodStateManagerImpl struct {
	Storage Storage
}

func (podStateManager *PodStateManagerImpl) IsCurrentPodDefinedAsMaster() (bool, error) {
	actualPodIp := configloader.GetKoanf().String("pod.ip")
	actualPodName := configloader.GetKoanf().String("pod.name")
	electionRecords, err := podStateManager.Storage.FindAllElectionRecords()
	if err != nil {
		return false, err
	}

	if len(electionRecords) != 1 {
		return false, errors.New("count of records in Election table is not equal to 1")
	}

	masterRecord := electionRecords[0]
	return masterRecord.Name == actualPodName && masterRecord.NodeInfo.IP == actualPodIp, nil
}
