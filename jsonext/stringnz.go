package jsonext

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// StringNonZero is a string that can also be unmarshalled from a JSON integer,
// converting integer 0 into an empty string.
type StringNonZero string

func (v *StringNonZero) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, nullb) {
		return nil
	}
	origErr := json.Unmarshal(b, (*string)(v))
	if origErr != nil {
		var i int64
		if err := json.Unmarshal(b, &i); err == nil {
			if i == 0 {
				*v = ""
			} else {
				*v = StringNonZero(strconv.FormatInt(i, 10))
			}
			return nil
		}
	}
	return origErr
}

func (v StringNonZero) IsZero() bool {
	return v == ""
}
func (v StringNonZero) String() string {
	return string(v)
}
