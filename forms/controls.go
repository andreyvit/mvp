package forms

import (
	"fmt"
	"strings"
)

type Checkbox struct {
	RenderableImpl[Checkbox]
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
	RenderableImpl[InputText]
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
	RenderableImpl[InputInt]
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
	RenderableImpl[InputInt64]
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
	RenderableImpl[InputFloat64]
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

type SpecialValue[T comparable] struct {
	ModelValue    T
	PostbackValue string
}

type Button struct {
	RenderableImpl[Button]
	Template
	TemplateStyle
	Identity
	TagOpts
	Action     string
	FullAction string
	Activated  bool
	Title      string
	Handler    func()

	ConfirmationMessage string
}

func (Button) DefaultTemplate() string { return "control-button" }

func (c *Button) Finalize(state *State) {
	if c.Action == "" {
		c.Action = "submit"
	}
	state.PushName(c.Action)
	state.AssignIdentity(&c.Identity)
	if c.FullAction == "" {
		if c.Action == "submit" {
			c.FullAction = "submit"
		} else {
			c.FullAction = c.FullName
		}
	}
}

func (c *Button) EnumFields(f func(name string, field *Field)) {
}

func (c *Button) EnumBindings(f func(AnyBinding)) {
}

func (c *Button) Process(fd *FormData) {
	// log.Printf("Button.Process c.FullAction=%q fd.Action=%q", c.FullAction, fd.Action)
	c.Activated = (fd.Action == c.FullAction)
	if c.Activated && c.Handler != nil {
		c.Handler()
	}
}

type Hidden struct {
	RenderableImpl[Hidden]
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[string]
}

func (Hidden) DefaultTemplate() string { return "control-hidden" }

func (c *Hidden) Finalize(state *State) {}

func (c *Hidden) Process(*FormData) {
	c.Binding.Set(c.RawFormValue)
}
