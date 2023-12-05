package forms

type Select[T comparable] struct {
	RenderableImpl[Select[T]]
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[T]

	Required           bool
	UpdateFormOnChange bool
	Options            []*Option[T]
}

type Option[T comparable] struct {
	ModelValue T
	HTMLValue  string
	Label      string
}

func (Select[T]) DefaultTemplate() string { return "control-select" }

func (c *Select[T]) Finalize(state *State) {
	if c.RawFormValue == "" {
		opt := c.OptionByModelValue(c.Binding.Value())
		if opt != nil {
			c.RawFormValue = opt.HTMLValue
		}
	}
}

func (c *Select[T]) Process(*FormData) {
	opt := c.OptionByHTMLValue(c.RawFormValue)
	if opt != nil {
		c.Binding.Set(opt.ModelValue)
	}
}

func (c *Select[T]) OptionByModelValue(value T) *Option[T] {
	for _, opt := range c.Options {
		if opt.ModelValue == value {
			return opt
		}
	}
	return nil
}

func (c *Select[T]) OptionByHTMLValue(value string) *Option[T] {
	for _, opt := range c.Options {
		if opt.HTMLValue == value {
			return opt
		}
	}
	return nil
}

// func FormatOptions[T any, MV comparable](items []T, modelValue func(T) MV, htmlValue func(T) string, title func(T) string) []*Option[MV] {
// 	result := make([]*Option[MV], 0, len(items))
// 	for _, item := range items {
// 		result = append(result, &Option[MV]{
// 			ModelValue: modelValue(),
// 			HTMLValue:  "",
// 			Label:      "",
// 		})
// 	}
// }
