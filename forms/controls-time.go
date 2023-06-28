package forms

import (
	"fmt"
	"time"
)

type InputTime struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[time.Time]
	InputWellOpts
	Min      time.Time
	Max      time.Time
	Location *time.Location
}

func (InputTime) DefaultTemplate() string { return "control-datetime-date" }

func (c *InputTime) Finalize(state *State) {
	c.Validate(func(value time.Time) (time.Time, error) {
		if !c.Min.IsZero() && value.Before(c.Min) {
			return value, fmt.Errorf("must be on or after %v", c.Min)
		}
		if !c.Max.IsZero() && value.After(c.Max) {
			return value, fmt.Errorf("must be before %v", c.Max)
		}
		return value, nil
	})
}

func (c *InputTime) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, func(s string) (time.Time, error) {
		return time.ParseInLocation(time.DateOnly, s, c.Location)
	})
}

func (c *InputTime) FormattedValue() string {
	return c.Binding.Value.In(c.Location).Format(time.DateOnly)
}
