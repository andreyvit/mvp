package forms

import "html/template"

type Template string

func (t Template) CurrentTemplate() string {
	return string(t)
}

type Header struct {
	RenderableImpl[Header]
	Template
	TemplateStyle
	TagOpts
	Text string
}

func (Header) DefaultTemplate() string { return "embed-header" }

func (Header) Finalize(state *State) {}

type Image struct {
	RenderableImpl[Image]
	Template
	TemplateStyle
	TagOpts
	Src string
}

func (*Image) DefaultTemplate() string { return "embed-image" }

func (*Image) Finalize(state *State) {}

type Text struct {
	RenderableImpl[Text]
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
	RenderableImpl[HTMLFragment]
	Template
	TemplateStyle
	TagOpts
	Text template.HTML
}

func (HTMLFragment) DefaultTemplate() string { return "embed-text" }

func (HTMLFragment) Finalize(state *State) {}

type Link struct {
	RenderableImpl[Link]
	Template
	TemplateStyle
	TagOpts
	Link string
	Text string
}

func (Link) DefaultTemplate() string { return "embed-link-a" }

func (Link) Finalize(state *State) {}

type FreeButton struct {
	RenderableImpl[FreeButton]
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
