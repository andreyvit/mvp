package mvp

import (
	"fmt"

	"golang.org/x/exp/maps"
)

type CacheBustingSession struct {
	Keys   []any
	Busted []bool
}

func (rc *RC) BustCache(keys ...any) {
	if rc.IsInWriteTx() {
		if rc.cacheBusting == nil {
			rc.cacheBusting = make(map[any]struct{})
		}
		for _, key := range keys {
			rc.cacheBusting[key] = struct{}{}
		}
	} else {
		bustCaches(rc, keys)
	}
}

func (rc *RC) applyDelayedCacheBusting() {
	if len(rc.cacheBusting) == 0 {
		return
	}
	keys := maps.Keys(rc.cacheBusting)
	maps.Clear(rc.cacheBusting)
	bustCaches(rc, keys)
}

func bustCaches(rc *RC, keys []any) {
	if len(keys) == 0 {
		return
	}
	for _, key := range keys {
		if !runHooksFwd2Or(rc.app.Hooks.bustCache, rc, key) {
			panic(fmt.Errorf("don't know how to bust cache for key %T %v", key, key))
		}
	}
	rc.DoneReading()
}
