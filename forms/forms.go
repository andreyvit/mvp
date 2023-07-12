package forms

import (
	"html/template"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Child interface {
	Finalize(state *State)
}

type Children []Child

func (cc *Children) Add(c ...Child) {
	if c == nil {
		return
	}
	*cc = append(*cc, c...)
}

func (cc Children) Finalize(state *State) {
}

func (cc Children) EnumChildren(f func(Child)) {
	for _, c := range cc {
		f(c)
	}
}

type Templated interface {
	Child
	TemplateStylePtr() *TemplateStyle
	CurrentTemplate() string
	DefaultTemplate() string
}

type Renderable interface {
	Child
	RenderInto(buf *strings.Builder, r *Renderer)
}

type Container interface {
	Child
	EnumChildren(f func(Child))
}

type Processor interface {
	Child
	EnumFields(f func(*Field))
	EnumBindings(f func(AnyBinding))
	Process(data *FormData)
}

type FormData struct {
	Action string
	Values url.Values
	Files  map[string][]*multipart.FileHeader
}

func walk(child Child, f func(Child)) {
	if child == nil {
		return
	}
	f(child)
	if cntr, ok := child.(Container); ok {
		cntr.EnumChildren(func(c Child) {
			walk(c, f)
		})
	}
}

type Form struct {
	finalized bool
	fields    map[string]*Field
	Multipart bool
	URL       string
	Group

	Action string
}

func (form *Form) Render(r *Renderer) template.HTML {
	form.finalize(nil)
	return r.Render(&form.Group)
}

func (form *Form) ProcessRequest(r *http.Request) bool {
	fd := FormData{
		Values: r.Form,
	}
	if r.MultipartForm != nil {
		fd.Files = r.MultipartForm.File
	}
	return form.Process(&fd)
}

func (form *Form) Process(data *FormData) bool {
	if data.Action == "" {
		data.Action = data.Values.Get("action")
		if data.Action == "" {
			data.Action = "submit"
		}
	}
	form.Action = data.Action

	form.finalize(data)
	for name, field := range form.fields {
		field.RawFormValues = data.Values[name]
		field.RawFormValue = ""
		field.RawFormValuePresent = false
		for _, v := range field.RawFormValues {
			field.RawFormValue = v
			field.RawFormValuePresent = true
		}
	}
	walk(&form.Group, func(c Child) {
		if p, ok := c.(Processor); ok {
			p.Process(data)
		}
	})
	return !form.Invalid()
}

func (form *Form) finalize(data *FormData) {
	if form.finalized {
		return
	}
	form.finalized = true

	state := State{
		Data:          data,
		path:          make([]string, 0, 10),
		fields:        make(map[string]*Field, 100),
		classes:       make([]map[string]string, 0, 10),
		classesCopied: make([]bool, 0, 50),
	}
	state.classes = append(state.classes, nil)
	form.Group.isRootForm = true
	form.Group.Form = form
	state.finalizeTree(&form.Group)
	form.fields = state.fields
	state.Fin()
}

type Group struct {
	Name       string
	Title      string
	WrapperTag TagOpts
	Styles     []*Style
	Template   string
	TemplateStyle
	Options  any
	Children Children
	MultiErrorSite

	isRootForm bool
	Form       *Form

	InnerHTML template.HTML
}

func (group *Group) EnumChildren(f func(Child)) {
	f(group.Children)
}

func (group *Group) Finalize(state *State) {
	state.PushName(group.Name)
	state.PushErrorSite(&group.MultiErrorSite)
	state.PushStyles(group.Styles)
	if group.Template == "" {
		if group.isRootForm {
			group.Template = state.LookupTemplate("form")
		} else {
			group.Template = state.LookupTemplate("group")
		}
	}
}

func (group *Group) RenderInto(buf *strings.Builder, r *Renderer) {
	group.InnerHTML = r.Render(group.Children)
	r.RenderWrapperTemplateInto(buf, group.Template, group, group.InnerHTML)
}

type Item struct {
	Name string
	Identity

	Label    string
	LabelTag TagOpts
	Desc     string
	DescTag  TagOpts
	ItemTag  TagOpts
	Styles   []*Style
	Template string
	TemplateStyle
	SingleErrorSite

	Child     Child
	Extra     Child
	InnerHTML template.HTML
	ExtraHTML template.HTML
}

func (item *Item) Finalize(state *State) {
	state.PushName(item.Name)
	state.PushErrorSite(&item.SingleErrorSite)
	state.PushStyles(item.Styles)
	if item.Template == "" {
		item.Template = state.LookupTemplate("item")
	}
	state.AssignIdentity(&item.Identity)
}

func (item *Item) EnumChildren(f func(Child)) {
	f(item.Child)
	f(item.Extra)
}

func (item *Item) RenderInto(buf *strings.Builder, r *Renderer) {
	item.InnerHTML = r.Render(item.Child)
	if item.Extra != nil {
		item.ExtraHTML = r.Render(item.Extra)
	}
	r.RenderWrapperTemplateInto(buf, item.Template, item, item.InnerHTML)
}

type Wrapper struct {
	WrapperTag TagOpts
	Template   string
	TemplateStyle
	Child     Child
	InnerHTML template.HTML
}

func (wrapper *Wrapper) Finalize(state *State) {
}

func (wrapper *Wrapper) EnumChildren(f func(Child)) {
	f(wrapper.Child)
}

func (wrapper *Wrapper) RenderInto(buf *strings.Builder, r *Renderer) {
	wrapper.InnerHTML = r.Render(wrapper.Child)
	r.RenderWrapperTemplateInto(buf, wrapper.Template, wrapper, wrapper.InnerHTML)
}

type TagOpts struct {
	Class string
	Attrs map[string]any
}
