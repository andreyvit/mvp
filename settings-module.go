package mvp

import (
	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/mvpjobs"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

type Module struct {
	Name string

	SetupHooks  func(app *App)
	LoadSecrets func(*Settings, Secrets)

	DBSchema  *edb.Schema
	JobSchema *mvpjobs.Schema
	Types     map[mvpm.Type][]string
}
