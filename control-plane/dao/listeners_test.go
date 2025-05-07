package dao

import (
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/ram"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemDao_FindAllListeners(t *testing.T) {
	testable := &InMemRepo{
		storage:     ram.NewStorage(),
		idGenerator: &GeneratorMock{},
	}
	listeners := []*domain.Listener{
		{
			Name:        "Listener1",
			NodeGroupId: "nodeGroup1",
		},
		{
			Name:        "Listener2",
			NodeGroupId: "nodeGroup1",
		},
		{
			Name:        "Listener3",
			NodeGroupId: "nodeGroup2",
		},
	}
	_, err := testable.WithWTx(func(dao Repository) error {
		for _, listener := range listeners {
			assert.Nil(t, dao.SaveListener(listener))
		}
		return nil
	})
	assert.Nil(t, err)

	foundListeners, err := testable.FindAllListeners()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(foundListeners))

	foundListener, err := testable.FindListenerByNodeGroupIdAndName("nodeGroup1", "Listener2")
	assert.Nil(t, err)
	assert.Equal(t, listeners[1], foundListener)
}
