package monitoring

import (
	"fmt"
	"sync/atomic"
	"time"
)

const throttlingInterval = time.Hour

var (
	notices = make(map[uint64]*Notice)
)

type (
	Notice struct {
		ordinal    int
		id         uint64
		name       string
		paramNames []string
	}

	NoticeState struct {
		last atomic.Int64
	}
)

func DefineNotice(id uint64, name string, paramNames []string) *Notice {
	if runtimeCreated.Load() {
		panic("cannot define a notice after creating a runtime")
	}
	if len(paramNames) != 0 {
		panic("notice params not implemented yet")
	}
	ord := len(notices)
	n := &Notice{ord, id, name, paramNames}
	if notices[id] != nil {
		panic(fmt.Errorf("duplicate notice id %x: %q and %q", id, name, notices[id].name))
	}
	notices[id] = n
	return n
}

func (n *Notice) Name() string {
	return n.name
}
func (n *Notice) String() string {
	return n.name
}

// TODO: , params ...uint64
func (rt *Runtime) AllowNotice(n *Notice) bool {
	state := &rt.noticeStates[n.ordinal]
	for {
		last := state.last.Load()
		now := runtimeNano()
		if last != 0 && now < last+int64(throttlingInterval) && !rt.verbose {
			return false
		}
		if state.last.CompareAndSwap(last, now) {
			return true
		}
		time.Sleep(time.Nanosecond)
	}
}
