package mvp

import (
	"fmt"
	"net/url"
	"strings"

	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

var (
	Absolute = mvpm.NewURLOption("absolute")
)

type (
	PathParamsMapStr map[string]string
	PathParamsMapAny map[string]any
)

// URL generates a relative (default) or absolute URL based on a named route.
// Supply map[string]string or pairs of (key string, value string) arguments
// to supply path parameters for the route.
// Supply url.Values to add query parameters.
// Supply Absolute flag to return an absolute URL.
func (app *App) URL(name string, extras ...any) string {
	route := app.routesByName[name]
	if route == nil {
		panic(fmt.Errorf("route %s does not exist", name))
	}

	var g URLGen
	for i := 0; i < len(extras); i++ {
		switch extra := extras[i].(type) {
		case PathParamsMapStr:
			if g.PathKeys == nil {
				g.PathKeys = extra
			} else {
				for k, v := range extra {
					g.PathKeys[k] = v
				}
			}
		case PathParamsMapAny:
			if g.PathKeys == nil {
				g.PathKeys = make(map[string]string)
			}
			for k, v := range extra {
				if v == nil {
					continue
				}
				g.PathKeys[k] = fmt.Sprint(v)
			}
		case url.Values:
			if g.QueryParams == nil {
				g.QueryParams = extra
			} else {
				for k, vv := range extra {
					g.QueryParams[k] = vv
				}
			}
		case string:
			if strings.HasPrefix(extra, "#") {
				switch extra {
				case "#abs":
					g.Absolute = true
				default:
					if !runHooksFwd3Or(app.Hooks.urlGenOption, app, &g, extra) {
						panic(fmt.Errorf("route %s: invalid option %q", name, extra))
					}
				}
			} else if s, ok := strings.CutPrefix(extra, "?"); ok {
				if i+1 >= len(extras) {
					panic(fmt.Errorf("route %s: no value following query param %q", name, extra))
				}
				i++
				if g.QueryParams == nil {
					g.QueryParams = make(url.Values)
				}
				g.QueryParams.Set(s, fmt.Sprint(extras[i]))
			} else if s, ok := strings.CutPrefix(extra, ":"); ok {
				if i+1 >= len(extras) {
					panic(fmt.Errorf("route %s: no value following extra key %q", name, extra))
				}
				i++
				if g.PathKeys == nil {
					g.PathKeys = make(map[string]string)
				}
				g.PathKeys[s] = fmt.Sprint(extras[i])
			}
		case mvpm.URLOption:
			if extra == Absolute {
				g.Absolute = true
			}
		default:
			panic(fmt.Errorf("route %s: unsupported extra %T %v", name, extra, extra))
		}
	}

	path := route.path
	if g.PathKeys != nil {
		for k, v := range g.PathKeys {
			if !strings.Contains(path, ":"+k) {
				panic(fmt.Errorf("route %s: unsupported path param %q", name, k))
			}
			path = strings.ReplaceAll(path, ":"+k, v)
		}
	}

	if g.Absolute {
		g.URL = *app.BaseURL
	}
	g.Path = path
	if g.QueryParams != nil {
		g.RawQuery = g.QueryParams.Encode()
	}

	runHooksFwd2(app.Hooks.urlGen, app, &g)

	if strings.Contains(path, ":") {
		panic(fmt.Errorf("URL(%s, %#v): not all path params specified in %q, effective path keys = %v", name, extras, path, g.PathKeys))
	}

	// log.Printf("URL(%s, %v) = %s", name, extras, g.URL.String())
	return g.URL.String()
}

type URLGen struct {
	url.URL
	Absolute    bool
	Options     []string
	PathKeys    map[string]string
	QueryParams url.Values
}

// Redirect returns a response object that redirects to a given route.
// Status code defaults to 303 See Other.
// Supply map[string]string or pairs of (key string, value string) arguments
// to supply path parameters for the route.
// Supply url.Values to add query parameters.
// Supply Absolute flag to use an absolute URL.
func (app *App) Redirect(name string, extras ...any) *Redirect {
	path := app.URL(name, extras...)
	return &Redirect{
		Path:      path,
		RouteName: name,
	}
}

func (rc *RC) DefaultPathParams() PathParamsMapStr {
	// TODO: hook
	return nil
}

func (rc *RC) Redirect(name string, extras ...any) *Redirect {
	dflts := rc.DefaultPathParams()
	if extras == nil {
		extras = []any{dflts}
	} else {
		e := make([]any, 0, len(extras)+1)
		e = append(e, dflts)
		e = append(e, extras...)
		extras = e
	}
	return rc.app.Redirect(name, extras...)
}

func (rc *RC) RedirectBack() *Redirect {
	referer := rc.Request.Header.Get("Referer")
	// log.Printf("Referer = %q", referer)
	return &Redirect{Path: referer}
}
