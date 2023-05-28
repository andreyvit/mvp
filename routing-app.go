package mvp

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/andreyvit/httpform"
	"github.com/uptrace/bunrouter"
)

func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := TrimPort(r.Host)
	d := app.domainRouter.find(domain)
	if d == nil {
		http.Error(w, fmt.Sprintf("Invalid domain %q (wanted %s)", domain, strings.Join(app.domainRouter.ValidDomains(), ", ")), http.StatusMisdirectedRequest)
		return
	}
	switch d := d.(type) {
	case *Site:
		app.siteRouters[d].ServeHTTP(w, r)
	case http.Handler:
		d.ServeHTTP(w, r)
	}
}

func initRouting(app *App) {
	var root routingContext
	runHooksFwd1(app.Hooks.middleware, Router(&root))

	app.siteRouters = make(map[*Site]*bunrouter.Router)
	for site, siteHooks := range app.Hooks.siteRoutes {
		r := bunrouter.New()
		b := RouteBuilder{
			app:            app,
			bg:             &r.Group,
			path:           "",
			routingContext: root.clone(),
		}
		runHooksFwd1(siteHooks, &b)
		app.siteRouters[site] = r
	}

	dr := &DomainRouter{}
	if app.BaseURL != nil && app.BaseURL.Host != "" {
		dr.Site(TrimPort(app.BaseURL.Host), DefaultSite)
	}
	runHooksFwd2(app.Hooks.domainRoutes, app, dr)
	app.domainRouter = dr
}

// callRoute handles an HTTP request using the middleware, handler and parameters of the given route.
//
// Note: we might be doing too much here, some logic should probably be moved into middleware.
func (app *App) callRoute(route *Route, rc *RC, w http.ResponseWriter, req bunrouter.Request) error {
	rc.Route = route
	if rc.RateLimitPreset == "" {
		rc.RateLimitPreset = route.rateLimitPreset
	}

	if !route.idempotent {
		// TODO: check CSRF
	}

	inVal := reflect.New(route.inType)
	err := httpform.Default.DecodeVal(req.Request, req.Params(), inVal)
	if err != nil {
		return err
	}

	var output any

	err = app.InTx(rc, route.storeAffinity, func() error {
		for _, mw := range route.middleware {
			if mw.f == nil {
				continue
			}
			// if mw.name != "" {
			// 	flogger.Log(rc, "middleware %s", mw.name)
			// }
			var err error
			output, err = mw.f(rc)
			if err != nil {
				return err
			}
			if output != nil {
				return nil
			}
		}

		inputs := make([]reflect.Value, route.funcVal.Type().NumIn())
		inputs[0] = reflect.ValueOf(route.rcFacet.AnyFrom(rc))
		inputs[1] = inVal

		results := route.funcVal.Call(inputs)
		output = results[0].Interface()
		if errVal := results[1].Interface(); errVal != nil {
			return errVal.(error)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return app.writeResponse(rc, output, route, w, req.Request)
}
