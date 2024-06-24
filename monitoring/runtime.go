package monitoring

import "sync/atomic"

type Runtime struct {
	verbose      bool
	noticeStates []NoticeState
}

var runtimeCreated atomic.Bool

func NewRuntime(verbose bool) *Runtime {
	runtimeCreated.Store(true)
	return &Runtime{
		verbose:      verbose,
		noticeStates: make([]NoticeState, len(notices)),
	}
}
