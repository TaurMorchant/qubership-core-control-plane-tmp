package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type TestStruct struct {
	Field NullBool
}

func TestNullBool_MarshalJSON(t *testing.T) {
	nullValue := NewNullBool(true)
	assert.True(t, nullValue.Valid)
	assert.True(t, nullValue.Bool)
	result, err := json.Marshal(&TestStruct{nullValue})
	assert.Nil(t, err)
	assert.Equal(t, "{\"Field\":true}", string(result))

	nullValue = NewNullBool(false)
	assert.True(t, nullValue.Valid)
	assert.False(t, nullValue.Bool)
	result, err = json.Marshal(&TestStruct{nullValue})
	assert.Nil(t, err)
	assert.Equal(t, "{\"Field\":false}", string(result))

	nullValue = NewNullBool(false)
	nullValue.Valid = false
	result, err = json.Marshal(&TestStruct{nullValue})
	assert.Nil(t, err)
	assert.Equal(t, "{\"Field\":null}", string(result))

	nullValue = NewNullBool(true)
	nullValue.Valid = false
	result, err = json.Marshal(&TestStruct{nullValue})
	assert.Nil(t, err)
	assert.Equal(t, "{\"Field\":null}", string(result))
}

func TestNullBool_UnmarshalJSON(t *testing.T) {
	var val1 NullBool
	err := json.Unmarshal([]byte("true"), &val1)
	assert.Nil(t, err)
	assert.True(t, val1.Valid)
	assert.True(t, val1.Bool)

	var val2 NullBool
	err = json.Unmarshal([]byte("false"), &val2)
	assert.Nil(t, err)
	assert.True(t, val2.Valid)
	assert.False(t, val2.Bool)

	var val3 NullBool
	err = json.Unmarshal([]byte("\"false\""), &val3)
	assert.NotNil(t, err)

	var val4 NullBool
	err = json.Unmarshal([]byte("\"null\""), &val4)
	assert.NotNil(t, err)

	var val5 NullBool
	err = json.Unmarshal([]byte("null"), &val5)
	assert.Nil(t, err)
	assert.False(t, val5.Valid)
}

func TestToNullsHookFunc(t *testing.T) {
	var nbType = reflect.TypeOf(NullBool{})
	var snbType = reflect.TypeOf(sql.NullBool{})

	var boolType = reflect.TypeOf(true)
	res, err := ToNullsHookFunc(boolType, nbType, true)
	assert.Nil(t, err)
	assert.Equal(t, nbType, reflect.TypeOf(res))

	res, err = ToNullsHookFunc(boolType, nbType, false)
	assert.Nil(t, err)
	assert.Equal(t, nbType, reflect.TypeOf(res))

	res, err = ToNullsHookFunc(boolType, snbType, true)
	assert.Nil(t, err)
	assert.Equal(t, snbType, reflect.TypeOf(res))

	res, err = ToNullsHookFunc(boolType, snbType, false)
	assert.Nil(t, err)
	assert.Equal(t, snbType, reflect.TypeOf(res))

	var nsType = reflect.TypeOf(NullString{})
	var snsType = reflect.TypeOf(sql.NullString{})

	var stringVal = "string"
	var stringType = reflect.TypeOf(stringVal)
	res, err = ToNullsHookFunc(stringType, nsType, stringVal)
	assert.Nil(t, err)
	assert.Equal(t, nsType, reflect.TypeOf(res))

	res, err = ToNullsHookFunc(stringType, snsType, stringVal)
	assert.Nil(t, err)
	assert.Equal(t, snsType, reflect.TypeOf(res))

	var niType = reflect.TypeOf(NullInt{})
	var sniType = reflect.TypeOf(sql.NullInt64{})

	var int64Val = int64(10)
	var int64Type = reflect.TypeOf(int64Val)
	res, err = ToNullsHookFunc(int64Type, niType, int64Val)
	assert.Nil(t, err)
	assert.Equal(t, niType, reflect.TypeOf(res))

	res, err = ToNullsHookFunc(int64Type, sniType, int64Val)
	assert.Nil(t, err)
	assert.Equal(t, sniType, reflect.TypeOf(res))
}

func TestNewNullString_shouldNotValid_whenStringEmpty(t *testing.T) {
	testString := ""
	nullString := NewNullString(testString)
	assert.Equal(t, false, nullString.Valid)
	assert.Equal(t, testString, nullString.String)
}

