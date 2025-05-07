package clustering

import (
	"database/sql"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/db"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"testing"
	"time"
)

var sqlMockError = fmt.Errorf("SqlServiceMock invoked error")

func TestElectorService_ClusteringService_CreateWithExistDb(t *testing.T) {
	electionSrv, err := CreateWithExistDb(NewSqlServiceMock(&DbProviderStub{nil}))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.NotNil(t, electionSrv)
}

func TestElectorService_ClusteringService_CreateWithExistDb_BadDbProvider(t *testing.T) {
	_, err := CreateWithExistDb(
		NewSqlServiceMock(&DbProviderStub{fmt.Errorf("DbProviderStub")}),
	)
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_CreateWithExistDb_ErrorOnCreate(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 1)
	_, err := CreateWithExistDb(sqlServiceMock)
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_CreateWithExistDb_ErrorOnInternal(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 3)
	_, err := CreateWithExistDb(sqlServiceMock)
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_CreateWithExistDb_SeveralRecordsExist(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	_, err := sqlServiceMock.InsertRecord(nil, &MasterMetadata{Name: "test-record-1"})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = sqlServiceMock.InsertRecord(nil, &MasterMetadata{Name: "test-record-2"})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
	_, err = CreateWithExistDb(sqlServiceMock)
	assert.Nil(t, err)
	count, err = sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
}

func TestElectorService_ClusteringService_CreateWithExistDb_SeveralRecordsExist_ErrorOnInternal(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	_, err := sqlServiceMock.InsertRecord(nil, &MasterMetadata{Name: "test-record-1"})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = sqlServiceMock.InsertRecord(nil, &MasterMetadata{Name: "test-record-2"})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 3)
	_, err = CreateWithExistDb(sqlServiceMock)
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_DeleteSeveralRecordsFromDb(t *testing.T) {
	sqlService := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlService)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	err = electionSrv.DeleteSeveralRecordsFromDb()
	assert.Nil(t, err)
}

func TestElectorService_ClusteringService_DeleteSeveralRecordsFromDb_ErrorOnDeleting(t *testing.T) {
	sqlService := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlService)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	// Error on deleting
	sqlService.SetCustomResult(&CustomResult{0, sqlMockError}, 1)
	err = electionSrv.DeleteSeveralRecordsFromDb()
	assert.NotNil(t, err)

}

func TestElectorService_ClusteringService_DeleteSeveralRecordsFromDb_ErrorOnInsert(t *testing.T) {
	sqlService := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlService)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	// Error on inserting 'internal' record
	sqlService.ClearTable()
	sqlService.SetCustomResult(&CustomResult{0, sqlMockError}, 3)
	err = electionSrv.DeleteSeveralRecordsFromDb()
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_DeleteSeveralRecordsFromDb_SeveralRecordsOnInsert(t *testing.T) {
	sqlService := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlService)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	sqlService.ClearTable()
	sqlService.SetCustomResult(&CustomResult{2, nil}, 2)
	err = electionSrv.DeleteSeveralRecordsFromDb()
	assert.Nil(t, err)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_OneRecord(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	// Overwrite 'internal'
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.True(t, written)
	recordBefore, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, electionRecord.Name, recordBefore.Name)

	// Try to overwrite
	written = electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	recordAfter, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, recordBefore, recordAfter)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_OneRecord_CannotOverride(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	// Overwrite 'internal'
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.True(t, written)
	recordBefore, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, electionRecord.Name, recordBefore.Name)

	// Try to overwrite
	written = electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	recordAfter, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, recordBefore, recordAfter)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_NoRecords(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	sqlServiceMock.ClearTable()
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)

	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.True(t, written)
	recordAfter, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, electionRecord.Name, recordAfter.Name)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_SeveralRecords(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	_, err = sqlServiceMock.InsertRecordOutdated(nil, electionRecord)
	assert.Nil(t, err)
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.True(t, written)
	_, err = electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
}

func TestElectorService_ClusteringService_TryWriteAsMaster_CountError(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 1)
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	recordAfter, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.NotEqual(t, electionRecord.Name, recordAfter.Name)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_UpdateAsMaster_UpdateMasterRecord_Error(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 2)
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	recordAfter, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.NotEqual(t, electionRecord.Name, recordAfter.Name)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_UpdateAsMaster_SeveralRecordsAffected(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	sqlServiceMock.SetCustomResult(&CustomResult{2, nil}, 2)
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.True(t, written)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_RestoreElectionTable_DeleteAllRecordsError(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	_, err = sqlServiceMock.InsertRecordOutdated(nil, electionRecord)
	assert.Nil(t, err)
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 2)
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	count, err = sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)
}

func TestElectorService_ClusteringService_TryWriteAsMaster_RestoreElectionTable_InsertRecordError(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	electionRecord := &MasterMetadata{Name: "test-record"}

	_, err = sqlServiceMock.InsertRecordOutdated(nil, electionRecord)
	assert.Nil(t, err)
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, count)

	sqlServiceMock.SetCustomResult(&CustomResult{0, sqlMockError}, 3)
	written := electionSrv.TryWriteAsMaster(electionRecord)
	assert.False(t, written)
	count, err = sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
}

func TestElectorService_ClusteringService_GetMaster(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	masterRecord, err := electionSrv.GetMaster()
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "internal", masterRecord.Name)
}

