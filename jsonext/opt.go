package jsonext

import (
	"bytes"
	"encoding/json"
)

var nullb = []byte("null")

type Opt[T any] struct {
	Valid bool
	Value T
}

func (v Opt[T]) MarshalJSON() ([]byte, error) {
	if v.Valid {
		return json.Marshal(v.Value)
	} else {
		return nullb, nil
	}
}

func (v *Opt[T]) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, nullb) {
		*v = Opt[T]{}
		return nil
	} else {
		v.Valid = true
		return json.Unmarshal(b, &v.Value)
	}
}
