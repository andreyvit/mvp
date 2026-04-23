package forms

import (
	"fmt"
	"html/template"
	"strings"
)

type PresetSelectWithCustom[T comparable] struct {
	RenderableImpl[PresetSelectWithCustom[T]]
	Template
	TemplateStyle
	TagOpts
	*Binding[T]

	Options         []*Option[T]
	Custom          Child
	SelectFieldName string
	CustomFieldName string

	selectField Field
	customSlot  *namedChild
	customHTML  template.HTML
}

func (PresetSelectWithCustom[T]) DefaultTemplate() string {
	return "control-preset-select-with-custom"
}

func (c *PresetSelectWithCustom[T]) EnumFields(f func(name string, field *Field)) {
	f(c.selectFieldName(), &c.selectField)
}

func (c *PresetSelectWithCustom[T]) EnumChildren(f func(Child, ChildFlags)) {
	if c.Custom == nil {
		return
	}
	c.customSlot = &namedChild{
		Name:  c.customFieldName(),
		Child: c.Custom,
	}
	f(c.customSlot, ChildFlagSkipProcessing)
}

func (c *PresetSelectWithCustom[T]) Finalize(*State) {
	if len(c.Options) == 0 {
		panic("PresetSelectWithCustom requires at least one option")
	}
	if c.Custom == nil {
		panic("PresetSelectWithCustom requires a custom child")
	}
	if c.selectField.RawFormValue == "" {
		c.selectField.RawFormValue = c.HTMLValueForModelValue(c.Binding.Value())
	}
}

func (c *PresetSelectWithCustom[T]) Process(data *FormData) {
	if c.IsCustomSelected() {
		if data.Action == "update" && !childHasSubmittedFields(c.Custom, data) {
			return
		}
		walk(c.Custom, ChildFlagSkipProcessing, func(child Child) {
			if p, ok := child.(PreProcessor); ok {
				p.PreProcess(data)
			}
		}, func(child Child) {
			if p, ok := child.(Processor); ok {
				p.Process(data)
			}
		})
		return
	}
	opt := c.optionByHTMLValue(c.selectField.RawFormValue)
	if opt == nil {
		c.Binding.ErrSite.AddError(fmt.Errorf("invalid preset %q", c.selectField.RawFormValue))
		return
	}
	c.Binding.Set(opt.ModelValue)
}

func (c *PresetSelectWithCustom[T]) RenderInto(buf *strings.Builder, r *Renderer) {
	if c.IsCustomSelected() {
		c.customHTML = r.Render(c.Custom)
	} else {
		c.customHTML = ""
	}
	templ := c.CurrentTemplate()
	if templ == "" {
		templ = c.DefaultTemplate()
	}
	r.RenderTemplateInto(buf, templ, c)
}

func (c *PresetSelectWithCustom[T]) HTMLValueForModelValue(value T) string {
	for _, opt := range c.presetOptions() {
		if opt.ModelValue == value {
			return opt.HTMLValue
		}
	}
	return c.customOption().HTMLValue
}

func (c *PresetSelectWithCustom[T]) IsHTMLValueSelected(htmlValue string) bool {
	return htmlValue != "" && c.selectedHTMLValue() == htmlValue
}

func (c *PresetSelectWithCustom[T]) IsCustomSelected() bool {
	return c.selectedHTMLValue() == c.customOption().HTMLValue
}

func (c *PresetSelectWithCustom[T]) SelectFullName() string { return c.selectField.FullName }
func (c *PresetSelectWithCustom[T]) SelectID() string       { return c.selectField.ID }
func (c *PresetSelectWithCustom[T]) CustomHTML() template.HTML {
	return c.customHTML
}

func (c *PresetSelectWithCustom[T]) selectedHTMLValue() string {
	if c.selectField.RawFormValue != "" {
		return c.selectField.RawFormValue
	}
	return c.HTMLValueForModelValue(c.Binding.Value())
}

func (c *PresetSelectWithCustom[T]) optionByHTMLValue(value string) *Option[T] {
	for _, opt := range c.presetOptions() {
		if opt.HTMLValue == value {
			return opt
		}
	}
	return nil
}

func (c *PresetSelectWithCustom[T]) presetOptions() []*Option[T] {
	if len(c.Options) == 0 {
		return nil
	}
	return c.Options[:len(c.Options)-1]
}

func (c *PresetSelectWithCustom[T]) customOption() *Option[T] {
	if len(c.Options) == 0 {
		return nil
	}
	return c.Options[len(c.Options)-1]
}

func childHasSubmittedFields(child Child, data *FormData) bool {
	submitted := false
	walk(child, ChildFlagSkipProcessing, nil, func(child Child) {
		if submitted {
			return
		}
		p, ok := child.(Processor)
		if !ok {
			return
		}
		p.EnumFields(func(_ string, field *Field) {
			if field.RawFormValuePresent || len(data.Files[field.FullName]) > 0 {
				submitted = true
			}
		})
	})
	return submitted
}

func (c *PresetSelectWithCustom[T]) selectFieldName() string {
	if c.SelectFieldName != "" {
		return c.SelectFieldName
	}
	return "preset"
}

func (c *PresetSelectWithCustom[T]) customFieldName() string {
	if c.CustomFieldName != "" {
		return c.CustomFieldName
	}
	return "custom"
}

type namedChild struct {
	Name  string
	Child Child
}

func (c *namedChild) Finalize(state *State) {
	state.PushName(c.Name)
}

func (c *namedChild) EnumChildren(f func(Child, ChildFlags)) {
	f(c.Child, 0)
}
