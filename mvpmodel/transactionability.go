package mvpm

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type StoreAffinity int

const (
	UnknownAffinity StoreAffinity = 0
	SafeReader      StoreAffinity = 1
	SafeWriter      StoreAffinity = 2
	ExclusiveWriter StoreAffinity = 3
	Manual          StoreAffinity = 4
	DBUnused        StoreAffinity = 5
)

var _storeAffinityStrings = []string{
	"unknown",
	"safe-reader",
	"safe-writer",
	"exclusive-writer",
	"manual",
	"db-unused",
}

func (v StoreAffinity) IsWriter() bool {
	return v == SafeWriter || v == ExclusiveWriter
}
func (v StoreAffinity) IsReader() bool {
	return v == SafeReader
}
func (v StoreAffinity) WantsAutomaticTx() bool {
	return v == SafeReader || v == SafeWriter || v == ExclusiveWriter
}

func (v StoreAffinity) String() string {
	return _storeAffinityStrings[v]
}

func ParseStoreAffinity(s string) (StoreAffinity, error) {
	if i := slices.Index(_storeAffinityStrings, s); i >= 0 {
		return StoreAffinity(i), nil
	} else {
		return DBUnused, fmt.Errorf("invalid StoreAffinity %q", s)
	}
}

func (v StoreAffinity) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *StoreAffinity) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseStoreAffinity(string(b))
	return err
}
