package forms

import (
	"fmt"
	"strings"
)

type Checkbox struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[bool]
}

func (Checkbox) DefaultTemplate() string { return "control-checkbox" }

func (c *Checkbox) Finalize(state *State) {
}

func (c *Checkbox) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, parseBool)
}

type InputWellOpts struct {
	LeftLabel     string
	LeftLabelTag  TagOpts
	RightLabel    string
	RightLabelTag TagOpts
}

type InputText struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[string]

	Required    bool
	MinLen      int
	MaxLen      int
	Placeholder string
}

func (InputText) DefaultTemplate() string { return "control-input-text" }

func (c *InputText) Finalize(state *State) {
	c.Binding.Validate(func(value string) (string, error) {
		value = strings.TrimSpace(value)
		if value == "" && c.Required {
			return value, ErrRequired
		}
		if len(value) < c.MinLen {
			return value, fmt.Errorf("must be %d+ chars", c.MinLen)
		}
		if c.MaxLen > 0 && len(value) > c.MaxLen {
			return value, fmt.Errorf("cannot be longer than %d chars", c.MaxLen)
		}
		return value, nil
	})
}

func (c *InputText) Process(*FormData) {
	c.Binding.Set(c.RawFormValue)
}

type InputInt struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[int]
	InputWellOpts
	Min    int
	Max    int
	Step   int
	HasMin bool
	HasMax bool
}

func (InputInt) DefaultTemplate() string { return "control-input-integer" }

func (c *InputInt) Finalize(state *State) {
	if c.Min != 0 {
		c.HasMin = true
	}
	if c.Max != 0 {
		c.HasMax = true
	}
	c.Validate(func(value int) (int, error) {
		if c.HasMin && value < c.Min {
			return value, fmt.Errorf("cannot be less than %v", c.Min)
		}
		if c.HasMax && value > c.Max {
			return value, fmt.Errorf("cannot be greater than %v", c.Max)
		}
		return value, nil
	})
}

func (c *InputInt) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, parseInt)
}

type InputInt64 struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[int64]
	InputWellOpts
	Min    int64
	Max    int64
	HasMin bool
	HasMax bool
	Step   int64
}

func (InputInt64) DefaultTemplate() string { return "control-input-integer" }

func (c *InputInt64) Finalize(state *State) {
	if c.Min != 0 {
		c.HasMin = true
	}
	if c.Max != 0 {
		c.HasMax = true
	}
	c.Validate(func(value int64) (int64, error) {
		if c.HasMin && value < c.Min {
			return value, fmt.Errorf("cannot be less than %v", c.Min)
		}
		if c.HasMax && value > c.Max {
			return value, fmt.Errorf("cannot be greater than %v", c.Max)
		}
		return value, nil
	})
}

func (c *InputInt64) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, parseInt64)
}

type InputFloat64 struct {
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[float64]
	InputWellOpts
	Min       float64
	Max       float64
	HasMin    bool
	HasMax    bool
	Precision int
}

func (InputFloat64) DefaultTemplate() string { return "control-input-float" }

func (c *InputFloat64) Finalize(state *State) {
	if c.Min != 0 {
		c.HasMin = true
	}
	if c.Max != 0 {
		c.HasMax = true
	}
	c.Validate(func(value float64) (float64, error) {
		if c.HasMin && value < c.Min {
			return value, fmt.Errorf("cannot be less than %v", c.Min)
		}
		if c.HasMax && value > c.Max {
			return value, fmt.Errorf("cannot be greater than %v", c.Max)
		}
		return value, nil
	})
}

func (c *InputFloat64) Process(*FormData) {
	c.Binding.SetString(c.RawFormValue, parseFloat64)
}

type Button struct {
	Template
	TemplateStyle
	Field
	TagOpts
	Value     string
	Activated bool
	Title     string
}

func (Button) DefaultTemplate() string { return "control-button" }

func (c *Button) Finalize(state *State) {}

func (c *Button) EnumBindings(f func(AnyBinding)) {
}

func (c *Button) Process(*FormData) {
	if c.Value != "" {
		c.Activated = (c.RawFormValue == c.Value)
	} else {
		c.Activated = c.RawFormValuePresent
	}
}
