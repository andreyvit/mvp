package mvp

import (
	"fmt"
	"runtime/debug"
)

type SkipStackOnPanic interface {
	SkipStackOnPanic() bool
}

type Panic struct {
	cause error
	stack []byte
}

func NewPanic(errValue any) error {
	var err error
	if e, ok := errValue.(error); ok {
		err = e
	} else {
		err = fmt.Errorf("non-error (%T): %v", errValue, errValue)
	}

	var skipStack bool
	for _, f := range skipPanicStackDetectors {
		if f(err) {
			skipStack = true
			break
		}
	}
	var stack []byte
	if !skipStack {
		stack = debug.Stack()
	}

	return &Panic{err, stack}
}

func (p *Panic) Error() string {
	if p.stack == nil {
		return fmt.Sprintf("panic: %v", p.cause)
	} else {
		return fmt.Sprintf("panic: %v\n\n%s", p.cause, p.stack)
	}
}

func (p *Panic) Unwrap() error {
	return p.cause
}

var skipPanicStackDetectors = []func(error) bool{
	func(err error) bool {
		if e, ok := err.(SkipStackOnPanic); ok {
			return e.SkipStackOnPanic()
		}
		return false
	},
}

func SkipPanicStackIf(f func(error) bool) {
	skipPanicStackDetectors = append(skipPanicStackDetectors, f)
}

func SafeCall(f func() error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanic(p)
		}
	}()
	return f()
}
func SafeCall01[R1 any](f func() (R1, error)) (r1 R1, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanic(p)
		}
	}()
	return f()
}
func SafeCall11[A1, R1 any](f func(a1 A1) (R1, error), a1 A1) (r1 R1, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanic(p)
		}
	}()
	return f(a1)
}
func SafeCall21[A1, A2, R1 any](f func(a1 A1, a2 A2) (R1, error), a1 A1, a2 A2) (r1 R1, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = NewPanic(p)
		}
	}()
	return f(a1, a2)
}
