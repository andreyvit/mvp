package forms

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
	Src string
}

func (Image) DefaultTemplate() string { return "embed-image" }

func (Image) Finalize(state *State) {}

type Text struct {
	Template
	TemplateStyle
	TagOpts
	Text string
}

func (Text) DefaultTemplate() string { return "embed-text" }

func (Text) Finalize(state *State) {}

type Link struct {
	Template
	TemplateStyle
	TagOpts
	Link string
}

func (Link) DefaultTemplate() string { return "embed-link-a" }

func (Link) Finalize(state *State) {}

type FreeButton struct {
	Template
	TemplateStyle
	Field
	TagOpts
	Value string
	Title string
}

func (FreeButton) DefaultTemplate() string { return "control-button" }

func (c *FreeButton) Finalize(state *State) {}

func (c *FreeButton) EnumBindings(f func(AnyBinding)) {
}
