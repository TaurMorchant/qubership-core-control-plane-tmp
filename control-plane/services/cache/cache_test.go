package cache

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	client   = NewCacheClient()
	testData = map[string]string{
		"uid-1": "ns-1",
		"uid-2": "ns-2",
		"uid-3": "ns-3",
		"uid-4": "ns-4",
	}
	expectedWrongKey = "uid-1000"
)

func TestSize(t *testing.T) {
	for key, value := range testData {
		client.Set(key, value)
	}
	selectedItems := client.GetAll()
	assert.Equal(t, len(testData), len(selectedItems))
}

func TestGet(t *testing.T) {
	for key, expectedValue := range testData {
		actualValue, ok := client.Get(key)
		assert.True(t, ok)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestDelete(t *testing.T) {
	expectedKey := "uid-100"
	expectedValue := "ns-100"
	client.Set(expectedKey, expectedValue)
	actualValue, ok := client.Get(expectedKey)
	assert.True(t, ok)
	assert.Equal(t, expectedValue, actualValue)
	client.Delete(expectedWrongKey)
	assert.NotPanics(t, wrongKeyTest)
	client.Delete(expectedKey)
	actualValue, ok = client.Get(expectedKey)
	assert.False(t, ok)
	assert.Empty(t, actualValue)
}

func wrongKeyTest() {
	client.Delete(expectedWrongKey)
}
