package mvpjobs

import (
	"fmt"

	"golang.org/x/exp/slices"
)

type Status int

const (
	StatusNone Status = iota
	StatusQueued
	StatusBlocked
	StatusRunning
	StatusRunningPending // for re-entrant jobs
	StatusWaiting
	StatusRetrying
	StatusDone
	StatusFailed
	StatusFailedSkipped
)

var _statusStrings = []string{
	"",
	"queued",
	"blocked",
	"running",
	"running-pending",
	"waiting",
	"retrying",
	"done",
	"failed",
	"skipped",
}

func (s Status) IsPending() bool {
	return s == StatusQueued || s == StatusRetrying
}

func (s Status) IsRunning() bool {
	return s == StatusRunning || s == StatusRunningPending
}

func (s Status) IsTerminal() bool {
	return s == StatusDone || s == StatusFailed || s == StatusFailedSkipped
}

func (v Status) String() string {
	return _statusStrings[v]
}

func ParseStatus(s string) (Status, error) {
	if i := slices.Index(_statusStrings, s); i >= 0 {
		return Status(i), nil
	} else {
		return StatusNone, fmt.Errorf("invalid Status %q", s)
	}
}

func (v Status) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Status) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseStatus(string(b))
	return err
}
