package mvp

import (
	"html/template"

	"github.com/andreyvit/edb"
)

type Hooks struct {
	initApp      []func(app *App)
	closeApp     []func(app *App)
	initRC       []func(app *App, rc *RC)
	closeRC      []func(app *App, rc *RC)
	initDB       []func(app *App, rc *RC)
	migrate      []func(app *App, b *MigrationBuilder)
	makeRowKey   []func(app *App, tbl *edb.Table) any
	resetAuth    []func(app *App, rc *RC)
	postAuth     []func(app *App, rc *RC) error
	helpers      []func(m template.FuncMap)
	middleware   []func(r Router)
	domainRoutes []func(app *App, b *DomainRouter)
	siteRoutes   map[*Site][]func(b *RouteBuilder)
	urlGenOption []func(app *App, g *URLGen, option string) bool
	urlGen       []func(app *App, g *URLGen)
	jwtTokenKey  []func(rc *RC, c *TokenDecoding) error
}

func (h *Hooks) InitApp(f func(app *App)) {
	h.initApp = append(h.initApp, f)
}

func (h *Hooks) CloseApp(f func(app *App)) {
	h.closeApp = append(h.closeApp, f)
}

func (h *Hooks) InitRC(f func(app *App, rc *RC)) {
	h.initRC = append(h.initRC, f)
}

func (h *Hooks) CloseRC(f func(app *App, rc *RC)) {
	h.closeRC = append(h.closeRC, f)
}

func (h *Hooks) InitDB(f func(app *App, rc *RC)) {
	h.initDB = append(h.initDB, f)
}

func (h *Hooks) Migrate(f func(app *App, b *MigrationBuilder)) {
	h.migrate = append(h.migrate, f)
}

func (h *Hooks) MakeRowKey(f func(app *App, tbl *edb.Table) any) {
	h.makeRowKey = append(h.makeRowKey, f)
}

func (h *Hooks) ResetAuth(f func(app *App, rc *RC)) {
	h.resetAuth = append(h.resetAuth, f)
}

func (h *Hooks) PostAuth(f func(app *App, rc *RC) error) {
	h.postAuth = append(h.postAuth, f)
}

func (h *Hooks) Helpers(f func(m template.FuncMap)) {
	h.helpers = append(h.helpers, f)
}

func (h *Hooks) Middleware(f func(r Router)) {
	h.middleware = append(h.middleware, f)
}

func (h *Hooks) DomainRoutes(f func(app *App, b *DomainRouter)) {
	h.domainRoutes = append(h.domainRoutes, f)
}

func (h *Hooks) SiteRoutes(site *Site, f func(b *RouteBuilder)) {
	if h.siteRoutes == nil {
		h.siteRoutes = make(map[*Site][]func(b *RouteBuilder))
	}
	h.siteRoutes[site] = append(h.siteRoutes[site], f)
}

func (h *Hooks) URLGenOption(f func(app *App, g *URLGen, option string) bool) {
	h.urlGenOption = append(h.urlGenOption, f)
}

func (h *Hooks) URLGen(f func(app *App, g *URLGen)) {
	h.urlGen = append(h.urlGen, f)
}

func (h *Hooks) JWTTokenKey(f func(rc *RC, c *TokenDecoding) error) {
	h.jwtTokenKey = append(h.jwtTokenKey, f)
}

func runHooksFwd1[T1 any](hooks []func(a1 T1), a1 T1) {
	for _, f := range hooks {
		f(a1)
	}
}

func runHooksRev1[T1 any](hooks []func(a1 T1), a1 T1) {
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i](a1)
	}
}

func runHooksFwd2[T1, T2 any](hooks []func(a1 T1, a2 T2), a1 T1, a2 T2) {
	for _, f := range hooks {
		f(a1, a2)
	}
}

func runHooksRev2[T1, T2 any](hooks []func(a1 T1, a2 T2), a1 T1, a2 T2) {
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i](a1, a2)
	}
}

func runHooksFwd2E[T1, T2 any](hooks []func(a1 T1, a2 T2) error, a1 T1, a2 T2) error {
	for _, f := range hooks {
		err := f(a1, a2)
		if err != nil {
			return err
		}
	}
	return nil
}

func runHooksFwd2A[T1, T2 any](hooks []func(a1 T1, a2 T2) any, a1 T1, a2 T2) any {
	for _, f := range hooks {
		r := f(a1, a2)
		if r != nil {
			return r
		}
	}
	return nil
}

func runHooksFwd3[T1, T2, T3 any](hooks []func(a1 T1, a2 T2, a3 T3), a1 T1, a2 T2, a3 T3) {
	for _, f := range hooks {
		f(a1, a2, a3)
	}
}

func runHooksRev3[T1, T2, T3 any](hooks []func(a1 T1, a2 T2, a3 T3), a1 T1, a2 T2, a3 T3) {
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i](a1, a2, a3)
	}
}

func runHooksFwd3Or[T1, T2, T3 any](hooks []func(a1 T1, a2 T2, a3 T3) bool, a1 T1, a2 T2, a3 T3) bool {
	for _, f := range hooks {
		if f(a1, a2, a3) {
			return true
		}
	}
	return false
}

func runHooksRevBltin2EUntil[T1, T2 any](hooks []func(a1 T1, a2 T2) error, bltin func(a1 T1, a2 T2) error, a1 T1, a2 T2, until *bool) error {
	for i := len(hooks) - 1; i >= 0; i-- {
		err := hooks[i](a1, a2)
		if err != nil || *until {
			return err
		}
	}
	return bltin(a1, a2)
}
