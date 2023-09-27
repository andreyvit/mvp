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
	componentWithFuncTempl
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

func RenderPartial(rc *RC, vd *ViewData) template.HTML {
	var buf strings.Builder
	RenderPartialTo(&buf, rc, vd)
	return template.HTML(buf.String())
}

func RenderPartialTo(wr io.Writer, rc *RC, vd *ViewData) {
	rc.app.fillViewData(vd, rc)

	err := rc.app.freshTemplates(rc).ExecuteTemplate(wr, vd.View, &RenderData{Data: vd.Data, ViewData: vd})
	if err != nil {
		panic(PartialRenderingError(vd.View, err))
	}
}

func (app *App) Render(lc flogger.Context, data *ViewData) ([]byte, error) {
	if data.Layout == "" {
		data.Layout = "default"
	}

	t := app.freshTemplates(lc)

	rdata := &RenderData{Data: data.Data, ViewData: data}

	if data.View != "none" {
		var buf strings.Builder
		err := t.ExecuteTemplate(&buf, data.View, rdata)
		if err != nil {
			return nil, err
		}
		data.Content = template.HTML(buf.String())
	}

	if data.Layout == "none" {
		return []byte(data.Content), nil
	}

	var buf2 bytes.Buffer
	err := t.ExecuteTemplate(&buf2, "layouts/"+data.Layout, rdata)
	if err != nil {
		return nil, err
	}
	return buf2.Bytes(), nil
}

func (app *App) freshTemplates(lc flogger.Context) *template.Template {
	if app.Settings.ServeAssetsFromDisk {
		// flogger.Log(lc, "reloading templates from disk")
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
		app.staticFS = must(fs.Sub(ge.EmbeddedStaticFS, ge.StaticSubdir))
		app.viewsFS = must(fs.Sub(ge.EmbeddedViewsFS, ge.ViewsSubdir))
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

	err := fs.WalkDir(app.viewsFS, ".", func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", fullPath, err)
		}
		relPath := fullPath //strings.TrimPrefix(fullPath, app.Configuration.ViewsSubdir+"/")
		// log.Printf("loadTemplates sees: %v", relPath)
		if !strings.HasSuffix(relPath, templateSuffix) {
			return nil
		}
		name := strings.TrimSuffix(relPath, templateSuffix)
		baseName := strings.TrimSuffix(d.Name(), templateSuffix)
		code := string(must(fs.ReadFile(app.viewsFS, fullPath)))

		var kind templKind
		if strings.HasPrefix(relPath, "layouts/") {
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
			path: fullPath,
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
	compTempls := make(map[string]*templDef)
	for _, tmpl := range templs {
		if tmpl.kind == componentTempl {
			comps[tmpl.name] = minicomponents.ScanTemplate(tmpl.code)
			compTempls[tmpl.name] = tmpl
		}
	}
	for k := range funcs {
		if strings.HasPrefix(k, "c_") {
			name := strings.ReplaceAll(k, "_", "-")
			if comps[name] != nil {
				panic(fmt.Errorf("%s is defined as both a function and a component; use prep_c_ prefix for a code-behind func", name))
			} else {
				comps[name] = &minicomponents.ComponentDef{
					RenderMethod: minicomponents.RenderMethodFunc,
					FuncName:     k,
				}
			}
		} else if strings.HasPrefix(k, "prep_c_") {
			name := strings.ReplaceAll(strings.TrimPrefix(k, "prep_"), "_", "-")
			c := comps[name]
			if c == nil {
				panic(fmt.Errorf("no template to match code-behind func %s, wanted %s", k, name))
			}
			c.RenderMethod = minicomponents.RenderMethodFuncThenTemplate
			c.FuncName = k
			compTempls[name].kind = componentWithFuncTempl
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
		} else if tmpl.kind == pageTempl || tmpl.kind == partialTempl || tmpl.kind == componentWithFuncTempl {
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
