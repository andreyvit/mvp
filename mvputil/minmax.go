package mvputil

import "time"

// TODO: use cmp.Ordered in Go.1.21
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func BumpMin[T Ordered](d *T, s T) {
	if s < *d {
		*d = s
	}
}
func BumpMinIfNonZero[T Ordered](d *T, s T) {
	var zero T
	if s != zero && s < *d {
		*d = s
	}
}
func BumpMax[T Ordered](d *T, s T) {
	if s > *d {
		*d = s
	}
}
func BumpMaxIfNonZero[T Ordered](d *T, s T) {
	var zero T
	if s != zero && s > *d {
		*d = s
	}
}

func BumpMinTime(d *time.Time, s time.Time) {
	if !s.IsZero() && s.Before(*d) {
		*d = s
	}
}
func BumpMaxTime(d *time.Time, s time.Time) {
	if !s.IsZero() && s.After(*d) {
		*d = s
	}
}
