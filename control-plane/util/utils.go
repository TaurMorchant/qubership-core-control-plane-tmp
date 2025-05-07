package util

import (
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/netcracker/qubership-core-control-plane/control-plane/v2/domain"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	appConfigSync   sync.Once
	yamlDelimiter   = regexp.MustCompile(`(?m)(^---\s*)`)
	TlsPort         = 8443
	DefaultPort     = 8080
	DefaultProtocol = "http"
	TlsProtocol     = "https"
)

func ToSlice(slice interface{}) []interface{} {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}
	result := make([]interface{}, s.Len())
	for i := 0; i < s.Len(); i++ {
		result[i] = s.Index(i).Interface()
	}
	return result
}

func SliceContainsElement[T comparable](arr []T, elem T) bool {
	for _, str := range arr {
		if str == elem {
			return true
		}
	}
	return false
}

func MapKeysToSlice[K comparable, V any](m map[K]V) []K {
	keys := make([]K, len(m))

	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

type NamedResourceLock struct {
	lockedNames *sync.Map
}

func NewNamedResourceLock() *NamedResourceLock {
	return &NamedResourceLock{lockedNames: &sync.Map{}}
}

func (namedResLock *NamedResourceLock) Lock(name string) {
	_, alreadyExisted := namedResLock.lockedNames.LoadOrStore(name, name)
	for alreadyExisted {
		time.Sleep(100 * time.Millisecond)
		_, alreadyExisted = namedResLock.lockedNames.LoadOrStore(name, name)
	}
}

func (namedResLock *NamedResourceLock) Unlock(name string) {
	namedResLock.lockedNames.Delete(name)
}

type AtomicCounter struct {
	val int
	mu  sync.RWMutex
}

func NewAtomicCounter(initialValue int) *AtomicCounter {
	return &AtomicCounter{val: initialValue, mu: sync.RWMutex{}}
}

func (counter *AtomicCounter) Value() int {
	counter.mu.RLock()
	defer counter.mu.RUnlock()

	return counter.val
}

func (counter *AtomicCounter) AddAndReturn(delta int) int {
	counter.mu.Lock()
	defer counter.mu.Unlock()

	counter.val += delta
	return counter.val
}

func BoolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func SliceContains[T comparable](slice []T, sub ...T) bool {
	if len(sub) == 0 {
		return true
	}
	if len(slice) < len(sub) {
		return false
	}
	for _, subStr := range sub {
		if !SliceContainsElement(slice, subStr) {
			return false
		}
	}
	return true
}

func SubtractFromSlice(slice []string, elements ...string) []string {
	result := make([]string, 0, len(slice)-len(elements))
	for _, sliceElement := range slice {
		if !SliceContainsElement(elements, sliceElement) {
			result = append(result, sliceElement)
		}
	}
	return result
}

func MergeStringSlices(sourceSlice, sliceToMerge []string) []string {
	elemMap := make(map[string]bool)
	for _, sourceElem := range sourceSlice {
		elemMap[sourceElem] = true
	}
	for _, elemToMerge := range sliceToMerge {
		elemMap[elemToMerge] = true
	}
	mergedSlice := make([]string, 0, len(elemMap))
	for elem, _ := range elemMap {
		mergedSlice = append(mergedSlice, elem)
	}
	return mergedSlice
}

func MergeHeaderMatchersSlices(sourceSlice, sliceToMerge []*domain.HeaderMatcher) []*domain.HeaderMatcher {
	for _, elemToAdd := range sliceToMerge {
		isContains := false
		for _, sourceElem := range sourceSlice {
			if elemToAdd.Equals(sourceElem) {
				isContains = true
				break
			}
		}
		if !isContains {
			sourceSlice = append(sourceSlice, elemToAdd)
		}
	}
	return sourceSlice
}

func MergeHeaderSlices(sourceSlice, sliceToMerge []domain.Header) []domain.Header {
	for _, elemToAdd := range sliceToMerge {
		isContains := false
		for _, sourceElem := range sourceSlice {
			if elemToAdd.Equals(sourceElem) {
				isContains = true
				break
			}
		}
		if !isContains {
			sourceSlice = append(sourceSlice, elemToAdd)
		}
	}
	return sourceSlice
}

func MergeVirtualHostDomainsSlices(sourceSlice, sliceToMerge []*domain.VirtualHostDomain) []*domain.VirtualHostDomain {
	for _, elemToAdd := range sliceToMerge {
		isContains := false
		for _, sourceElem := range sourceSlice {
			if elemToAdd.Equals(sourceElem) {
				isContains = true
				break
			}
		}
		if !isContains {
			sourceSlice = append(sourceSlice, elemToAdd)
		}
	}
	return sourceSlice
}

func NormalizeJsonOrYamlInput(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "[") {
		// it's JSON and already in desired format
		return trimmed, nil
	} else if strings.HasPrefix(trimmed, "{") {
		// to json array
		return "[" + trimmed + "]", nil
	} else {
		reformatted := &strings.Builder{}
		reformatted.WriteString("[")
		needComma := false
		for _, chunk := range yamlDelimiter.Split(raw, -1) {
			if len(strings.TrimSpace(chunk)) == 0 {
				// ignore empty chinks
				continue
			}

			if needComma {
				reformatted.WriteString(",")
			}

			if json, err := yaml.YAMLToJSON([]byte(chunk)); err == nil {
				if string(json) != "null" { // ignore commented sections
					reformatted.Write(json)
					needComma = true
				}
			} else {
				return "", err
			}
		}
		reformatted.WriteString("]")

		return reformatted.String(), nil
	}
}

