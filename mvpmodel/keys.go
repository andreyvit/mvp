package mvpm

import (
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

type NamedKeySet struct {
	Keys          map[string][]byte
	ActiveKeyName string
}

func (ks *NamedKeySet) ActiveKey() []byte {
	return ks.Keys[ks.ActiveKeyName]
}

func ParseKeys(s string) ([][]byte, error) {
	var keys [][]byte
	for _, ks := range strings.FieldsFunc(s, isWhitespaceOrComma) {
		if ks == "" {
			continue
		}
		key, err := hex.DecodeString(ks)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

type Keys [][]byte

func KeysVar(v *[][]byte) *Keys {
	return (*Keys)(v)
}

func (v Keys) String() string {
	var buf strings.Builder
	for i, k := range v {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(hex.EncodeToString(k))
	}
	return buf.String()
}

func (v Keys) Get() interface{} {
	return [][]byte(v)
}

func (v *Keys) Set(raw string) (err error) {
	*v, err = ParseKeys(raw)
	return
}

func (v *Keys) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if b[0] == '[' {
		var strs []string
		if err := json.Unmarshal(b, &strs); err != nil {
			return err
		}
		*v = make(Keys, 0, len(strs))
		for _, str := range strs {
			key, err := hex.DecodeString(str)
			if err != nil {
				return err
			}
			*v = append(*v, key)
		}
		return nil
	}
	return fmt.Errorf("keys must be an array")
}

func (v Keys) Match(b []byte) bool {
	for _, k := range v {
		if 1 == subtle.ConstantTimeCompare(k, b) {
			return true
		}
	}
	return false
}

func (v Keys) MatchString(str string) bool {
	return v.Match([]byte(str))
}
