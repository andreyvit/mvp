package mvp

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/andreyvit/minicomponents"
	"github.com/andreyvit/mvp/flogger"
)

type templKind int

const (
	partialTempl = templKind(iota)
	componentTempl
	layoutTempl
	pageTempl
)

func (app *App) EvalTemplate(templateName string, data any) (template.HTML, error) {
	var buf strings.Builder
	err := app.ExecTemplate(&buf, templateName, data)
	if err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func (app *App) ExecTemplate(w io.Writer, templateName string, data any) error {
	t := app.templates
	if app.Settings.ServeAssetsFromDisk {
		t = app.templatesDev.Load().(*template.Template)
	}
	return t.ExecuteTemplate(w, templateName, data)
}

func (app *App) RenderPartial(rc *RC, view string, data any) template.HTML {
	vd := &ViewData{
		View: view,
		Data: data,
	}
	app.fillViewData(vd, rc)

	var buf strings.Builder
	err := app.freshTemplates(rc).ExecuteTemplate(&buf, view, &RenderData{Data: data, ViewData: vd})
	if err != nil {
		flogger.Log(rc, "FATAL: partial rendering failed: %v: %v", view, err)
		panic(fmt.Sprintf("partial rendering failed: %v: %v", view, err))
	}
	return template.HTML(buf.String())
}

func (app *App) Render(lc flogger.Context, data *ViewData) ([]byte, error) {
	if data.Layout == "" {
		data.Layout = "default"
	}

	t := app.freshTemplates(lc)

	rdata := &RenderData{Data: data.Data, ViewData: data}

	var buf strings.Builder
	err := t.ExecuteTemplate(&buf, data.View, rdata)
	if err != nil {
		return nil, err
	}
	data.Content = template.HTML(buf.String())

	if data.Layout == "none" {
		return []byte(data.Content), nil
	}

	var buf2 bytes.Buffer
	err = t.ExecuteTemplate(&buf2, "layouts/"+data.Layout, rdata)
	if err != nil {
		return nil, err
	}
	return buf2.Bytes(), nil
}

func (app *App) freshTemplates(lc flogger.Context) *template.Template {
	if app.Settings.ServeAssetsFromDisk {
		flogger.Log(lc, "reloading templates from disk")
		t, err := app.loadTemplates()
		if err != nil {
			panic(fmt.Errorf("reloading templates: %v", err))
		}
		app.templatesDev.Store(t)
		return t
	} else {
		return app.templates
	}
}

type templDef struct {
	name string
	path string
	code string
	kind templKind
	tmpl *template.Template
}

func initViews(app *App, opt *AppOptions) {
	ge := app.Settings.Configuration
	if app.Settings.ServeAssetsFromDisk {
		app.staticFS = os.DirFS(filepath.Join(ge.LocalDevAppRoot, ge.StaticSubdir))
		app.viewsFS = os.DirFS(filepath.Join(ge.LocalDevAppRoot, ge.ViewsSubdir))
	} else {
		app.staticFS = ge.EmbeddedStaticFS
		app.viewsFS = ge.EmbeddedViewsFS
	}

	var err error
	app.templates, err = app.loadTemplates()
	if err != nil {
		log.Fatalf("failed to load templates: %v", err)
	}
	if app.Settings.ServeAssetsFromDisk {
		app.templatesDev.Store(app.templates)
	}
}

func (app *App) loadTemplates() (*template.Template, error) {
	const templateSuffix = ".html"

	funcs := make(template.FuncMap, 100)
	app.registerBuiltinViewHelpers(funcs)
	for _, f := range app.Hooks.helpers {
		f(funcs)
	}

	root := template.New("")
	root.Funcs(funcs)

	var templs []*templDef

	err := fs.WalkDir(app.viewsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !strings.HasSuffix(path, templateSuffix) {
			return nil
		}
		name := strings.TrimSuffix(path, templateSuffix)
		baseName := strings.TrimSuffix(d.Name(), templateSuffix)
		code := string(must(fs.ReadFile(app.viewsFS, path)))

		var kind templKind
		if strings.HasPrefix(path, "layouts/") {
			kind = layoutTempl
		} else if strings.HasPrefix(baseName, "c-") {
			kind = componentTempl
			name = baseName
		} else if strings.Contains(baseName, "__") {
			kind = partialTempl
		} else {
			kind = pageTempl
		}

		templs = append(templs, &templDef{
			name: name,
			path: path,
			code: code,
			kind: kind,
			tmpl: root.New(name),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	comps := make(map[string]*minicomponents.ComponentDef)
	for _, tmpl := range templs {
		if tmpl.kind == componentTempl {
			comps[tmpl.name] = minicomponents.ScanTemplate(tmpl.code)
		}
	}
	for k := range funcs {
		if strings.HasPrefix(k, "c_") {
			name := strings.ReplaceAll(k, "_", "-")
			comps[name] = &minicomponents.ComponentDef{
				RenderMethod: minicomponents.RenderMethodFunc,
				ImplName:     k,
			}
		}
	}

	for _, tmpl := range templs {
		code := tmpl.code
		code, _ = minicomponents.Rewrite(code, tmpl.name, comps)

		var defines string
		if i := strings.Index(code, "{{define"); i >= 0 {
			code, defines = code[:i], code[i:]
		}

		if tmpl.kind == componentTempl {
			code = "{{with .Args}}" + code + "{{end}}" + defines
		} else if tmpl.kind == pageTempl || tmpl.kind == partialTempl {
			code = "{{with .Data}}" + code + "{{end}}" + defines
		} else {
			code = code + defines
		}

		_, err = tmpl.tmpl.Parse(code)
		if err != nil || strings.Contains(code, "{{error") {
			log.Printf("Code of %s:\n==========\n%s\n==========", tmpl.name, code)
		}
		if err != nil {
			return nil, fmt.Errorf("error parsing %v: %w", tmpl.path, err)
		}
	}

	return root, nil
}
