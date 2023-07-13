package mvp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/andreyvit/mvp/httperrors"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/andreyvit/mvp/mvputil"
	"github.com/uptrace/bunrouter"
	"golang.org/x/exp/maps"
)

type (
	// RouteBuilder helps to define named routes, and holds the path and middleware.
	RouteBuilder struct {
		app  *App
		bg   *bunrouter.Group
		path string
		routingContext
	}

	RouteOption     = any
	RouteFlagOption int
)

const (
	ReadOnly RouteFlagOption = 1 + iota
	Mutator
	IdempotentMutator
)

const (
	methodGetOrPost = "GET/POST"
)

var (
	funcNameRegexp = regexp.MustCompile(`^[A-Z][a-zA-Z0-9_]*$`)
)

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

	return g.addRoute(routeName, method, path, mi.Idempotent, f, options...)
}

func (g *RouteBuilder) Func(f any, options ...RouteOption) *Route {
	var funcName string
	var possibleFuncNameOpt string

	if len(options) > 0 {
		if s, ok := options[0].(string); ok {
			if funcNameRegexp.MatchString(s) {
				funcName = s
			} else {
				possibleFuncNameOpt = s
			}
		}
	}

	if funcName == "" {
		implFuncName := mvputil.FuncName(f)

		var ok bool
		funcName, ok = strings.CutPrefix(implFuncName, "do")
		if ok && !funcNameRegexp.MatchString(funcName) {
			ok = false
		}
		if !ok {
			if possibleFuncNameOpt == "" {
				panic(fmt.Errorf(`func %q must have "do" prefix followed by a valid func name matching %v (e.g. "doMyCall")`, implFuncName, funcNameRegexp))
			} else {
				panic(fmt.Errorf(`either func %q must have "do" prefix followed by a valid func name (e.g. "doMyCall"), or option %q must be a valid func name matching %v`, implFuncName, possibleFuncNameOpt, funcNameRegexp))
			}
		}
	}
	return g.addRoute(funcName, methodGetOrPost, "/"+funcName, false, f, options...)
}

func (g *RouteBuilder) addRoute(routeName, method, path string, isIdempotentByDefault bool, f any, options ...RouteOption) *Route {
	fv := reflect.ValueOf(f)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(fmt.Errorf(`%s: function is invalid, got %v, wanted a function`, routeName, ft))
	}
	if ft.NumOut() != 2 || ft.Out(1) != errorType {
		panic(fmt.Errorf(`%s: got %v, wanted a function returning (something, error)`, routeName, ft))
	}
	if ft.NumIn() != 2 || ft.In(0).Kind() != reflect.Ptr || ft.In(1).Kind() != reflect.Ptr || ft.In(1).Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf(`%s: got %v, wanted a function accepting (*RC, *SomeStruct)`, routeName, ft))
	}
	rcFacet := BaseRC.FacetByPtrType(ft.In(0))
	if rcFacet == nil {
		panic(fmt.Errorf(`%s: got %v, wanted a function accepting (*RC, *SomeStruct), where RC is any of RC facets`, routeName, ft))
	}

	// inTypPtr := ft.In(1)
	inTyp := ft.In(1).Elem()

	route := &Route{
		routeName:      routeName,
		method:         method,
		path:           g.path + path,
		funcVal:        fv,
		rcFacet:        rcFacet,
		inType:         inTyp,
		idempotent:     isIdempotentByDefault,
		routingContext: g.routingContext.clone(),
	}

	exampleInput := reflect.New(inTyp).Interface()
	exampleInputBytes := must(json.Marshal(exampleInput))
	var exampleInputMap map[string]any
	ensure(json.Unmarshal(exampleInputBytes, &exampleInputMap))
	bodyParamNames := maps.Keys(exampleInputMap)
	sort.Strings(bodyParamNames)
	route.bodyParamNames = bodyParamNames

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
		case RouteFlagOption:
			switch opt {
			case Mutator, IdempotentMutator:
				route.idempotent = false
			case ReadOnly:
				route.idempotent = true
			}
		default:
			panic(fmt.Errorf("%s: invalid option %T %v", routeName, opt, opt))
		}
	}

	var addPostAlternative bool
	if route.method == methodGetOrPost {
		if route.idempotent {
			route.method = http.MethodGet
			addPostAlternative = true
		} else {
			route.method = http.MethodPost
		}
	}
	route.desc = routeName + " " + method + " " + route.path

	if route.storeAffinity == mvpm.UnknownAffinity {
		if route.idempotent {
			route.storeAffinity = mvpm.SafeReader
		} else {
			route.storeAffinity = mvpm.SafeWriter
		}
	}

	if prev := g.app.routesByName[route.routeName]; prev != nil {
		panic(fmt.Errorf("route %s: duplicate path %s, previous was %s", route.routeName, route.method+" "+route.path, prev.method+" "+prev.path))
	}
	g.app.routesByName[route.routeName] = route

	handler := func(w http.ResponseWriter, req bunrouter.Request) error {
		rc := g.app.NewHTTPRequestRC(w, req)
		defer rc.Close()

		err := g.app.callRoute(route, rc, w, req)
		logRequest(rc, req.Request, err)
		if err != nil {
			http.Error(w, err.Error(), httperrors.HTTPCode(err))
		}
		return nil
	}

	g.bg.Handle(route.method, path, handler)
	if addPostAlternative {
		g.bg.Handle(http.MethodPost, path, handler)
	}

	return route
}
