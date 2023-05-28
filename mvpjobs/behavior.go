package mvpjobs

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type Behavior int

const (
	Idempotent = Behavior(0)
	Repeatable = Behavior(1)
	Cron       = Behavior(2)
)

var _behaviorStrings = []string{
	"idempotent",
	"repeatable",
	"cron",
}

func (v Behavior) IsRepeatable() bool {
	return v == Repeatable || v == Cron
}

func (v Behavior) String() string {
	return _behaviorStrings[v]
}

func ParseBehavior(s string) (Behavior, error) {
	if i := slices.Index(_behaviorStrings, s); i >= 0 {
		return Behavior(i), nil
	} else {
		return Idempotent, fmt.Errorf("invalid Behavior %q", s)
	}
}

func (v Behavior) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Behavior) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseBehavior(string(b))
	return err
}
func (v Behavior) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *Behavior) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = Behavior(n)
	return err
}
