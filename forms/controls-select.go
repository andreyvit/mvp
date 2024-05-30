package forms

import (
	"fmt"
	"slices"
)

type Option[T comparable] struct {
	ModelValue T
	HTMLValue  string
	Label      string
}

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

func (Select[T]) IsMultiSelect() bool     { return false }
func (Select[T]) DefaultTemplate() string { return "control-select" }

func (c *Select[T]) IsHTMLValueSelected(htmlValue string) bool {
	return htmlValue != "" && c.RawFormValue == htmlValue
}

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

type RawSelect[T comparable] struct {
	RenderableImpl[RawSelect[T]]
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[T]

	Required           bool
	UpdateFormOnChange bool
	Options            []*Option[T]
	Stringify          func(T) string
	Parse              func(s string) (T, error)
}

func (RawSelect[T]) IsMultiSelect() bool     { return false }
func (RawSelect[T]) DefaultTemplate() string { return "control-select" }

func (c *RawSelect[T]) IsHTMLValueSelected(htmlValue string) bool {
	return htmlValue != "" && c.RawFormValue == htmlValue
}

func (c *RawSelect[T]) doStringify(item T) string {
	if c.Stringify == nil {
		return fmt.Sprint(item)
	} else {
		return c.Stringify(item)
	}
}

func (c *RawSelect[T]) Finalize(state *State) {
	optionsByHTMLValue := make(map[string]struct{}, len(c.Options))
	for _, opt := range c.Options {
		optionsByHTMLValue[opt.HTMLValue] = struct{}{}
	}

	if c.RawFormValue == "" {
		item := c.Binding.Value()
		var zero T
		if item != zero {
			s := c.doStringify(item)
			c.RawFormValue = s
			if _, ok := optionsByHTMLValue[s]; !ok {
				c.Options = append(c.Options, &Option[T]{
					ModelValue: item,
					HTMLValue:  s,
					Label:      s,
				})
			}
		}
	}
}

func (c *RawSelect[T]) Process(*FormData) {
	var item T
	if c.RawFormValue != "" {
		var err error
		item, err = c.Parse(c.RawFormValue)
		if err != nil {
			c.Binding.ErrSite.AddError(err)
			return
		}
	}
	c.Binding.Set(item)
}

type RawMultiSelect[T comparable] struct {
	RenderableImpl[RawMultiSelect[T]]
	Template
	TemplateStyle
	Field
	TagOpts
	*Binding[[]T]

	Required           bool
	UpdateFormOnChange bool
	Options            []*Option[T]
	Stringify          func(T) string
	Parse              func(s string) (T, error)
}

func (RawMultiSelect[T]) IsMultiSelect() bool     { return true }
func (RawMultiSelect[T]) DefaultTemplate() string { return "control-select" }

func (c *RawMultiSelect[T]) IsHTMLValueSelected(htmlValue string) bool {
	return slices.Contains(c.RawFormValues, htmlValue)
}

func (c *RawMultiSelect[T]) doStringify(item T) string {
	if c.Stringify == nil {
		return fmt.Sprint(item)
	} else {
		return c.Stringify(item)
	}
}

func (c *RawMultiSelect[T]) Finalize(state *State) {
	optionsByHTMLValue := make(map[string]struct{}, len(c.Options))
	for _, opt := range c.Options {
		optionsByHTMLValue[opt.HTMLValue] = struct{}{}
	}

	if c.RawFormValues == nil {
		for _, item := range c.Binding.Value() {
			s := c.doStringify(item)
			c.RawFormValues = append(c.RawFormValues, s)
			if _, ok := optionsByHTMLValue[s]; !ok {
				c.Options = append(c.Options, &Option[T]{
					ModelValue: item,
					HTMLValue:  s,
					Label:      s,
				})
			}
		}
	}
}

func (c *RawMultiSelect[T]) Process(*FormData) {
	var items []T
	for _, s := range c.RawFormValues {
		item, err := c.Parse(s)
		if err != nil {
			c.Binding.ErrSite.AddError(err)
			return
		}
		items = append(items, item)
	}
	c.Binding.Set(items)
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
