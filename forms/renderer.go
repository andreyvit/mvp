package forms

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

type Renderer struct {
	Exec func(w io.Writer, templateName string, data any) error
}

func (r *Renderer) RenderTemplateInto(buf *strings.Builder, templ string, data any) {
	err := r.Exec(buf, templ, data)
	if err != nil {
		panic(fmt.Errorf("%s: %v", templ, err))
	}
}

func (r *Renderer) RenderWrapperTemplateInto(buf *strings.Builder, templ string, data any, raw template.HTML) {
	if templ != "" && templ != "none" {
		r.RenderTemplateInto(buf, templ, data)
	} else {
		buf.WriteString(string(raw))
	}
}

func (r *Renderer) RenderItemInto(buf *strings.Builder, item any) {
	if upd, ok := item.(Renderable); ok {
		upd.BeforeRender()
		if !upd.IsRenderableVisible() {
			return
		}
	}
	if c, ok := item.(CustomRenderable); ok {
		c.RenderInto(buf, r)
	} else if t, ok := item.(Templated); ok {
		// log.Printf("rendering templated %T %v", t, t)
		templ := t.CurrentTemplate()
		if templ == "" {
			templ = t.DefaultTemplate()
		}
		r.RenderTemplateInto(buf, templ, t)
	} else if container, ok := item.(Container); ok {
		container.EnumChildren(func(c Child, cf ChildFlags) {
			r.RenderItemInto(buf, c)
		})
	} else if s, ok := item.(string); ok {
		buf.WriteString(template.HTMLEscapeString(s))
	} else if h, ok := item.(template.HTML); ok {
		buf.WriteString(string(h))
	} else if item == nil {
	} else {
		panic(fmt.Errorf("don't know how to render %T %v", item, item))
	}
}

func (r *Renderer) Render(item any) template.HTML {
	var buf strings.Builder
	r.RenderItemInto(&buf, item)
	return template.HTML(buf.String())
}
