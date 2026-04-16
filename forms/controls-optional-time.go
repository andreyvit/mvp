package forms

import (
	"fmt"
	"time"
)

// OptionalInputTime renders a two-mode date input: a "Default" option that
// submits a zero time.Time, and a "Custom" option that reveals a date/time
// input and submits its parsed value. Suitable for fields with a meaningful
// inherited default where a zero value is stored to mean "inherit".
//
// Storage remains a single time.Time on the bound model. The mode flag
// lives entirely in the form's raw-value pipeline; no model-side state is
// introduced.
type OptionalInputTime struct {
	RenderableImpl[OptionalInputTime]
	Template
	TemplateStyle
	TagOpts
	*Binding[time.Time]

	Location    *time.Location
	IncludeTime bool
	Min         time.Time
	Max         time.Time

	DefaultValue time.Time
	DefaultLabel string
	CustomLabel  string

	modeField Field
	timeField Field
}

const (
	optionalTimeModeDefault = "default"
	optionalTimeModeCustom  = "custom"
)

func (OptionalInputTime) DefaultTemplate() string { return "control-optional-datetime" }

func (c *OptionalInputTime) EnumFields(f func(name string, field *Field)) {
	f("mode", &c.modeField)
	f("time", &c.timeField)
}

func (c *OptionalInputTime) valueFormat() string {
	if c.IncludeTime {
		return DateTimeLocalInputTimeFormat
	}
	return time.DateOnly
}

func (c *OptionalInputTime) presentationFormat() string {
	if c.IncludeTime {
		return "2006-01-02 15:04"
	}
	return time.DateOnly
}

func (c *OptionalInputTime) Finalize(state *State) {
	if c.DefaultLabel == "" {
		c.DefaultLabel = "Default"
	}
	if c.CustomLabel == "" {
		c.CustomLabel = "Custom"
	}
	if c.Location == nil {
		c.Location = time.UTC
	}

	if c.modeField.RawFormValue == "" {
		if c.Binding.Value().IsZero() {
			c.modeField.RawFormValue = optionalTimeModeDefault
		} else {
			c.modeField.RawFormValue = optionalTimeModeCustom
		}
	}

	c.Validate(func(value time.Time) (time.Time, error) {
		if !c.Min.IsZero() && value.Before(c.Min) {
			return value, fmt.Errorf("must be on or after %v", c.Min.Format(c.presentationFormat()))
		}
		if !c.Max.IsZero() && value.After(c.Max) {
			return value, fmt.Errorf("must be before %v", c.Max.Format(c.presentationFormat()))
		}
		return value, nil
	})
}

func (c *OptionalInputTime) Process(*FormData) {
	modeRaw := c.modeField.RawFormValue
	timeRaw := c.timeField.RawFormValue

	if modeRaw == "" {
		modeRaw = optionalTimeModeDefault
	}

	switch modeRaw {
	case optionalTimeModeDefault:
		c.Binding.Set(time.Time{})
	case optionalTimeModeCustom:
		c.Binding.SetString(timeRaw, func(s string) (time.Time, error) {
			if s == "" {
				return time.Time{}, fmt.Errorf("enter a date/time or choose %s", c.DefaultLabel)
			}
			return time.ParseInLocation(c.valueFormat(), s, c.Location)
		})
	default:
		c.Binding.ErrSite.AddError(fmt.Errorf("invalid mode %q", modeRaw))
	}
}

func (c *OptionalInputTime) DefaultOptionLabel() string {
	label := c.DefaultLabel
	if label == "" {
		label = "Default"
	}
	if c.DefaultValue.IsZero() {
		return label
	}
	return label + " \u2014 " + c.DefaultValue.In(c.Location).Format(c.presentationFormat())
}

func (c *OptionalInputTime) CustomOptionLabel() string {
	if c.CustomLabel == "" {
		return "Custom"
	}
	return c.CustomLabel
}

func (c *OptionalInputTime) InputType() string {
	if c.IncludeTime {
		return "datetime-local"
	}
	return "date"
}

func (c *OptionalInputTime) IsDefaultSelected() bool {
	if c.modeField.RawFormValue != "" {
		return c.modeField.RawFormValue == optionalTimeModeDefault
	}
	return c.Binding.Value().IsZero()
}

func (c *OptionalInputTime) IsCustomSelected() bool {
	return !c.IsDefaultSelected()
}

func (c *OptionalInputTime) ModeFullName() string { return c.modeField.FullName }
func (c *OptionalInputTime) ModeID() string       { return c.modeField.ID }
func (c *OptionalInputTime) TimeFullName() string { return c.timeField.FullName }
func (c *OptionalInputTime) TimeID() string       { return c.timeField.ID }

func (c *OptionalInputTime) TimeFormattedValue() string {
	if c.timeField.RawFormValuePresent {
		return c.timeField.RawFormValue
	}
	v := c.Binding.Value()
	if v.IsZero() {
		if c.DefaultValue.IsZero() {
			return ""
		}
		return c.DefaultValue.In(c.Location).Format(c.valueFormat())
	}
	return v.In(c.Location).Format(c.valueFormat())
}
