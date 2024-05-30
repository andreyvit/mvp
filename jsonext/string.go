package jsonext

import (
	"bytes"
	"encoding/json"
	"strconv"
)

type String string

func (v *String) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, nullb) {
		return nil
	}
	origErr := json.Unmarshal(b, (*string)(v))
	if origErr == nil {
		return nil
	}

	var i int64
	if err := json.Unmarshal(b, &i); err == nil {
		*v = String(strconv.FormatInt(i, 10))
		return nil
	}

	var u uint64
	if err := json.Unmarshal(b, &u); err == nil {
		*v = String(strconv.FormatUint(u, 10))
		return nil
	}

	var f float64
	if err := json.Unmarshal(b, &f); err == nil {
		*v = String(strconv.FormatFloat(f, 'e', -1, 64))
		return nil
	}

	return origErr
}
