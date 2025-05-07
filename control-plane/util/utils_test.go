package util

import (
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	_assert "github.com/stretchr/testify/assert"
	"sort"
	"sync"
	"testing"
	"time"
)

var logger = logging.GetLogger("util")

const TestResourceName = "test-resource"
const TestProcessingTime = 500 * time.Millisecond
const ConcurrentRequestsNum = 10000 // high concurrency rate to test that we don't burn out CPU
const NumOfWorkersToWaitFor = 10    // we will wait only for NumOfWorkersToWaitFor workers so test won't take ages to complete

func TestToSlice_shouldPanic_whenSliceNotCorrect(t *testing.T) {
	var slice interface{}
	_assert.Panics(t, func() { ToSlice(slice) }, "The code should panic")
}

func TestToSlice_shouldExpectedSlice_whenSliceCorrect(t *testing.T) {
	initialSlice := make([]interface{}, 5)
	initialSlice[0] = "a"
	resultSlice := ToSlice(initialSlice)
	_assert.Equal(t, initialSlice, resultSlice)
}

func TestArrContainsStr_shouldTrue_whenElemContains(t *testing.T) {
	arr := []string{"1", "2", "3", "4", "test"}
	elem := "test"
	found := SliceContainsElement(arr, elem)
	_assert.True(t, found)
}

func TestArrContainsStr_shouldFalse_whenElemNotContains(t *testing.T) {
	arr := []string{"1", "2", "3", "4", "5"}
	elem := "test"
	found := SliceContainsElement(arr, elem)
	_assert.False(t, found)
}

func TestAtomicCounterValue_shouldReturnCorrectValue(t *testing.T) {
	initialValue := 0
	atomicCounterValue := NewAtomicCounter(initialValue).Value()
	_assert.Equal(t, initialValue, atomicCounterValue)
}

func TestAtomicCounterAddAndReturn_shouldReturnCorrectValue_whenValuePositive(t *testing.T) {
	initialValue := 1
	deltaValue := 4
	atomicCounterValue := NewAtomicCounter(initialValue).AddAndReturn(deltaValue)
	_assert.Equal(t, initialValue+deltaValue, atomicCounterValue)
}

func TestAtomicCounterAddAndReturn_shouldReturnCorrectValue_whenValueNegative(t *testing.T) {
	initialValue := 1
	deltaValue := -1
	atomicCounterValue := NewAtomicCounter(initialValue).AddAndReturn(deltaValue)
	_assert.Equal(t, initialValue+deltaValue, atomicCounterValue)
}

func TestBoolToInt_shouldReturn1_whenValuePositive(t *testing.T) {
	value := true
	expectedValue := 1
	boolValue := BoolToInt(value)
	_assert.Equal(t, expectedValue, boolValue)
}

func TestBoolToInt_shouldReturn0_whenValueNegative(t *testing.T) {
	value := false
	expectedValue := 0
	boolValue := BoolToInt(value)
	_assert.Equal(t, expectedValue, boolValue)
}

func TestGetVersionNumber_shouldReturnError_whenCanNotReadVersion(t *testing.T) {
	deploymentVersion := "test"
	expectedVersion := uint(0x0)
	version, err := GetVersionNumber(deploymentVersion)
	_assert.NotNil(t, err)
	_assert.Equal(t, expectedVersion, version)
}

func TestGetVersionNumber_shouldReturnError_whenVersionCorrect(t *testing.T) {
	deploymentVersion := "05"
	expectedVersion := uint(0x5)
	version, err := GetVersionNumber(deploymentVersion)
	_assert.Nil(t, err)
	_assert.Equal(t, expectedVersion, version)
}

func TestMillisToDuration_shouldReturnCorrectDuration(t *testing.T) {
	seconds := 1.5
	millis := int64(seconds * 1000)
	duration := MillisToDuration(millis)
	_assert.Equal(t, int64(1), duration.Seconds)
	_assert.Equal(t, int32(500000000), duration.Nanos)
}

func TestNanosToDuration_shouldReturnCorrectDuration(t *testing.T) {
	seconds := 1.5
	nanos := int64(seconds * 1000000000)
	duration := NanosToDuration(nanos)
	_assert.Equal(t, int64(1), duration.Seconds)
	_assert.Equal(t, int32(500000000), duration.Nanos)
}

