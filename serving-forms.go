package mvp

import (
	"html/template"
	"io"

	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/forms"
)

func (app *App) RenderForm(rc *RC, form *forms.Form) template.HTML {
	r := &forms.Renderer{
		Exec: func(w io.Writer, templateName string, data any) error {
			if templateName == "" {
				panic("empty template name")
			}
			templateName = "forms/" + templateName
			rd := &RenderData{
				Data:     data,
				ViewData: &ViewData{},
			}
			app.fillViewData(*&rd.ViewData, rc)
			flogger.Log(rc, "executing form template %s", templateName)
			return app.ExecTemplate(w, templateName, rd)
		},
	}
	return form.Render(r)
}

func (rc *RC) HandleForm(form *forms.Form) bool {
	return form.ProcessRequest(rc.Request.Request)
}
