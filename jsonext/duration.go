package jsonext

import (
	"time"
)

type Duration time.Duration

func (v Duration) String() string {
	return time.Duration(v).String()
}

func (v Duration) Value() time.Duration {
	return time.Duration(v)
}

func (v Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(v).String()), nil
}

func (v *Duration) UnmarshalText(b []byte) error {
	d, err := time.ParseDuration(string(b))
	*v = Duration(d)
	return err
}
