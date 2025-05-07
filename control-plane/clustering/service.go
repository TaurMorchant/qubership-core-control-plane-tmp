package clustering

import (
	"context"
	"fmt"
	"github.com/uptrace/bun"
	"time"
)

type NodeInfoServicePG struct {
	sqlService SqlService
}

func (c *NodeInfoServicePG) tryAddInternalRecord() error {
	if err := c.sqlService.WithTx(func(cnn *bun.Conn) error {
		count, _ := c.sqlService.Count(cnn)
		if count == 1 {
			return nil
		}

		if count > 1 {
			_, nerror := c.sqlService.DeleteAllRecords(cnn)
			if nerror != nil {
				return nerror
			}
		}

		electionRecord := MasterMetadata{
			Name: "internal",
			NodeInfo: NodeInfo{
				IP:       "0.0.0.0",
				SWIMPort: 0,
				BusPort:  0,
			},
		}
		_, nerror := c.sqlService.InsertRecordOutdated(cnn, &electionRecord)
		return nerror
	}); err != nil {
		log.Debug("Adding internal record to db caused error: %v", err.Error())
		return err
	}
	return nil
}

func CreateWithExistDb(sqlService SqlService) (ElectionService, error) {
	instance := &NodeInfoServicePG{sqlService: sqlService}

	conn, err := sqlService.Conn()
	if err != nil {
		return nil, err
	}

	defer func(conn *bun.Conn) {
		// Required for tests
		if conn != nil {
			conn.Close()
		}
	}(conn)

	_, err = sqlService.CreateElectionTable(conn)
	if err != nil {
		return nil, err
	}

	err = instance.tryAddInternalRecord()
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func (c *NodeInfoServicePG) DeleteSeveralRecordsFromDb() error {
	if err := c.sqlService.WithTx(func(cnn *bun.Conn) error {
		// TODO: review this step. Why do we just delete without examination of a content in the table?
		// Delete all not 'internal' records
		res, err := c.sqlService.DeleteNotInternalRecords(cnn)
		log.Info("Several records are found in election table. Delete result: %v", res)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Debug("Deleting several records from db caused error: %v", err.Error())
		return err
	}

	// Recreate 'internal' record
	err := c.tryAddInternalRecord()
	if err != nil {
		return err
	}

	return nil
}

func (c *NodeInfoServicePG) updateAsMaster(cnn *bun.Conn, electionRecord *MasterMetadata) (bool, error) {
	// Update as master
	rowsAffected, err := c.sqlService.UpdateMasterRecord(cnn, electionRecord)
	if err != nil {
		return false, err
	}

	// No suitable record
	if rowsAffected == 0 {
		return false, nil
	}

	// At least one row affected
	if rowsAffected == 1 {
		return true, nil
	}

	// Several rows affected
	return true, fmt.Errorf("affected several rows: %d", rowsAffected)
}

func (c *NodeInfoServicePG) restoreElectionTable(cnn *bun.Conn, electionRecord *MasterMetadata, erase bool) (bool, error) {
	log.Warn("Several records or no records in election table. Restoring…")

	// Delete all records as the election is broken
	if erase {
		res, err := c.sqlService.DeleteAllRecords(cnn)
		log.Debug("Several records are found in election table. Delete result: %v", res)
		if err != nil {
			return false, err
		}
	}

	// Insert as master
	log.Debug("Recreating current master record")
	_, err := c.sqlService.InsertRecord(cnn, electionRecord)
	if err != nil {
		return false, err
	}

	log.Info("Election table successfully restored")
	return true, nil
}

func (c *NodeInfoServicePG) TryWriteAsMaster(electionRecord *MasterMetadata) bool {
	writtenAsMaster := false
	if err := c.sqlService.WithTx(func(cnn *bun.Conn) error {
		// Total number of rows needs to be obtained as there can be a case of several rows with only one affected
		totalRowsCount, nerror := c.sqlService.Count(cnn)
		if nerror != nil {
			return nerror
		}

		if totalRowsCount == 1 {
			writtenAsMaster, nerror = c.updateAsMaster(cnn, electionRecord)
		} else {
			writtenAsMaster, nerror = c.restoreElectionTable(cnn, electionRecord, totalRowsCount != 0)
		}
		return nerror
	}); err != nil {
		log.Warn("Writing election info caused error: %v", err.Error())
	}
	if writtenAsMaster {
		writtenMaster, _ := c.GetMaster()
		electionRecord.Id = writtenMaster.Id
	}
	return writtenAsMaster
}

func (c *NodeInfoServicePG) GetMaster() (*MasterMetadata, error) {
	var result MasterMetadata
	err := c.sqlService.WithTx(func(conn *bun.Conn) error {
		record, nerror := c.sqlService.GetMaster(conn)
		if nerror != nil {
			return nerror
		}
		result = *record
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *NodeInfoServicePG) ShiftSyncClock(d time.Duration) error {
	return c.sqlService.WithTx(func(conn *bun.Conn) error {
		_, err := c.sqlService.ShiftSyncClock(conn, d)
		return err
	})
}

func (c *NodeInfoServicePG) ResetSyncClock(master string) error {
	return c.sqlService.WithTx(func(conn *bun.Conn) error {
		_, err := c.sqlService.ResetSyncClock(conn, master)
		return err
	})
}

func (c *NodeInfoServicePG) BeforeQuery(ctx context.Context, queryEvent *bun.QueryEvent) context.Context {
	return ctx
}

func (c *NodeInfoServicePG) AfterQuery(ctx context.Context, queryEvent *bun.QueryEvent) {
	sql := queryEvent.Query
	log.Debug("PG Query: %v\n", sql)
}
