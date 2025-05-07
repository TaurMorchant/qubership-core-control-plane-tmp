package events

import (
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChangeEventMarshalPrepare(t *testing.T) {
	pr := &mockPreparer{}
	changes := map[string][]memdb.Change{
		"A": []memdb.Change{{Before: pr}},
	}
	event := &ChangeEvent{Changes: changes}
	assert.Nil(t, event.MarshalPrepare())
	assert.True(t, pr.called)
}

func TestChangeEventToString(t *testing.T) {
	pr := &mockPreparer{}
	changes := map[string][]memdb.Change{
		"A": []memdb.Change{{Before: pr}},
	}
	event := &ChangeEvent{Changes: changes}
	assert.Contains(t, event.ToString(), "{ChangeEvent: nodeGroup='', changes=[\nA: [&{false}, nil, ], \n]}")
}

func TestMultipleChangeEventMarshalPrepare(t *testing.T) {
	pr := &mockPreparer{}
	changes := map[string][]memdb.Change{
		"A": []memdb.Change{{Before: pr}},
	}
	event := &MultipleChangeEvent{Changes: changes}
	assert.Nil(t, event.MarshalPrepare())
	assert.True(t, pr.called)
}

func TestReloadEventMarshalPrepare(t *testing.T) {
	pr := &mockPreparer{}
	changes := map[string][]memdb.Change{
		"A": []memdb.Change{{Before: pr}},
	}
	event := &ReloadEvent{Changes: changes}
	assert.Nil(t, event.MarshalPrepare())
	assert.True(t, pr.called)
}

type mockPreparer struct {
	called bool
}

func (m *mockPreparer) MarshalPrepare() error {
	m.called = true
	return nil
}
