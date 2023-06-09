package forms

import (
	"fmt"
	"log"
	"strconv"

	"golang.org/x/exp/slices"
)

type AnyBinding interface {
	Init(errs ErrorSite)
}

type Binding[T any] struct {
	Value      T
	Setter     func(value T) error
	Validators []Validator[T]
	ErrSite    ErrorSite
	Child      AnyBinding
}

type Validator[T any] func(value T) (T, error)

func Var[T any](ptr *T) *Binding[T] {
	return &Binding[T]{
		Value: *ptr,
		Setter: func(value T) error {
			*ptr = value
			return nil
		},
	}
}

func Const[T any](value T) *Binding[T] {
	return &Binding[T]{
		Value: value,
		Setter: func(value T) error {
			return nil
		},
	}
}

func (b *Binding[T]) Validate(validators ...Validator[T]) *Binding[T] {
	b.Validators = append(b.Validators, validators...)
	return b
}

// EnumBindings in Binding simplifies implementation of Processors that embed a Binding.
func (b *Binding[T]) EnumBindings(f func(AnyBinding)) {
	f(b)
}

func (b *Binding[T]) Init(errs ErrorSite) {
	b.ErrSite = errs
	if b.Child != nil {
		b.Child.Init(errs)
	}
}

func (b *Binding[T]) Set(value T) {
	if b.ErrSite == nil {
		panic("no err site for binding")
	}
	b.Value = value

	for _, validator := range b.Validators {
		var err error
		value, err = validator(value)
		b.Value = value
		if err != nil {
			b.ErrSite.AddError(err)
			return
		}
	}

	err := b.Setter(value)
	if err != nil {
		b.ErrSite.AddError(err)
	} else {
		log.Printf("Binding value set: %T %v", value, value)
	}
}

func (b *Binding[T]) AsString(stringer func(T) string, convert func(string) (T, error)) *Binding[string] {
	if stringer == nil {
		stringer = func(t T) string {
			return fmt.Sprint(t)
		}
	}
	return &Binding[string]{
		Value: stringer(b.Value),
		Setter: func(str string) error {
			value, err := convert(str)
			if err != nil {
				return err
			}
			b.Set(value)
			return nil
		},
		ErrSite: b.ErrSite,
	}
}

func Convert[T, S any](b *Binding[S], stringer func(S) T, convert func(T) (S, error)) *Binding[T] {
	return &Binding[T]{
		Value: stringer(b.Value),
		Setter: func(source T) error {
			value, err := convert(source)
			if err != nil {
				return err
			}
			b.Set(value)
			return nil
		},
		ErrSite: b.ErrSite,
		Child:   b,
	}
}

func IntAsString(b *Binding[int], opts ...any) *Binding[string] {
	var specials []SpecialValue[int]
	for _, opt := range opts {
		switch opt := opt.(type) {
		case []SpecialValue[int]:
			specials = opt
		case SpecialValue[int]:
			specials = append(specials, opt)
		default:
			panic(fmt.Errorf("invalid option %T %v", opt, opt))
		}
	}
	return Convert(b, func(v int) string {
		for _, sp := range specials {
			if sp.ModelValue == v {
				return sp.PostbackValue
			}
		}
		return strconv.FormatInt(int64(v), 10)
	}, func(str string) (int, error) {
		for _, sp := range specials {
			if sp.PostbackValue == str {
				return sp.ModelValue, nil
			}
		}
		v, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return 0, fmt.Errorf("invalid number")
		}
		return int(v), nil
	})
}

func Int64AsString(b *Binding[int64], opts ...any) *Binding[string] {
	var specials []SpecialValue[int64]
	for _, opt := range opts {
		switch opt := opt.(type) {
		case []SpecialValue[int64]:
			specials = opt
		case SpecialValue[int64]:
			specials = append(specials, opt)
		default:
			panic(fmt.Errorf("invalid option %T %v", opt, opt))
		}
	}
	return Convert(b, func(v int64) string {
		for _, sp := range specials {
			if sp.ModelValue == v {
				return sp.PostbackValue
			}
		}
		return strconv.FormatInt(int64(v), 10)
	}, func(str string) (int64, error) {
		for _, sp := range specials {
			if sp.PostbackValue == str {
				return sp.ModelValue, nil
			}
		}
		v, err := strconv.ParseInt(str, 10, 0)
		if err != nil {
			return 0, fmt.Errorf("invalid number")
		}
		return int64(v), nil
	})
}

func BindSliceContainsEl[T comparable](b *Binding[[]T], el T) *Binding[bool] {
	return &Binding[bool]{
		Value: slices.Contains(b.Value, el),
		Setter: func(source bool) error {
			i := slices.Index(b.Value, el)
			if i < 0 {
				if source {
					items := slices.Clone(b.Value)
					items = append(items, el)
					b.Set(items)
				}
			} else {
				if !source {
					items := slices.Clone(b.Value)
					items = slices.Delete(items, i, i+1)
					b.Set(items)
				}
			}
			return nil
		},
		ErrSite: b.ErrSite,
		Child:   b,
	}
}

func BindSliceEmptyOrSingle[T comparable](b *Binding[[]T], emptyValue T) *Binding[T] {
	var value T
	if len(b.Value) == 0 {
		value = emptyValue
	} else {
		value = b.Value[0]
	}

	return &Binding[T]{
		Value: value,
		Setter: func(source T) error {
			if source == emptyValue {
				b.Set(nil)
			} else {
				b.Set([]T{source})
			}
			return nil
		},
		ErrSite: b.ErrSite,
		Child:   b,
	}
}

func BindMapKey[T any, K comparable](m map[K]T, key K) *Binding[T] {
	return &Binding[T]{
		Value: m[key],
		Setter: func(value T) error {
			m[key] = value
			return nil
		},
	}
}

func BindNot(b *Binding[bool]) *Binding[bool] {
	return &Binding[bool]{
		Value: !b.Value,
		Setter: func(source bool) error {
			b.Set(!source)
			return nil
		},
		ErrSite: b.ErrSite,
		Child:   b,
	}
}

func (b *Binding[T]) SetString(str string, convert func(string) (T, error)) {
	if b.ErrSite == nil {
		panic("no err site for binding")
	}
	value, err := convert(str)
	if err != nil {
		b.ErrSite.AddError(err)
	} else {
		b.Set(value)
	}
}

// parseBool is similar to strconv.ParseBool, but recognizes on/off values that browsers send for checkboxes, and treats empty value as false.
func parseBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "on", "ON", "On":
		return true, nil
	case "", "0", "f", "F", "false", "FALSE", "False", "off", "OFF", "Off":
		return false, nil
	}
	return false, fmt.Errorf("invalid bool")
}

func parseInt(str string) (int, error) {
	v, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	return int(v), nil
}

func parseInt64(str string) (int64, error) {
	v, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	return v, nil
}

func parseFloat64(str string) (float64, error) {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number")
	}
	return v, nil
}
