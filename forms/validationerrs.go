package forms

import "errors"

type ErrorSet struct {
	byKey map[string][]error
}

func NewErrorSet() ErrorSet {
	return ErrorSet{
		byKey: make(map[string][]error),
	}
}

func (errs ErrorSet) Add(key string, err error) {
	if err == nil {
		return
	}
	errs.byKey[key] = append(errs.byKey[key], err)
}

func (errs ErrorSet) AddMsg(key string, msg string) {
	errs.Add(key, errors.New(msg))
}

func (errs ErrorSet) IsValid() bool {
	return len(errs.byKey) == 0
}
func (errs ErrorSet) IsInvalid() bool {
	return len(errs.byKey) > 0
}

func (errs ErrorSet) IsFieldValid(key string) bool {
	return len(errs.byKey[key]) == 0
}
func (errs ErrorSet) IsFieldInvalid(key string) bool {
	return len(errs.byKey[key]) > 0
}

func (errs ErrorSet) FieldErrors(key string) []error {
	return errs.byKey[key]
}

func (errs ErrorSet) FieldError(key string) error {
	all := errs.byKey[key]
	if len(all) > 0 {
		return all[0]
	} else {
		return nil
	}
}

func (errs ErrorSet) FieldMsg(key string) string {
	if err := errs.FieldError(key); err != nil {
		return err.Error()
	} else {
		return ""
	}
}
