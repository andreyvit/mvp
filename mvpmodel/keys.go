package mvpm

import (
	"encoding/hex"
	"strings"
)

type NamedKeySet struct {
	Keys     map[string][]byte
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
