package mvp

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/andreyvit/mvp/flogger"
)

type ViewData struct {
	View         string
	Title        string
	Layout       string
	Data         any
	SemanticPath string

	ContentType string

	*SiteData
	Route  *Route
	App    *App
	RC     any
	baseRC *RC

	// Content is only populated in layouts and contains the rendered content of the page
	Content template.HTML
	// Head is extra HEAD content
	Head template.HTML
}

func (vd *ViewData) DefaultPathParams() map[string]string {
	defaults := make(map[string]string)
	return defaults
}

func (vd *ViewData) IsActive(path string) bool {
	return vd.SemanticPath == path || strings.HasPrefix(vd.SemanticPath, path+"/")
}

func (vd *ViewData) BaseRC() *RC {
	return vd.baseRC
}

func (vd *ViewData) LC() flogger.Context {
	return vd.baseRC
}

type SiteData struct {
	AppName string
}

type RenderData struct {
	Data any
	Args map[string]any
	*ViewData
}

func (d *RenderData) ArgsByPrefix(prefix string) map[string]any {
	prefix = prefix + "_"
	var result map[string]any
	for k, v := range d.Args {
		if ck, ok := strings.CutPrefix(k, prefix); ok {
			if result == nil {
				result = make(map[string]any)
			}
			result[ck] = v
		}
	}
	return result
}

func (d *RenderData) Bind(value any, args ...any) *RenderData {
	n := len(args)
	if n%2 != 0 {
		panic(fmt.Errorf("odd number of arguments %d: %v", n, args))
	}
	m := make(map[string]any, n/2)
	for i := 0; i < n; i += 2 {
		key, value := args[i], args[i+1]
		if keyStr, ok := key.(string); ok {
			keyStr = strings.ReplaceAll(keyStr, "-", "_")
			m[keyStr] = value
		} else {
			panic(fmt.Errorf("argument %d must be a string, got %T: %v", i, key, key))
		}
	}
	if len(m) == 0 {
		m["__dummy"] = true
	}
	return &RenderData{
		Data:     value,
		Args:     m,
		ViewData: d.ViewData,
	}
}

func (d *RenderData) Value(name string) (any, bool) {
	v, found := d.Args[name]
	return v, found
}

func (d *RenderData) String(name string) (string, bool) {
	v, found := d.Value(name)
	if found {
		return Stringify(v), true
	}
	return "", false
}

func (d *RenderData) PopString(name string) (string, bool) {
	v, found := d.String(name)
	if found {
		delete(d.Args, name)
	}
	return v, found
}

func (d *RenderData) HTMLSafeString(name string) (template.HTML, bool) {
	v, found := d.Value(name)
	if found {
		return HTMLify(v), true
	}
	return "", false
}

func (d *RenderData) PopHTMLSafeString(name string) (template.HTML, bool) {
	v, found := d.HTMLSafeString(name)
	if found {
		delete(d.Args, name)
	}
	return v, found
}

func (d *RenderData) MapSA(name string) (map[string]any, bool) {
	v, found := d.Value(name)
	if found {
		if v == nil {
			return nil, true
		}
		if result, ok := v.(map[string]any); ok {
			return result, true
		} else {
			panic(fmt.Errorf("value of arg %q is %T %v, wanted map[string]any", name, v, v))
		}
	}
	return nil, false
}

func (d *RenderData) PopMapSA(name string) (map[string]any, bool) {
	v, found := d.MapSA(name)
	if found {
		delete(d.Args, name)
	}
	return v, found
}

func isPassThruArg(k string) bool {
	return passThruArgs[k] || strings.HasPrefix(k, "data-")
}

var passThruArgs = map[string]bool{
	"id":     true,
	"target": true,
	"rel":    true,
}
