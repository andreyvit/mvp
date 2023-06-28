package jsonext

import (
	"bytes"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/vmihailenco/msgpack/v5/msgpcode"
)

type OptTime time.Time

func (v OptTime) Value() time.Time {
	return time.Time(v)
}

func (v OptTime) IsZero() bool {
	return v.Value().IsZero()
}

func (v OptTime) Before(r OptTime) bool {
	return v.Value().Before(r.Value())
}

func (v OptTime) BeforeTime(r time.Time) bool {
	return v.Value().Before(r)
}

func (v OptTime) MarshalJSON() ([]byte, error) {
	if time.Time(v).IsZero() {
		return nullb, nil
	} else {
		return time.Time(v).MarshalJSON()
	}
}

func (v *OptTime) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || bytes.Equal(b, nullb) {
		*v = OptTime{}
		return nil
	} else {
		return (*time.Time)(v).UnmarshalJSON(b)
	}
}

var _ msgpack.CustomEncoder = (*OptTime)(nil)

func (v *OptTime) EncodeMsgpack(enc *msgpack.Encoder) error {
	if (*time.Time)(v).IsZero() {
		return enc.EncodeNil()
	} else {
		return enc.EncodeTime(time.Time(*v))
	}
}

var _ msgpack.CustomDecoder = (*OptTime)(nil)

func (v *OptTime) DecodeMsgpack(dec *msgpack.Decoder) error {
	code, err := dec.PeekCode()
	if err != nil {
		return err
	}
	if code == msgpcode.Nil {
		*v = OptTime{}
		return dec.DecodeNil()
	} else {
		*(*time.Time)(v), err = dec.DecodeTime()
		return err
	}
}
