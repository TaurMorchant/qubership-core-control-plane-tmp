package domain

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
)

type NullString struct {
	sql.NullString
}

type NullBool struct {
	sql.NullBool
}

type NullInt struct {
	sql.NullInt64
}

func NewNullBool(value bool) NullBool {
	return NullBool{NullBool: sql.NullBool{Bool: value, Valid: true}}
}

func NewNullString(str string) NullString {
	if len(str) == 0 {
		return NullString{NullString: sql.NullString{}}
	}
	return NullString{NullString: sql.NullString{String: str, Valid: true}}
}

func NewNullInt(value int64) NullInt {
	return NullInt{NullInt64: sql.NullInt64{Int64: value, Valid: true}}
}

func ToNullsHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	switch f.Kind() {
	case reflect.Bool:
		{
			var b sql.NullBool
			if t == reflect.TypeOf(b) {
				b = sql.NullBool{
					Bool:  data.(bool),
					Valid: true,
				}
				return b, nil
			}

			var nb NullBool
			if t == reflect.TypeOf(nb) {
				nb = NullBool{
					sql.NullBool{
						Bool:  data.(bool),
						Valid: true,
					},
				}
				return nb, nil
			}

			return data, nil
		}

	case reflect.String:
		{
			var s sql.NullString
			if t == reflect.TypeOf(s) {
				s = sql.NullString{
					String: data.(string),
					Valid:  true,
				}
				return s, nil
			}

			var ns NullString
			if t == reflect.TypeOf(ns) {
				ns = NullString{
					sql.NullString{
						String: data.(string),
						Valid:  true,
					},
				}
				return ns, nil
			}

			return data, nil
		}

	case reflect.Int64:
		{
			var i sql.NullInt64
			if t == reflect.TypeOf(i) {
				i = sql.NullInt64{
					Int64: data.(int64),
					Valid: true,
				}
				return i, nil
			}

			var ni NullInt
			if t == reflect.TypeOf(ni) {
				ni = NullInt{
					sql.NullInt64{
						Int64: data.(int64),
						Valid: true,
					},
				}
				return ni, nil
			}

			return data, nil
		}

	default:
		return data, nil
	}
}

func (ns *NullInt) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.Int64)
}

func (ns *NullInt) UnmarshalJSON(b []byte) error {
	var i interface{}
	err := json.Unmarshal(b, &i)
	switch v := i.(type) {
	case float64:
		*ns = NewNullInt(int64(v))
	case sql.NullInt64:
		ns.Int64 = v.Int64
		ns.Valid = v.Valid
	}
	return err
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullBool) UnmarshalJSON(b []byte) error {
	if b == nil || string(b) == "null" {
		ns.Valid = false
		return nil
	}
	err := json.Unmarshal(b, &ns.Bool)
	ns.Valid = err == nil
	return err
}

func (ns *NullBool) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.Bool)
}

func (ns NullBool) String() string {
	if ns.Valid {
		return fmt.Sprintf("%t", ns.Bool)
	} else {
		return "<nil>"
	}
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ns.String)
	ns.Valid = (err == nil)
	return err
}
