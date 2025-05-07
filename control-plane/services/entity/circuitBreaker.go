package entity

import "github.com/netcracker/qubership-core-control-plane/dao"

func (srv *Service) DeleteCircuitBreakerCascadeById(dao dao.Repository, id int32) error {
	circuitBreaker, err := dao.FindCircuitBreakerById(id)
	if err != nil {
		logger.Errorf("Error while searching for existing CircuitBreaker with id %v: %v", id, err)
		return err
	}
	if circuitBreaker.ThresholdId != 0 {
		err = dao.DeleteThresholdById(circuitBreaker.ThresholdId)
		if err != nil {
			logger.Errorf("Error while deleting threshold with id %v: %v", circuitBreaker.ThresholdId, err)
			return err
		}
	}
	err = dao.DeleteCircuitBreakerById(id)
	if err != nil {
		logger.Errorf("Error while deleting CircuitBreaker with id %v: %v", id, err)
		return err
	}
	return nil
}