func TestNamedResourceLock(t *testing.T) {
	lock := NewNamedResourceLock()
	sharedCounter := NewSharedCounter(0)
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(NumOfWorkersToWaitFor)

	for i := 0; i < ConcurrentRequestsNum; i++ {
		go func() {
			worker := Worker{
				Name:      fmt.Sprintf("Worker-%v", i),
				WaitGroup: waitGroup,
			}
			worker.doSomeProcessing(lock, sharedCounter)
		}()
	}
	waitGroup.Wait()
	result := sharedCounter.GetResultForIterations(NumOfWorkersToWaitFor)
	if NumOfWorkersToWaitFor != result { // expected counter value means no data races
		t.Fatalf("SharedCounter value is %v (expected: %v)", result, NumOfWorkersToWaitFor)
	}
}

func TestStringSliceContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	sub := []string{"a", "d", "c"}
	_assert.False(t, SliceContains(slice, sub...))
	sub = []string{"a", "b", "c", "d"}
	_assert.False(t, SliceContains(slice, sub...))
	sub = []string{"a", "b", "c"}
	_assert.True(t, SliceContains(slice, sub...))
	sub = []string{"a", "b"}
	_assert.True(t, SliceContains(slice, sub...))
}

func TestSubtractFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}

	_assert.Equal(t, []string{"b"}, SubtractFromSlice(slice, "a", "d", "c"))
	_assert.Equal(t, []string{"b", "c"}, SubtractFromSlice(slice, "a"))
	_assert.Equal(t, []string{}, SubtractFromSlice(slice, "a", "b", "c"))
	_assert.Equal(t, []string{"a", "b", "c"}, SubtractFromSlice(slice, ""))
	_assert.Equal(t, []string{"a", "b", "c"}, SubtractFromSlice(slice, "e"))
}

func TestMergeSlices_StringSlices(t *testing.T) {
	source := []string{"a", "b", "c"}
	toMerge := []string{"a", "d", "c"}
	expectedSlice := []string{"a", "b", "c", "d"}
	mergedSlice := MergeStringSlices(source, toMerge)
	sort.Slice(mergedSlice, func(i, j int) bool {
		return mergedSlice[i] < mergedSlice[j]
	})
	_assert.Equal(t, expectedSlice, mergedSlice)
}

func TestMergeSlices_HeaderMatcherSlice(t *testing.T) {
	source := []*domain.HeaderMatcher{
		{
			Name:       "header1",
			ExactMatch: "some1",
		},
		{
			Name:       "header2",
			ExactMatch: "some2",
		},
		{
			Name:       "header3",
			ExactMatch: "some3",
		},
	}
	toMerge := []*domain.HeaderMatcher{
		{
			Name:       "header1",
			ExactMatch: "some1",
		},
		{
			Name:       "header3",
			ExactMatch: "some3",
		},
		{
			Name:       "header5",
			ExactMatch: "some5",
		},
	}
	expectedSlice := []*domain.HeaderMatcher{
		{
			Name:       "header1",
			ExactMatch: "some1",
		},
		{
			Name:       "header2",
			ExactMatch: "some2",
		},
		{
			Name:       "header3",
			ExactMatch: "some3",
		},
		{
			Name:       "header5",
			ExactMatch: "some5",
		},
	}
	_assert.Equal(t, expectedSlice, MergeHeaderMatchersSlices(source, toMerge))
}

func TestMergeSlices_HeaderSlice(t *testing.T) {
	source := []domain.Header{
		{
			Name:  "header1",
			Value: "some1",
		},
		{
			Name:  "header2",
			Value: "some2",
		},
		{
			Name:  "header3",
			Value: "some3",
		},
	}
	toMerge := []domain.Header{
		{
			Name:  "header1",
			Value: "some1",
		},
		{
			Name:  "header5",
			Value: "some5",
		},
		{
			Name:  "header3",
			Value: "some3",
		},
	}
	expectedSlice := []domain.Header{
		{
			Name:  "header1",
			Value: "some1",
		},
		{
			Name:  "header2",
			Value: "some2",
		},
		{
			Name:  "header3",
			Value: "some3",
		},
		{
			Name:  "header5",
			Value: "some5",
		},
	}
	_assert.Equal(t, expectedSlice, MergeHeaderSlices(source, toMerge))
}

