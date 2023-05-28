package mvp

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/andreyvit/mvp/httperrors"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/uptrace/bunrouter"
)

// RouteBuilder helps to define named routes, and holds the path and middleware.
type RouteBuilder struct {
	app  *App
	bg   *bunrouter.Group
	path string
	routingContext
}

func (g *RouteBuilder) Raw() *bunrouter.Group {
	return g.bg
}

// group defines a group of routes with a common prefix and/or middleware.
func (g *RouteBuilder) Group(path string, f func(b *RouteBuilder)) {
	sg := RouteBuilder{
		app:            g.app,
		bg:             g.bg.NewGroup(path),
		path:           g.path + path,
		routingContext: g.routingContext.clone(),
	}
	f(&sg)
}

func (g *RouteBuilder) Static(path string) {
	setupStaticServer(g.bg, path, g.app.staticFS)
}

type (
	RouteOption         = any
	nonIdempotentOption struct{}
)

var (
	NonIdempotent = nonIdempotentOption{}
)

// Route defines a named route. methodAndPath are space-separated.
//
// The handler function must have func(rc *RC, in *SomeStruct) (output, error) signature,
// where output can be anything, but must be something that app.writeResponse accepts.
func (g *RouteBuilder) Route(routeName string, methodAndPath string, f any, options ...RouteOption) *Route {
	method, path, ok := strings.Cut(methodAndPath, " ")
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string`, routeName, methodAndPath))
	}
	mi, ok := validHTTPMethods[method]
	if !ok {
		panic(fmt.Errorf(`%s: %q is not a valid "METHOD path" string, method %q is invalid`, routeName, methodAndPath, method))
	}

	desc := routeName + " " + methodAndPath

	fv := reflect.ValueOf(f)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(fmt.Errorf(`%s: function is invalid, got %v, wanted a function`, desc, ft))
	}
	if ft.NumOut() != 2 || ft.Out(1) != errorType {
		panic(fmt.Errorf(`%s: got %v, wanted a function returning (something, error)`, desc, ft))
	}
	if ft.NumIn() != 2 || ft.In(0).Kind() != reflect.Ptr || ft.In(1).Kind() != reflect.Ptr || ft.In(1).Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf(`%s: got %v, wanted a function accepting (*RC, *SomeStruct)`, desc, ft))
	}
	rcFacet := BaseRC.FacetByPtrType(ft.In(0))
	if rcFacet == nil {
		panic(fmt.Errorf(`%s: got %v, wanted a function accepting (*RC, *SomeStruct), where RC is any of RC facets`, desc, ft))
	}

	// inTypPtr := ft.In(1)
	inTyp := ft.In(1).Elem()

	route := &Route{
		desc:           desc,
		routeName:      routeName,
		method:         method,
		path:           g.path + path,
		funcVal:        fv,
		rcFacet:        rcFacet,
		inType:         inTyp,
		idempotent:     mi.Idempotent,
		routingContext: g.routingContext.clone(),
	}

	for _, param := range pathParamsRe.FindAllString(route.path, -1) {
		route.pathParams = append(route.pathParams, param[1:])
	}

	for _, opt := range options {
		switch opt := opt.(type) {
		case RateLimitPreset:
			route.rateLimitPreset = opt
		case mvpm.StoreAffinity:
			route.storeAffinity = opt
			if opt.IsWriter() {
				route.idempotent = false
			}
		case nonIdempotentOption:
			route.idempotent = false
		default:
			panic(fmt.Errorf("%s: invalid option %T %v", desc, opt, opt))
		}
	}

	if route.storeAffinity == mvpm.UnknownAffinity {
		if mi.Idempotent {
			route.storeAffinity = mvpm.SafeReader
		} else {
			route.storeAffinity = mvpm.SafeWriter
		}
	}

	if prev := g.app.routesByName[route.routeName]; prev != nil {
		panic(fmt.Errorf("route %s: duplicate path %s, previous was %s", route.routeName, methodAndPath, prev.method+" "+prev.path))
	}
	g.app.routesByName[route.routeName] = route

	g.bg.Handle(method, path, func(w http.ResponseWriter, req bunrouter.Request) error {
		rc := g.app.NewHTTPRequestRC(w, req)
		defer rc.Close()

		err := g.app.callRoute(route, rc, w, req)
		logRequest(rc, req.Request, err)
		if err != nil {
			http.Error(w, err.Error(), httperrors.HTTPCode(err))
		}
		return nil
	})

	return route
}
