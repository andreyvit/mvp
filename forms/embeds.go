package forms

import "html/template"

type Template string

func (t Template) CurrentTemplate() string {
	return string(t)
}

type Header struct {
	Template
	TemplateStyle
	TagOpts
	Text string
}

func (Header) DefaultTemplate() string { return "embed-header" }

func (Header) Finalize(state *State) {}

type Image struct {
	Template
	TemplateStyle
	TagOpts
	Src    string
	Update func(el *Image)
}

func (*Image) DefaultTemplate() string { return "embed-image" }

func (*Image) Finalize(state *State) {}

func (el *Image) TriggerUpdate() {
	if el.Update != nil {
		el.Update(el)
	}
}

type Text struct {
	Template
	TemplateStyle
	TagOpts
	Text   string
	Update func(el *Text)
}

func (Text) DefaultTemplate() string { return "embed-text" }

func (Text) Finalize(state *State) {}

func (el *Text) TriggerUpdate() {
	if el.Update != nil {
		el.Update(el)
	}
}

type HTMLFragment struct {
	Template
	TemplateStyle
	TagOpts
	Text template.HTML
}

func (HTMLFragment) DefaultTemplate() string { return "embed-text" }

func (HTMLFragment) Finalize(state *State) {}

type Link struct {
	Template
	TemplateStyle
	TagOpts
	Link string
	Text string
}

func (Link) DefaultTemplate() string { return "embed-link-a" }

func (Link) Finalize(state *State) {}

type FreeButton struct {
	Template
	TemplateStyle
	Field
	TagOpts
	FullAction string
	Value      string
	Title      string
}

func (FreeButton) DefaultTemplate() string { return "control-button" }

func (c *FreeButton) Finalize(state *State) {}

func (c *FreeButton) EnumBindings(f func(AnyBinding)) {
}
