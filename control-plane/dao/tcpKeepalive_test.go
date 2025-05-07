package dao

import (
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/netcracker/qubership-core-control-plane/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_TcpKeepalive(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}

	keepAlives, err := testable.FindAllTcpKeepalives()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(keepAlives))

	_, err = testable.WithWTx(func(repo Repository) error {
		err := repo.SaveTcpKeepalive(&domain.TcpKeepalive{
			Probes:   1,
			Time:     1,
			Interval: 1,
		})
		assert.Nil(t, err)
		err = repo.SaveTcpKeepalive(&domain.TcpKeepalive{
			Probes:   2,
			Time:     2,
			Interval: 2,
		})
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)

	tcpKeepalive, err := testable.FindTcpKeepaliveById(3)
	assert.Nil(t, err)
	assert.Nil(t, tcpKeepalive)

	tcpKeepalive, err = testable.FindTcpKeepaliveById(1)
	assert.Nil(t, err)
	assert.NotNil(t, tcpKeepalive)
	assert.Equal(t, int32(1), tcpKeepalive.Probes)
	assert.Equal(t, int32(1), tcpKeepalive.Time)
	assert.Equal(t, int32(1), tcpKeepalive.Interval)

	tcpKeepalive, err = testable.FindTcpKeepaliveById(2)
	assert.Nil(t, err)
	assert.NotNil(t, tcpKeepalive)
	assert.Equal(t, int32(2), tcpKeepalive.Probes)
	assert.Equal(t, int32(2), tcpKeepalive.Time)
	assert.Equal(t, int32(2), tcpKeepalive.Interval)

	keepAlives, err = testable.FindAllTcpKeepalives()
	assert.Nil(t, err)
	assert.Equal(t, 2, len(keepAlives))

	_, err = testable.WithWTx(func(repo Repository) error {
		return repo.DeleteTcpKeepaliveById(1)
	})
	assert.Nil(t, err)

	keepAlives, err = testable.FindAllTcpKeepalives()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(keepAlives))

	tcpKeepalive, err = testable.FindTcpKeepaliveById(1)
	assert.Nil(t, err)
	assert.Nil(t, tcpKeepalive)
}
