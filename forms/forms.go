package forms

import (
	"html/template"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unsafe"
)

type Child interface {
	Finalize(state *State)
}

type ChildFlags uint64

const (
	ChildFlagSkipProcessing ChildFlags = 1 << iota
)

type ChildAdder interface {
	AddChild(c ...Child)
}

type Children []Child

func (cc *Children) Add(c ...Child) {
	if c == nil {
		return
	}
	*cc = append(*cc, c...)
}
func (cc *Children) AddChild(c ...Child) {
	cc.Add(c...)
}

func Add[T Child](container ChildAdder, child T) T {
	container.AddChild(child)
	return child
}

func (cc Children) Finalize(state *State) {
}

func (cc Children) EnumChildren(f func(Child, ChildFlags)) {
	for _, c := range cc {
		f(c, 0)
	}
}

type RenderableImpl[T any] struct {
	IsHidden       bool
	updater        func(*T)
	visibilityFunc func() bool
}

func (impl *RenderableImpl[T]) owner() *T {
	return (*T)(unsafe.Pointer(impl))
}

func (impl *RenderableImpl[T]) BeforeRender() {
	if impl.visibilityFunc != nil {
		impl.IsHidden = !impl.visibilityFunc()
	}
	if impl.updater != nil {
		impl.updater(impl.owner())
	}
}

func (impl *RenderableImpl[T]) IsRenderableVisible() bool {
	return !impl.IsHidden
}

func (impl *RenderableImpl[T]) WithUpdater(f func(*T)) *T {
	impl.updater = f
	return impl.owner()
}
func (impl *RenderableImpl[T]) WithVisibility(f func() bool) *T {
	impl.visibilityFunc = f
	return impl.owner()
}

type Renderable interface {
	Child
	BeforeRender()
	IsRenderableVisible() bool
}

type Templated interface {
	Renderable
	TemplateStylePtr() *TemplateStyle
	CurrentTemplate() string
	DefaultTemplate() string
}

type CustomRenderable interface {
	Renderable
	RenderInto(buf *strings.Builder, r *Renderer)
}

type Container interface {
	Child
	EnumChildren(f func(Child, ChildFlags))
}

type Processor interface {
	Child
	EnumFields(f func(name string, field *Field))
	EnumBindings(f func(AnyBinding))
	Process(data *FormData)
}
type PreProcessor interface {
	Processor
	PreProcess(data *FormData)
}

type FormData struct {
	Action string
	Values url.Values
	Files  map[string][]*multipart.FileHeader
}

func walk(child Child, skipFlags ChildFlags, pre func(Child), post func(Child)) {
	if child == nil {
		return
	}
	if pre != nil {
		pre(child)
	}
	if cntr, ok := child.(Container); ok {
		cntr.EnumChildren(func(c Child, flags ChildFlags) {
			if (flags & skipFlags) != 0 {
				return
			}
			walk(c, skipFlags, pre, post)
		})
	}
	if post != nil {
		post(child)
	}
}

type Form struct {
	finalized bool
	fields    map[string]*Field
	Multipart bool
	URL       string
	Group

	ID     string
	Turbo  bool
	Action string
}

func (form *Form) TurboFrameID() string {
	if form.ID == "" || !form.Turbo {
		return ""
	} else {
		return form.ID + "-frame"
	}
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
	walk(&form.Group, ChildFlagSkipProcessing, func(c Child) {
		if p, ok := c.(PreProcessor); ok {
			p.PreProcess(data)
		}
	}, func(c Child) {
		if p, ok := c.(Processor); ok {
			p.Process(data)
		}
	})
	return !form.Invalid() && (form.Action == "submit")
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
	form.Group.wrapperForm = form
	state.finalizeTree(&form.Group, 0)
	form.fields = state.fields
	state.Fin()
}

type Group struct {
	RenderableImpl[Group]
	HTMLID     string
	Name       string
	Title      string
	WrapperTag TagOpts
	Styles     []*Style
	Template   string
	TemplateStyle
	Options  any
	Children Children
	Specials map[string]Child
	MultiErrorSite

	wrapperForm any

	InnerHTML   template.HTML
	SpecialHTML map[string]template.HTML
}

func (group *Group) AddChild(children ...Child) {
	group.Children.Add(children...)
}

func (group *Group) AddSpecial(key string, children ...Child) {
	if group.Specials == nil {
		group.Specials = make(map[string]Child)
	}
	group.Specials[key] = AsChild(children...)
}

func AsChild(children ...Child) Child {
	switch len(children) {
	case 0:
		return nil
	case 1:
		return children[0]
	default:
		return Children(children)
	}
}

func (group *Group) EnumChildren(f func(Child, ChildFlags)) {
	f(group.Children, 0)
}

func (group *Group) Finalize(state *State) {
	state.PushName(group.Name)
	state.PushErrorSite(&group.MultiErrorSite)
	state.PushStyles(group.Styles)
	if group.Template == "" {
		if group.wrapperForm != nil {
			group.Template = state.LookupTemplate("form")
		} else {
			group.Template = state.LookupTemplate("group")
		}
	}
}

func (group *Group) RenderInto(buf *strings.Builder, r *Renderer) {
	group.InnerHTML = r.Render(group.Children)
	for key, child := range group.Specials {
		if group.SpecialHTML == nil {
			group.SpecialHTML = make(map[string]template.HTML)
		}
		group.SpecialHTML[key] = r.Render(child)
	}
	if group.wrapperForm != nil {
		r.RenderWrapperTemplateInto(buf, group.Template, group.wrapperForm, group.InnerHTML)
	} else {
		r.RenderWrapperTemplateInto(buf, group.Template, group, group.InnerHTML)
	}
}

type Item struct {
	RenderableImpl[Item]
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

func (item *Item) EnumChildren(f func(Child, ChildFlags)) {
	f(item.Child, 0)
	f(item.Extra, 0)
}

func (item *Item) RenderInto(buf *strings.Builder, r *Renderer) {
	item.InnerHTML = r.Render(item.Child)
	if item.Extra != nil {
		item.ExtraHTML = r.Render(item.Extra)
	}
	r.RenderWrapperTemplateInto(buf, item.Template, item, item.InnerHTML)
}

type Wrapper struct {
	RenderableImpl[Wrapper]
	WrapperTag TagOpts
	Template   string
	TemplateStyle
	Child     Child
	InnerHTML template.HTML
}

func (wrapper *Wrapper) Finalize(state *State) {
}

func (wrapper *Wrapper) EnumChildren(f func(Child, ChildFlags)) {
	f(wrapper.Child, 0)
}

func (wrapper *Wrapper) RenderInto(buf *strings.Builder, r *Renderer) {
	wrapper.InnerHTML = r.Render(wrapper.Child)
	r.RenderWrapperTemplateInto(buf, wrapper.Template, wrapper, wrapper.InnerHTML)
}

type TagOpts struct {
	Class string
	Attrs map[string]any
}

func (to *TagOpts) SetAttr(name string, value any) {
	if to.Attrs == nil {
		to.Attrs = make(map[string]any)
	}
	to.Attrs[name] = value
}