func TestElectorService_ClusteringService_GetMaster_NoRecords(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	sqlServiceMock.ClearTable()
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	_, err = electionSrv.GetMaster()
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_ShiftSyncClock(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	master, err := sqlServiceMock.GetMaster(nil)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	timeBefore := master.SyncClock

	err = electionSrv.ShiftSyncClock(1 * time.Hour)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	master, err = sqlServiceMock.GetMaster(nil)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	timeAfter := master.SyncClock

	assert.True(t, timeBefore.Before(timeAfter))
}

func TestElectorService_ClusteringService_ShiftSyncClock_NoRecords(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	sqlServiceMock.ClearTable()
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	err = electionSrv.ShiftSyncClock(1 * time.Hour)
	assert.NotNil(t, err)
}

func TestElectorService_ClusteringService_ResetSyncClock(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)

	master, err := sqlServiceMock.GetMaster(nil)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	timeBefore := master.SyncClock

	err = electionSrv.ResetSyncClock("internal")
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	master, err = sqlServiceMock.GetMaster(nil)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}
	timeAfter := master.SyncClock

	assert.True(t, timeBefore.Before(timeAfter))
}

func TestElectorService_ClusteringService_ResetSyncClock_NoRecords(t *testing.T) {
	sqlServiceMock := NewSqlServiceMock(&DbProviderStub{nil})
	electionSrv, err := CreateWithExistDb(sqlServiceMock)
	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	sqlServiceMock.ClearTable()
	count, err := sqlServiceMock.Count(nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	err = electionSrv.ResetSyncClock("internal")
	assert.NotNil(t, err)
}

type CustomResult struct {
	Result int
	Error  error
}

type SqlServiceMock struct {
	masterMetadata    []*MasterMetadata
	customResultDelay uint
	customResult      *CustomResult
	dbProvider        db.DBProvider
}

func NewSqlServiceMock(dbProvider db.DBProvider) *SqlServiceMock {
	return &SqlServiceMock{
		masterMetadata:    make([]*MasterMetadata, 0),
		customResultDelay: 0,
		customResult:      nil,
		dbProvider:        dbProvider,
	}
}

func (m *SqlServiceMock) returnCustomResponseIfCounted() *CustomResult {
	if m.customResultDelay == 0 {
		return nil
	}
	if m.customResultDelay == 1 {
		m.customResultDelay = 0
		return m.customResult
	}
	m.customResultDelay--
	return nil
}

func (m *SqlServiceMock) SetCustomResult(customResult *CustomResult, delay uint) {
	m.customResult = customResult
	m.customResultDelay = delay
}

func (m *SqlServiceMock) CreateElectionTable(cnn *bun.Conn) (sql.Result, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return nil, cRes.Error
	}
	return nil, nil
}

func (m *SqlServiceMock) InsertRecordOutdated(cnn *bun.Conn, record *MasterMetadata) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	record.SyncClock = time.Now().Add(-1 * time.Hour)
	m.masterMetadata = append(m.masterMetadata, record)
	return 1, nil
}

func (m *SqlServiceMock) InsertRecord(cnn *bun.Conn, record *MasterMetadata) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	curTime := time.Now().Add(1 * time.Hour)
	record.SyncClock = curTime
	m.masterMetadata = append(m.masterMetadata, record)
	return 1, nil
}

func (m *SqlServiceMock) DeleteAllRecords(cnn *bun.Conn) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	affectedRowsCount := len(m.masterMetadata)
	m.masterMetadata = make([]*MasterMetadata, 0)
	return int64(affectedRowsCount), nil
}

func (m *SqlServiceMock) DeleteNotInternalRecords(cnn *bun.Conn) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	var newMetadata []*MasterMetadata
	for _, metadata := range m.masterMetadata {
		if metadata.Name == "internal" {
			newMetadata = append(newMetadata, metadata)
		}
	}
	affectedRowsCount := len(m.masterMetadata) - len(newMetadata)
	m.masterMetadata = newMetadata
	return int64(affectedRowsCount), nil
}

func (m *SqlServiceMock) UpdateMasterRecord(cnn *bun.Conn, record *MasterMetadata) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	affectedRowsCount := 0
	curTime := time.Now()
	for _, metadata := range m.masterMetadata {
		if metadata.SyncClock.Before(curTime) {
			*metadata = *record
			metadata.SyncClock = curTime.Add(60 * time.Second)
			affectedRowsCount++
		}
	}
	return int64(affectedRowsCount), nil
}

func (m *SqlServiceMock) ShiftSyncClock(cnn *bun.Conn, d time.Duration) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	if len(m.masterMetadata) == 0 {
		return 0, fmt.Errorf("no rows in db")
	}
	m.masterMetadata[0].SyncClock = time.Now().Add(d)
	return 1, nil
}

func (m *SqlServiceMock) ResetSyncClock(cnn *bun.Conn, recordName string) (int64, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return int64(cRes.Result), cRes.Error
	}
	if len(m.masterMetadata) == 0 {
		return 0, fmt.Errorf("no rows in db")
	}
	affectedRowsCount := 0
	for _, metadata := range m.masterMetadata {
		if metadata.Name == recordName {
			metadata.SyncClock = time.Now()
			affectedRowsCount++
		}
	}
	return int64(affectedRowsCount), nil
}

func (m *SqlServiceMock) GetMaster(cnn *bun.Conn) (*MasterMetadata, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return nil, cRes.Error
	}
	if len(m.masterMetadata) == 0 {
		return nil, fmt.Errorf("no rows in db")
	}
	return m.masterMetadata[0], nil
}

func (m *SqlServiceMock) Count(cnn *bun.Conn) (int, error) {
	if cRes := m.returnCustomResponseIfCounted(); cRes != nil {
		return cRes.Result, cRes.Error
	}
	return len(m.masterMetadata), nil
}

func (m *SqlServiceMock) Conn() (*bun.Conn, error) {
	return m.dbProvider.GetConn(ctx)
}

func (m *SqlServiceMock) WithTx(f func(conn *bun.Conn) error) error {
	return f(nil)
}

func (m *SqlServiceMock) ClearTable() {
	m.masterMetadata = make([]*MasterMetadata, 0)
}
