package forms

import (
	"fmt"
	"html/template"
	"io"
	"log"
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

func (r *Renderer) RenderInto(buf *strings.Builder, item any) {
	if c, ok := item.(Renderable); ok {
		c.RenderInto(buf, r)
	} else if t, ok := item.(Templated); ok {
		log.Printf("rendering templated %T %v", t, t)
		templ := t.CurrentTemplate()
		if templ == "" {
			templ = t.DefaultTemplate()
		}
		r.RenderTemplateInto(buf, templ, t)
	} else if subitems, ok := item.(Children); ok {
		for _, subitem := range subitems {
			r.RenderInto(buf, subitem)
		}
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
	r.RenderInto(&buf, item)
	return template.HTML(buf.String())
}
