package mvpsecrets

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"strings"
)

type ByteStrings [][]byte

func ByteStringsVar(v *[][]byte) *ByteStrings {
	return (*ByteStrings)(v)
}

func (v ByteStrings) String() string {
	var buf strings.Builder
	for i, k := range v {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(string(k))
	}
	return buf.String()
}

func (v *ByteStrings) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if b[0] == '[' {
		var strs []string
		if err := json.Unmarshal(b, &strs); err != nil {
			return err
		}
		*v = make(ByteStrings, 0, len(strs))
		for _, str := range strs {
			*v = append(*v, []byte(str))
		}
		return nil
	}
	return fmt.Errorf("keys must be an array")
}

func (v ByteStrings) Match(b []byte) bool {
	for _, k := range v {
		if 1 == subtle.ConstantTimeCompare(k, b) {
			return true
		}
	}
	return false
}

func (v ByteStrings) MatchString(str string) bool {
	return v.Match([]byte(str))
}
