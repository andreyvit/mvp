package mvpjobs

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type Persistence int

const (
	Persistent Persistence = 1
	Ephemeral  Persistence = 2
)

var _persistenceStrings = []string{
	"",
	"persistent",
	"ephemeral",
}

func (v Persistence) String() string {
	return _persistenceStrings[v]
}

func ParsePersistence(s string) (Persistence, error) {
	if i := slices.Index(_persistenceStrings, s); i >= 0 {
		return Persistence(i), nil
	} else {
		return 0, fmt.Errorf("invalid Persistence %q", s)
	}
}

func (v Persistence) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Persistence) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParsePersistence(string(b))
	return err
}
func (v Persistence) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *Persistence) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = Persistence(n)
	return err
}