func TestNewNullString_shouldValid_whenStringNotEmpty(t *testing.T) {
	testString := "12345"
	nullString := NewNullString(testString)
	assert.Equal(t, true, nullString.Valid)
	assert.Equal(t, testString, nullString.String)
}

func TestNewNullInt_shouldValid_whenValueIsZero(t *testing.T) {
	testInt := int64(0)
	nullInt64 := NewNullInt(testInt)
	assert.Equal(t, true, nullInt64.Valid)
	assert.Equal(t, testInt, nullInt64.Int64)
}

func TestNewNullInt_shouldValid_whenValueNotZero(t *testing.T) {
	testInt := int64(1)
	nullInt64 := NewNullInt(testInt)
	assert.Equal(t, true, nullInt64.Valid)
	assert.Equal(t, testInt, nullInt64.Int64)
}

func TestNullStringMarshalJSON_shouldExpectedValue_whenValueNotNull(t *testing.T) {
	expectedResult := "test value"
	nullString := NullString{NullString: sql.NullString{String: expectedResult, Valid: true}}

	value, err := nullString.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, len(expectedResult)+2, len(value))
	assert.Equal(t, fmt.Sprintf("\"%s\"", expectedResult), string(value))
}

func TestNullStringMarshalJSON_shouldNull_whenValueIsZero(t *testing.T) {
	expectedResult := "null"
	nullString := NullString{NullString: sql.NullString{}}

	value, err := nullString.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, len(expectedResult), len(value))
	assert.Equal(t, expectedResult, string(value))
}

func TestNullStringUnmarshalJSON_shouldReplace_whenSecondValueValid(t *testing.T) {
	expectedResult := "test value"
	nullStringFirst := NullString{NullString: sql.NullString{String: expectedResult, Valid: true}}
	nullStringSecond := NullString{NullString: sql.NullString{String: expectedResult, Valid: true}}
	value, _ := nullStringSecond.MarshalJSON()

	err := nullStringFirst.UnmarshalJSON(value)
	assert.Nil(t, err)
	assert.Equal(t, expectedResult, nullStringFirst.String)
}

func TestNullInt64MarshalJSON_shouldNull_whenValueIsZero(t *testing.T) {
	expectedResult := "null"
	nullInt64 := NullInt{NullInt64: sql.NullInt64{Int64: int64(0), Valid: false}}

	value, err := nullInt64.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, len(expectedResult), len(value))
	assert.Equal(t, expectedResult, string(value))
}

func TestNullInt64MarshalJSON_shouldNotNull_whenValueNotZero(t *testing.T) {
	expectedResult := "1"
	nullInt64 := NullInt{NullInt64: sql.NullInt64{Int64: int64(1), Valid: true}}

	value, err := nullInt64.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, len(expectedResult), len(value))
	assert.Equal(t, expectedResult, string(value))
}

func TestNullInt64UnmarshalJSON_shouldReplace_whenSecondValueValid(t *testing.T) {
	nullInt64First := NullInt{NullInt64: sql.NullInt64{Int64: int64(1), Valid: true}}
	nullInt64Second := NullInt{NullInt64: sql.NullInt64{Int64: int64(2), Valid: true}}
	secondValue, _ := nullInt64Second.MarshalJSON()
	err := nullInt64First.UnmarshalJSON(secondValue)
	assert.Nil(t, err)
	assert.Equal(t, nullInt64Second.Int64, nullInt64First.Int64)
}

func TestNullInt64UnmarshalJSON_shouldNotReplace_whenSecondValueNull(t *testing.T) {
	nullInt64First := NullInt{NullInt64: sql.NullInt64{Int64: int64(1), Valid: true}}
	err := nullInt64First.UnmarshalJSON([]byte("null"))
	assert.Nil(t, err)
	assert.Equal(t, int64(1), nullInt64First.Int64)
}

func TestNullBool_shouldExpectedValue_whenNilBoolValid(t *testing.T) {
	nullBool := NewNullBool(false)
	stringBool := nullBool.String()
	assert.Equal(t, "false", stringBool)
}

func TestNullBool_shouldNil_whenNilBoolNotValid(t *testing.T) {
	nullBool := NullBool{NullBool: sql.NullBool{Bool: true, Valid: false}}
	stringBool := nullBool.String()
	assert.Equal(t, "<nil>", stringBool)
}