func GetVersionNumber(deploymentVersion string) (uint, error) {
	ver, err := strconv.Atoi(deploymentVersion[1:])
	if err != nil {
		return 0, err
	}
	return uint(ver), nil
}

func MillisToDuration(millis int64) *duration.Duration {
	return NanosToDuration(millis * 1e6)
}

func NanosToDuration(nanos int64) *duration.Duration {
	secs := nanos / 1e9
	nanos -= secs * 1e9
	return &duration.Duration{Seconds: secs, Nanos: int32(nanos)}
}

type DefaultRetryProvider struct{}

func (d DefaultRetryProvider) SleepPeriodOnSnapshotSend() time.Duration {
	return 5 * time.Second
}

func (d DefaultRetryProvider) SleepPeriodGetSnapshot() time.Duration {
	return 3 * time.Second
}

func (d DefaultRetryProvider) AttemptAmountGetSnapshot() int {
	return 10
}

func (d DefaultRetryProvider) DeferredMessagesTTL() time.Duration {
	return time.Hour
}

func (d DefaultRetryProvider) SleepPeriodSubscribe() time.Duration {
	return 100 * time.Millisecond
}

func SliceToSet[T comparable](slice []T) map[T]bool {
	result := make(map[T]bool, len(slice))
	for _, el := range slice {
		result[el] = true
	}
	return result
}

func SliceToMap[T any, K comparable, V any](slice []T, keyMapper func(element T) K, valueMapper func(element T) V) map[K]V {
	result := make(map[K]V, len(slice))
	for _, element := range slice {
		result[keyMapper(element)] = valueMapper(element)
	}
	return result
}

var dns1123LabelRegexp = regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")

// DNS1123LabelMaxLength is a label's max length in DNS (RFC 1123)
const DNS1123LabelMaxLength int = 63

// IsDNS1123Label tests for a string that conforms to the definition of a label in
// DNS (RFC 1123).
func IsDNS1123Label(value string) error {
	if len(value) > DNS1123LabelMaxLength {
		return errors.New(fmt.Sprintf("value length must not be greater than that %d", DNS1123LabelMaxLength))
	}
	if !dns1123LabelRegexp.MatchString(value) {
		return errors.New("value must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}
	return nil
}

func WrapValue[T any](value T) *T {
	return &value
}
