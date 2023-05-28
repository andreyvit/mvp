package mvp

import (
	"context"
	"fmt"
	"sync"
)

type ExpandableType struct {
	StoredValues []AnyVal
}

type ExpandableObject interface {
	// ValueAt(idx int) any
	ValueSet() ValueSet
}

type ValueSet []any

func newValueSet() []any {
	definedValuesSafety.once.Do(finalizeDefinedValues)
	values := make([]any, len(definedValues))
	for i, val := range definedValues {
		values[i] = val.initialValue()
	}
	return values
}

var (
	definedValues       []AnyVal
	definedValuesSafety struct {
		final bool
		mut   sync.Mutex
		once  sync.Once
	}

	GoCtx = Value[context.Context]("goctx")
)

type undefined struct{}

func (_ undefined) String() string {
	return "<undefined>"
}

func IsUndefined(v any) bool {
	return v == undefined{}
}

type AnyVal interface {
	Name() string
	initialValue() any
}

type Val[T any] struct {
	index       int
	name        string
	initializer func() T
}

func (val *Val[T]) Name() string {
	return val.name
}

func (val *Val[T]) initialValue() any {
	if val.initializer != nil {
		return val.initializer()
	}
	return undefined{}
}

func (val *Val[T]) Get(o ExpandableObject) T {
	switch v := o.ValueSet()[val.index].(type) {
	case T:
		return v
	case undefined:
		panic(fmt.Errorf("%s has no defined value in this context", val.name))
	default:
		panic("unreachable: unexpected type")
	}
}

func (val *Val[T]) Set(o ExpandableObject, v T) {
	o.ValueSet()[val.index] = v
}

func Value[T any](name string, opts ...func(val *Val[T])) *Val[T] {
	definedValuesSafety.mut.Lock()
	defer definedValuesSafety.mut.Unlock()
	if definedValuesSafety.final {
		panic("cannot define new values after Context has been created")
	}
	val := &Val[T]{
		index: len(definedValues),
		name:  name,
	}
	for _, opt := range opts {
		opt(val)
	}
	definedValues = append(definedValues, val)
	return val
}

// ValueDefault is an option to pass to Value.
func ValueDefault[T any](v T) func(val *Val[T]) {
	return ValueInitializer(func() T { return v })
}

// ValueZeroInit is an option to pass to Value.
func ValueZeroInit[T any](val *Val[T]) {
	val.initializer = func() T {
		var zero T
		return zero
	}
}

// ValueInitializer is an option to pass to Value.
func ValueInitializer[T any](f func() T) func(val *Val[T]) {
	return func(val *Val[T]) { val.initializer = f }
}

// ValueDep is an option to pass to Value. It signals that this value should be
// initialized after those other values.
func ValueDep[T any](deps ...AnyVal) func(val *Val[T]) {
	return func(val *Val[T]) {
		// No code necessary: just because we're referencing other values,
		// they will defined before the current value and thus will have lower indices.
	}
}

func finalizeDefinedValues() {
	definedValuesSafety.mut.Lock()
	defer definedValuesSafety.mut.Unlock()
	definedValuesSafety.final = true
}
