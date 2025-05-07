package constancy

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/netcracker/qubership-core-control-plane/clustering"
	mock_constancy "github.com/netcracker/qubership-core-control-plane/test/mock/constancy"
	"github.com/netcracker/qubership-core-lib-go/v3/configloader"
	asrt "github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestIsCurrentPodDefinedAsMaster_ElectionRecordMatchesPodInfo_ReturnsTrueWithoutError(t *testing.T) {
	assert := asrt.New(t)

	podName := "control-plane-5b7f6c77d6-76fr7"
	podIp := "172.31.26.7"

	os.Setenv("POD_IP", podIp)
	os.Setenv("POD_NAME", podName)
	defer os.Unsetenv("POD_IP")
	defer os.Unsetenv("POD_NAME")
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mock_constancy.NewMockStorage(ctrl)
	electionRecords := getElectionRecordsMock(podName, podIp, 1)
	storage.EXPECT().FindAllElectionRecords().Return(electionRecords, nil)

	podStateManagerImpl := &PodStateManagerImpl{Storage: storage}

	isCurrentPodDefinedAsMaster, err := podStateManagerImpl.IsCurrentPodDefinedAsMaster()

	assert.Nil(err)
	assert.True(isCurrentPodDefinedAsMaster)
}

func TestIsCurrentPodDefinedAsMaster_ElectionRecordDoesNotMatchesPodIp_ReturnsFalseWithoutError(t *testing.T) {
	assert := asrt.New(t)

	podName := "control-plane-5b7f6c77d6-76fr7"
	actualPodIp := "172.31.26.7"
	electionRecordPodIp := "172.11.11.11"

	os.Setenv("POD_IP", actualPodIp)
	os.Setenv("POD_NAME", podName)
	defer os.Unsetenv("POD_IP")
	defer os.Unsetenv("POD_NAME")
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mock_constancy.NewMockStorage(ctrl)
	electionRecords := getElectionRecordsMock(podName, electionRecordPodIp, 1)
	storage.EXPECT().FindAllElectionRecords().Return(electionRecords, nil)

	podStateManagerImpl := &PodStateManagerImpl{Storage: storage}

	isCurrentPodDefinedAsMaster, err := podStateManagerImpl.IsCurrentPodDefinedAsMaster()

	assert.Nil(err)
	assert.False(isCurrentPodDefinedAsMaster)
}

func TestIsCurrentPodDefinedAsMaster_ElectionRecordDoesNotMatchesPodName_ReturnsFalseWithoutError(t *testing.T) {
	assert := asrt.New(t)

	actualPodName := "control-plane-5b7f6c77d6-76fr7"
	electionRecordPodName := "control-plane-1111111-111111"
	podIp := "172.31.26.7"

	os.Setenv("POD_IP", podIp)
	os.Setenv("POD_NAME", actualPodName)
	defer os.Unsetenv("POD_IP")
	defer os.Unsetenv("POD_NAME")
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mock_constancy.NewMockStorage(ctrl)
	electionRecords := getElectionRecordsMock(electionRecordPodName, podIp, 1)
	storage.EXPECT().FindAllElectionRecords().Return(electionRecords, nil)

	podStateManagerImpl := &PodStateManagerImpl{Storage: storage}

	isCurrentPodDefinedAsMaster, err := podStateManagerImpl.IsCurrentPodDefinedAsMaster()

	assert.Nil(err)
	assert.False(isCurrentPodDefinedAsMaster)
}

func TestIsCurrentPodDefinedAsMaster_ErrorDuringFindElectionRecords_ReturnsFalseWithError(t *testing.T) {
	assert := asrt.New(t)

	podName := "control-plane-5b7f6c77d6-76fr7"
	podIp := "172.31.26.7"

	os.Setenv("POD_IP", podIp)
	os.Setenv("POD_NAME", podName)
	defer os.Unsetenv("POD_IP")
	defer os.Unsetenv("POD_NAME")
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mock_constancy.NewMockStorage(ctrl)
	errorMsg := "error during find election records"
	storage.EXPECT().FindAllElectionRecords().Return(nil, errors.New(errorMsg))

	podStateManagerImpl := &PodStateManagerImpl{Storage: storage}

	isCurrentPodDefinedAsMaster, err := podStateManagerImpl.IsCurrentPodDefinedAsMaster()

	assert.EqualError(err, errorMsg)
	assert.False(isCurrentPodDefinedAsMaster)
}

func TestIsCurrentPodDefinedAsMaster_ElectionRecordsCountMoreThenOne_ReturnsFalseWithError(t *testing.T) {
	assert := asrt.New(t)

	podName := "control-plane-5b7f6c77d6-76fr7"
	podIp := "172.31.26.7"

	os.Setenv("POD_IP", podIp)
	os.Setenv("POD_NAME", podName)
	defer os.Unsetenv("POD_IP")
	defer os.Unsetenv("POD_NAME")
	configloader.Init(configloader.EnvPropertySource())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mock_constancy.NewMockStorage(ctrl)
	electionRecords := getElectionRecordsMock(podName, podIp, 2)
	storage.EXPECT().FindAllElectionRecords().Return(electionRecords, nil)

	podStateManagerImpl := &PodStateManagerImpl{Storage: storage}

	isCurrentPodDefinedAsMaster, err := podStateManagerImpl.IsCurrentPodDefinedAsMaster()

	assert.Error(err)
	assert.False(isCurrentPodDefinedAsMaster)
}

func getElectionRecordsMock(podName string, podIp string, count int) []*clustering.MasterMetadata {
	electionRecords := make([]*clustering.MasterMetadata, 0)
	for i := 0; i < count; i++ {
		electionRecord := &clustering.MasterMetadata{
			Id:        int64(i + 1),
			Name:      podName,
			NodeInfo:  clustering.NodeInfo{IP: podIp, SWIMPort: 1234, BusPort: 5431, HttpPort: 8080},
			SyncClock: time.Now(),
		}
		electionRecords = append(electionRecords, electionRecord)
	}
	return electionRecords
}
