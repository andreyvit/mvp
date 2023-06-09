package flake

import (
	"testing"
	"time"
)

func TestGen(t *testing.T) {
	zerro := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	onnne := time.Date(2020, 1, 1, 0, 0, 0, int(time.Millisecond), time.UTC)
	twooo := time.Date(2020, 1, 1, 0, 0, 0, 2*int(time.Millisecond), time.UTC)
	prsnt := time.Date(2022, 9, 1, 0, 0, 0, int(time.Millisecond), time.UTC)

	g := NewGen(0, 0x42)
	try(t, g, zerro, 0x0000000001420001)
	try(t, g, zerro, 0x0000000001420002)
	try(t, g, onnne, 0x0000000001420003)
	try(t, g, onnne, 0x0000000001420004)
	try(t, g, twooo, 0x0000000002420001)
	try(t, g, prsnt, 0x1397f20801420001)
}

func try(t *testing.T, g *Gen, tm time.Time, e ID) {
	t.Helper()
	a := g.NewAt(tm)
	if a != e {
		t.Errorf("NewAt(%v) = %v, wanted %v  (ms at that time = %X)", tm, a, e, MillisAt(tm))
	}
	if at := a.Time(); at.Unix() != tm.Unix() {
		t.Errorf("%v.Time() = %v, wanted %v", a, at, tm)
	}
}
