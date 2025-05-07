package constancy

import (
	"errors"
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/dr"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/errorcodes"
	asrt "github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type BatchStorageMock struct{}

type BatchTransactionManagerMock struct {
	withTxBatchFunctionInvoked bool
}

type PodStateManagerMock struct {
	podStateDefinedAsMaster bool
	podStateDefinitionError error
}

func (s *BatchStorageMock) Insert(entities []interface{}) error {
	return nil
}

func (s *BatchStorageMock) Update(entities []interface{}) error {
	return nil
}

func (s *BatchStorageMock) Delete(entities []interface{}) error {
	return nil
}

func (b *BatchTransactionManagerMock) WithTxBatch(f func(tx BatchStorage) error) error {
	b.withTxBatchFunctionInvoked = true
	return f(&BatchStorageMock{})
}

func (podStateManager *PodStateManagerMock) IsCurrentPodDefinedAsMaster() (bool, error) {
	return podStateManager.podStateDefinedAsMaster, podStateManager.podStateDefinitionError
}

/* impact testing of DR mode */

func TestFlush_CurrentPodDefinedAsMasterAndDrActiveMode_TransactionWasStarted(t *testing.T) {
	assert := asrt.New(t)

	os.Setenv("EXECUTION_MODE", "active")
	dr.ReloadMode()
	defer dr.ReloadMode()
	defer os.Unsetenv("EXECUTION_MODE")

	podStateManagerMock := &PodStateManagerMock{podStateDefinedAsMaster: true}
	batchTmMock := &BatchTransactionManagerMock{}
	flusher := &Flusher{BatchTm: batchTmMock, PodStateManager: podStateManagerMock}
	var changes []memdb.Change

	err := flusher.Flush(changes)

	assert.Nil(err)
	assert.True(batchTmMock.withTxBatchFunctionInvoked)
}

func TestFlush_CurrentPodDefinedAsMasterAndDrStandbyMode_TransactionWasNotStartedWithoutErrors(t *testing.T) {
	assert := asrt.New(t)

	os.Setenv("EXECUTION_MODE", "standby")
	dr.ReloadMode()
	defer dr.ReloadMode()
	defer os.Unsetenv("EXECUTION_MODE")

	podStateManagerMock := &PodStateManagerMock{podStateDefinedAsMaster: true}
	batchTmMock := &BatchTransactionManagerMock{}
	flusher := &Flusher{BatchTm: batchTmMock, PodStateManager: podStateManagerMock}
	var changes []memdb.Change

	err := flusher.Flush(changes)

	assert.Nil(err)
	assert.False(batchTmMock.withTxBatchFunctionInvoked)
}

/* impact testing of current pod state */

func TestFlush_CurrentPodDefinedAsMaster_TransactionWasStarted(t *testing.T) {
	assert := asrt.New(t)

	podStateManagerMock := &PodStateManagerMock{podStateDefinedAsMaster: true}
	batchTmMock := &BatchTransactionManagerMock{}
	flusher := &Flusher{BatchTm: batchTmMock, PodStateManager: podStateManagerMock}
	var changes []memdb.Change

	err := flusher.Flush(changes)

	assert.True(batchTmMock.withTxBatchFunctionInvoked)
	assert.Nil(err)
}

func TestFlush_CurrentPodDefinedAsSlave_TransactionRolledBackWithError(t *testing.T) {
	assert := asrt.New(t)

	podStateManagerMock := &PodStateManagerMock{podStateDefinedAsMaster: false}
	batchTmMock := &BatchTransactionManagerMock{}
	flusher := &Flusher{BatchTm: batchTmMock, PodStateManager: podStateManagerMock}
	var changes []memdb.Change

	err := flusher.Flush(changes)

	assert.True(batchTmMock.withTxBatchFunctionInvoked)
	assert.Error(err)
}

func TestFlush_ErrorDuringDefinitionCurrentPodState_TransactionRolledBackWithError(t *testing.T) {
	assert := asrt.New(t)

	expectedErrorMsg := "error during definition current pod state"

	podStateManagerMock := &PodStateManagerMock{podStateDefinitionError: errors.New(expectedErrorMsg)}
	batchTmMock := &BatchTransactionManagerMock{}
	flusher := &Flusher{BatchTm: batchTmMock, PodStateManager: podStateManagerMock}
	var changes []memdb.Change

	err := flusher.Flush(changes)

	assert.True(batchTmMock.withTxBatchFunctionInvoked)
	assert.Equal(errorcodes.DbOperationError.ErrorCode, err.(*errorcodes.CpErrCodeError).ErrorCode)
	assert.Equal("Error during flush transaction", err.(*errorcodes.CpErrCodeError).Detail)
	assert.Equal(expectedErrorMsg, err.(*errorcodes.CpErrCodeError).Cause.Error())
}
