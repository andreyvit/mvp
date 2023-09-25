package forms

import (
	"errors"
	"strings"
)

var (
	ErrRequired = errors.New("Required")
	ErrTooShort = errors.New("Too short")
	ErrTooLong  = errors.New("Too long")
)

type ValidationFlags uint64

type StringRequirements struct {
	Required bool
	MinLen   int
	MaxLen   int
}

func ValidateString(s string, reqs StringRequirements) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		if reqs.Required {
			return s, ErrRequired
		}
	} else if reqs.MaxLen > 0 && len(s) > reqs.MaxLen {
		return s, ErrTooLong
	} else if reqs.MinLen > 0 && len(s) < reqs.MinLen {
		return s, ErrTooShort
	}
	return s, nil
}

func Validate[T, R any](errs ErrorSet, key string, value T, f func(value T, reqs R) (T, error), reqs R) T {
	clean, err := f(value, reqs)
	if err != nil {
		errs.Add(key, err)
	}
	return clean
}