func TestMergeSlices_VirtualHostDomainsSlice(t *testing.T) {
	source := []*domain.VirtualHostDomain{
		{
			Domain:        "domain1",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain5",
			VirtualHostId: 2,
		},
		{
			Domain:        "domain3",
			VirtualHostId: 1,
		},
	}
	toMerge := []*domain.VirtualHostDomain{
		{
			Domain:        "domain3",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain5",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain6",
			VirtualHostId: 1,
		},
	}
	expectedSlice := []*domain.VirtualHostDomain{
		{
			Domain:        "domain1",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain5",
			VirtualHostId: 2,
		},
		{
			Domain:        "domain3",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain5",
			VirtualHostId: 1,
		},
		{
			Domain:        "domain6",
			VirtualHostId: 1,
		},
	}
	_assert.Equal(t, expectedSlice, MergeVirtualHostDomainsSlices(source, toMerge))
}

// Shared non-thread-safe resource
type SharedCounter struct {
	// val is an array where index is order number of the Inc() call and value is the result of this Inc() call.
	val             []int
	currentPosition int
}

func NewSharedCounter(initVal int) *SharedCounter {
	arr := make([]int, ConcurrentRequestsNum)
	arr[0] = initVal
	return &SharedCounter{val: arr, currentPosition: 1}
}

func (c *SharedCounter) Inc() {
	c.val[c.currentPosition] = c.val[c.currentPosition-1] + 1
	c.currentPosition++
}

func (c *SharedCounter) GetResultForIterations(iterationsNum int) int {
	return c.val[iterationsNum]
}

type Worker struct {
	Name      string
	WaitGroup *sync.WaitGroup
}

func (w Worker) doSomeProcessing(lock *NamedResourceLock, sharedCounter *SharedCounter) {
	lock.Lock(TestResourceName)
	defer lock.Unlock(TestResourceName)

	logger.Infof("Worker %v starts processing", w.Name)
	time.Sleep(TestProcessingTime) // we sleep = we doing some work for TestProcessingTime duration
	sharedCounter.Inc()            // increment shared non-thread-safe counter
	logger.Infof("Worker %v ended processing", w.Name)
	w.WaitGroup.Done()
}

func TestNormalizeYaml1(t *testing.T) {
	assert := _assert.New(t)
	input := `  
---
abc: cde
--- 
foo: bar

`
	actual, err := NormalizeJsonOrYamlInput(input)
	assert.NoError(err)
	assert.Equal(`[{"abc":"cde"},{"foo":"bar"}]`, string(actual))
}

func TestNormalizeYaml2(t *testing.T) {
	assert := _assert.New(t)
	input := `  
abc: "cde ---"
`
	actual, err := NormalizeJsonOrYamlInput(input)
	assert.NoError(err)
	assert.Equal(`[{"abc":"cde ---"}]`, string(actual))
}

func TestNormalizeJson1(t *testing.T) {
	assert := _assert.New(t)
	input := `
{
	"abc": "cde ---"
}
`
	actual, err := NormalizeJsonOrYamlInput(input)
	assert.NoError(err)
	assert.Equal("[{\n\t\"abc\": \"cde ---\"\n}]", string(actual))
}

func TestNormalizeJson2(t *testing.T) {
	assert := _assert.New(t)
	input := `
[{
	"abc": "cde ---"
}]
`
	actual, err := NormalizeJsonOrYamlInput(input)
	assert.NoError(err)
	assert.Equal("[{\n\t\"abc\": \"cde ---\"\n}]", string(actual))
}

func TestNormalizeJson_IgnoreEmptySections(t *testing.T) {
	assert := _assert.New(t)
	input := `
# some text 
---
abc: cde
`
	actual, err := NormalizeJsonOrYamlInput(input)
	assert.NoError(err)
	assert.Equal("[{\"abc\":\"cde\"}]", string(actual))
}

func TestWrapValue(t *testing.T) {
	assert := _assert.New(t)

	assert.Equal(true, *WrapValue(true))
	assert.Equal(11, *WrapValue(11))
	assert.Equal("test", *WrapValue("test"))
}
