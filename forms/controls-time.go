package forms

import (
	"fmt"
	"time"
)

type InputTime struct {
	RenderableImpl[InputTime]
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[time.Time]
	InputWellOpts
	IncludeTime bool
	Min         time.Time
	Max         time.Time
	Location    *time.Location
}

const (
	DateTimeLocalInputTimeFormat = "2006-01-02T15:04"
)

func (InputTime) DefaultTemplate() string { return "control-datetime-date" }

func (c *InputTime) InputType() string {
	if c.IncludeTime {
		return "datetime-local"
	} else {
		return "date"
	}
}

func (c *InputTime) ValueFormat() string {
	if c.IncludeTime {
		return DateTimeLocalInputTimeFormat
	} else {
		return time.DateOnly
	}
}

func (c *InputTime) PresentationFormat() string {
	if c.IncludeTime {
		return "2006-01-02 15:04"
	} else {
		return time.DateOnly
	}
}

func (c *InputTime) Finalize(state *State) {
	c.Validate(func(value time.Time) (time.Time, error) {
		if !c.Min.IsZero() && value.Before(c.Min) {
			return value, fmt.Errorf("must be on or after %v", c.Min.Format(c.PresentationFormat()))
		}
		if !c.Max.IsZero() && value.After(c.Max) {
			return value, fmt.Errorf("must be before %v", c.Max.Format(c.PresentationFormat()))
		}
		return value, nil
	})
}

func (c *InputTime) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, func(s string) (time.Time, error) {
		if s == "" {
			return time.Time{}, nil
		}
		return time.ParseInLocation(c.ValueFormat(), s, c.Location)
	})
}

func (c *InputTime) FormattedValue() string {
	return c.Binding.Value().In(c.Location).Format(c.ValueFormat())
}
