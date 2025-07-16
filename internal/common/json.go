package common

// kudos to https://github.com/stephenafamo/bob

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Array[T comparable] []T

func (a Array[T]) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Array[T]) Scan(value any) error {
	switch x := value.(type) {
	case string:
		return json.NewDecoder(bytes.NewBuffer([]byte(x))).Decode(a)
	case []byte:
		return json.NewDecoder(bytes.NewBuffer(x)).Decode(a)
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan type %T: %v", value, value)
	}
}

func (a Array[T]) Includes(val T) bool {
	for _, v := range a {
		if v == val {
			return true
		}
	}
	return false
}

type JSONB struct {
	Data json.RawMessage
}

func (a *JSONB) Scan(value any) error {
	switch x := value.(type) {
	case string:
		return json.NewDecoder(bytes.NewBuffer([]byte(x))).Decode(&a.Data)
	case []byte:
		return json.NewDecoder(bytes.NewBuffer(x)).Decode(&a.Data)
	case nil:
		return nil
	default:
		return fmt.Errorf("cannot scan type %T: %v", value, value)
	}
}

func (a JSONB) Value() (driver.Value, error) {
	return json.Marshal([]byte(a.Data))
}
