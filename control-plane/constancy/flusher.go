package constancy

import (
	"github.com/hashicorp/go-memdb"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/dr"
	"github.com/netcracker/qubership-core-control-plane/errorcodes"
)

type BatchStorage interface {
	Insert([]interface{}) error
	Update([]interface{}) error
	Delete([]interface{}) error
}

type BatchTransactionManager interface {
	WithTxBatch(func(tx BatchStorage) error) error
}

type Flusher struct {
	BatchTm         BatchTransactionManager
	PodStateManager PodStateManager
}

func (f *Flusher) Flush(changes []memdb.Change) error {
	log.Debug("flashing changes started")
	if dr.GetMode() == dr.Standby {
		return nil
	}
	changesByType := make(map[string][]memdb.Change)
	for _, change := range changes {
		var typedChanges []memdb.Change
		if value, ok := changesByType[change.Table]; ok {
			typedChanges = value
		} else {
			typedChanges = make([]memdb.Change, 0)
		}
		typedChanges = append(typedChanges, change)
		changesByType[change.Table] = typedChanges
	}
	err := f.BatchTm.WithTxBatch(func(storage BatchStorage) error {
		deleteStuck := newStack()
		for _, table := range domain.TableRelationOrder {
			if typedChanges, ok := changesByType[table]; ok {
				insertsPack, updatesPack, deletesPack := make([]interface{}, 0), make([]interface{}, 0), make([]interface{}, 0)
				for _, change := range typedChanges {
					if change.Created() {
						insertsPack = append(insertsPack, change.After)
					} else if change.Updated() {
						updatesPack = append(updatesPack, change.After)
					} else if change.Deleted() {
						deletesPack = append(deletesPack, change.Before)
					}
				}
				log.Debugf("insertsPack size %v", len(insertsPack))
				if err := storage.Insert(insertsPack); err != nil {
					return err
				}

				log.Debugf("updatesPack size %v", len(updatesPack))
				if err := storage.Update(updatesPack); err != nil {
					return err
				}

				log.Debugf("deletesPack size %v", len(deletesPack))
				deleteStuck.Push(deletesPack)
			}
		}
		// Because delete must have reverse order
		for deletesPack := deleteStuck.Pop(); deletesPack != nil; deletesPack = deleteStuck.Pop() {
			if err := storage.Delete(deletesPack); err != nil {
				return err
			}
		}

		return f.CheckCurrentPodState()
	})
	if err != nil {
		if rootCauseErr := errorcodes.GetCpErrCodeErrorOrNil(err); rootCauseErr != nil {
			return rootCauseErr
		}
		log.Errorf("error during flush transaction, cause: %v", err.Error())
		return errorcodes.NewCpError(errorcodes.DbOperationError, "Error during flush transaction", err)
	}
	return nil
}

func (f *Flusher) CheckCurrentPodState() error {
	log.Debugf("checking current pod state")
	// Transaction can't be completed in case of a switch of the master node.
	// This can lead to data inconsistency in the new master node.
	isCurrentPodDefinedAsMaster, err := f.PodStateManager.IsCurrentPodDefinedAsMaster()
	if err != nil {
		return err
	}
	if !isCurrentPodDefinedAsMaster {
		return errorcodes.NewCpError(errorcodes.MasterNodeError, "Master node was switched at the moment. Please, Try again later", nil)
	}
	return nil
}

type stuck struct {
	s [][]interface{}
}

func newStack() *stuck {
	return &stuck{s: make([][]interface{}, 0)}
}

func (s *stuck) Push(elem []interface{}) {
	s.s = append(s.s, elem)
}

func (s *stuck) Pop() []interface{} {
	l := len(s.s)
	if l == 0 {
		return nil
	}
	elem := s.s[l-1]
	s.s = s.s[:l-1]
	return elem
}
