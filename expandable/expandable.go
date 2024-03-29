package expandable

import (
	"fmt"
	"reflect"
	"unsafe"
)

var nextBaseIndex = 0

type Schema struct {
	name  string
	types []any
}

func NewSchema(name string) *Schema {
	return &Schema{
		name: name,
	}
}

type Any[B any] interface {
	String() string
	Base() *Base[B]
	Type() reflect.Type
	PtrType() reflect.Type
	newThis() *B
	AnyFrom(base *B) any
	thisAnyToBase(drvd any) *B
}

type Impl struct {
	extraValues []any
}

type Expandable interface {
	ExpandableImpl() *Impl
}

func (impl *Impl) ExpandableImpl() *Impl {
	return impl
}

type Base[B any] struct {
	index  int
	schema *Schema
	name   string
	typ    reflect.Type
	ptrTyp reflect.Type
	new    func() *B
	full   Any[B]
}

func NewBase[B any](schema *Schema, name string) *Base[B] {
	ptrTyp := reflect.TypeOf((*B)(nil))
	t := &Base[B]{
		index:  nextBaseIndex,
		schema: schema,
		name:   name,
		ptrTyp: ptrTyp,
		typ:    ptrTyp.Elem(),
	}
	t.full = t
	nextBaseIndex++
	return t
}
func (t *Base[B]) WithNew(f func() *B) *Base[B] {
	t.new = f
	return t
}

func (t *Base[B]) FacetByPtrType(typ reflect.Type) Any[B] {
	if t.ptrTyp == typ {
		return t
	}
	if t.full.PtrType() == typ {
		return t.full
	}
	return nil
}

func (t *Base[B]) String() string {
	return t.schema.name + "." + t.name
}

func (t *Base[B]) Base() *Base[B] {
	return t
}
func (t *Base[B]) Type() reflect.Type {
	return t.typ
}
func (t *Base[B]) PtrType() reflect.Type {
	return t.ptrTyp
}

func (t *Base[B]) New() *B {
	return t.full.newThis()
}

func (t *Base[B]) AnyFull(base *B) any {
	return t.full.AnyFrom(base)
}

func (t *Base[B]) newThis() *B {
	if t.new != nil {
		return t.new()
	} else {
		return reflect.New(t.typ).Interface().(*B)
	}
}

func (t *Base[B]) AnyFrom(base *B) any {
	return base
}
func (t *Base[B]) thisAnyToBase(drvd any) *B {
	return drvd.(*B)
}

func (t *Base[B]) addDerived(d Any[B]) {
	if t.full != t {
		panic(fmt.Errorf("trying to derive another %s when %s has already been derived", d.String(), t.full.String()))
	}
	t.full = d
}

type Derived[T, B any] struct {
	schema *Schema
	typ    reflect.Type
	ptrTyp reflect.Type
	base   *Base[B]
	new    func() *T
}

func Derive[T, B any](schema *Schema, base *Base[B]) *Derived[T, B] {
	ptrTyp := reflect.TypeOf((*T)(nil))
	t := &Derived[T, B]{
		schema: schema,
		ptrTyp: ptrTyp,
		typ:    ptrTyp.Elem(),
		base:   base,
	}
	base.addDerived(t)
	return t
}
func (t *Derived[T, B]) WithNew(f func() *T) *Derived[T, B] {
	t.new = f
	return t
}

func (t *Derived[T, B]) String() string {
	return t.schema.name + "." + t.base.name
}

func (t *Derived[T, B]) Base() *Base[B] {
	return t.base
}
func (t *Derived[T, B]) Type() reflect.Type {
	return t.typ
}
func (t *Derived[T, B]) PtrType() reflect.Type {
	return t.ptrTyp
}

func (t *Derived[T, B]) newThis() *B {
	if t.new != nil {
		return t.ToBase(t.new())
	} else {
		return t.ToBase(reflect.New(t.typ).Interface().(*T))
	}
}
func (t *Derived[T, B]) From(base *B) *T {
	return (*T)(unsafe.Pointer(base))
}
func (t *Derived[T, B]) thisAnyToBase(drvd any) *B {
	return t.ToBase(drvd.(*T))
}
func (t *Derived[T, B]) ToBase(drvd *T) *B {
	return (*B)(unsafe.Pointer(drvd))
}
func (t *Derived[T, B]) AnyFrom(base *B) any {
	return t.From(base)
}
func (t *Derived[T, B]) Wrap(f func(*T)) func(*B) {
	return func(base *B) {
		f(t.From(base))
	}
}
func (t *Derived[T, B]) WrapAE(f func(*T) (any, error)) func(*B) (any, error) {
	return func(base *B) (any, error) {
		return f(t.From(base))
	}
}

func Wrap2[D1, D2, B1, B2 any](f func(*D1, *D2), d1 *Derived[D1, B1], d2 *Derived[D2, B2]) func(*B1, *B2) {
	return func(v1 *B1, v2 *B2) {
		f(d1.From(v1), d2.From(v2))
	}
}

func Wrap21[D1, T2, B1 any](f func(*D1, T2), d1 *Derived[D1, B1]) func(*B1, T2) {
	return func(v1 *B1, v2 T2) {
		f(d1.From(v1), v2)
	}
}

func Wrap21A[D1, T2, B1 any](f func(*D1, *T2) any, d1 *Derived[D1, B1]) func(*B1, *T2) any {
	return func(v1 *B1, v2 *T2) any {
		return f(d1.From(v1), v2)
	}
}

func Wrap2E[D1, D2, B1, B2 any](f func(*D1, *D2) error, d1 *Derived[D1, B1], d2 *Derived[D2, B2]) func(*B1, *B2) error {
	return func(v1 *B1, v2 *B2) error {
		return f(d1.From(v1), d2.From(v2))
	}
}

func Wrap21B[D1, T2, B1 any](f func(*D1, T2) bool, d1 *Derived[D1, B1]) func(*B1, T2) bool {
	return func(v1 *B1, v2 T2) bool {
		return f(d1.From(v1), v2)
	}
}

func Wrap1RE[T1, R1, B1 any](f func(*T1) (R1, error), d1 *Derived[T1, B1]) func(*B1) (R1, error) {
	return func(v1 *B1) (R1, error) {
		return f(d1.From(v1))
	}
}
