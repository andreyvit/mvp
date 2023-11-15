package mvp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/andreyvit/minicomponents"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/mvphelpers"
	"golang.org/x/exp/maps"
)

func RegisterBuiltinUtilityViewHelpers(m template.FuncMap) {
	for k, v := range mvphelpers.FuncMap() {
		m[k] = v
	}
	m["attr"] = Attr
	m["attrs"] = AttrsSwitch
	m["error"] = func(text string) template.HTML {
		panic(fmt.Errorf("%s", text))
	}
	m["iif"] = func(cond any, trueVal any, falseVal any) any {
		if mvphelpers.FuzzyBool(cond) {
			return trueVal
		} else {
			return falseVal
		}
	}
	m["repeat"] = func(n int) []int {
		r := make([]int, 0, n)
		for i := 1; i <= n; i++ {
			r = append(r, i)
		}
		return r
	}
	m["pick"] = func(i int, values ...any) any {
		return values[i%len(values)]
	}
	m["switch"] = func(actual any, items ...any) any {
		for i := 0; i < len(items)-1; i += 2 {
			if i+1 >= len(items) {
				// "else" clause
				return items[i]
			}
			if items[i] == actual {
				return items[i+1]
			}
		}
		return nil
	}
	m["switchstr"] = func(actual any, items ...any) any {
		actualStr := fmt.Sprint(actual)
		for i := 0; i < len(items)-1; i += 2 {
			if i+1 >= len(items) {
				// "else" clause
				return items[i]
			}
			if fmt.Sprint(items[i]) == actualStr {
				return items[i+1]
			}
		}
		return nil
	}
	m["classes"] = JoinClasses
	m["add"] = func(a, b any) int {
		return mvphelpers.FuzzyInt(a) + mvphelpers.FuzzyInt(b)
	}
	m["addf"] = func(a, b any) float64 {
		return mvphelpers.FuzzyFloat64(a) + mvphelpers.FuzzyFloat64(b)
	}
	m["mulf"] = func(a, b any) float64 {
		return mvphelpers.FuzzyFloat64(a) * mvphelpers.FuzzyFloat64(b)
	}
	m["addone"] = func(v any) int {
		return mvphelpers.FuzzyInt(v) + 1
	}
	m["gridoffsetf"] = func(start, size, gap, coord any) float64 {
		return mvphelpers.FuzzyFloat64(start) + (mvphelpers.FuzzyFloat64(size)+mvphelpers.FuzzyFloat64(gap))*mvphelpers.FuzzyFloat64(coord)
	}
	m["gridsizef"] = func(base, size, gap, coord any) float64 {
		return mvphelpers.FuzzyFloat64(base) + mvphelpers.FuzzyFloat64(size)*mvphelpers.FuzzyFloat64(coord) + mvphelpers.FuzzyFloat64(gap)*(mvphelpers.FuzzyFloat64(coord)-1)
	}
	m["div"] = func(a, b any) float64 {
		return mvphelpers.FuzzyFloat64(a) / mvphelpers.FuzzyFloat64(b)
	}
	m["defined"] = func(cond any) bool {
		return cond != nil
	}
	m["json"] = func(v any) template.JS {
		return template.JS(string(must(json.Marshal(v))))
	}
	m["jsonpp"] = func(v any) template.JS {
		return template.JS(string(must(json.MarshalIndent(v, "", "    "))))
	}
	m["indentrawjson"] = func(v json.RawMessage) template.JS {
		if len(v) == 0 {
			return ""
		}
		var buf bytes.Buffer
		ensure(json.Indent(&buf, []byte(v), "", "    "))
		return template.JS(buf.String())
	}
	m["dump"] = func(v any) string {
		return fmt.Sprintf("%T %v", v, v)
	}
	m["fallback"] = func(values ...any) any {
		for _, v := range values {
			if mvphelpers.FuzzyBool(v) {
				return v
			}
		}
		return nil
	}
	m["kv"] = minicomponents.Args
	m["multipar"] = func(str any) template.HTML {
		return HTMLifyMultiparValue(str)
	}
}

func (app *App) registerBuiltinViewHelpers(m template.FuncMap) {
	RegisterBuiltinUtilityViewHelpers(m)
	m["c_link"] = app.renderLink
	m["c_icon"] = app.renderIcon
	m["c_image"] = app.renderIcon
	m["c_func_button"] = app.renderFuncButton
	m["eval"] = app.EvalTemplate
	m["url_for"] = func(d *RenderData, name string, extras ...any) template.URL {
		defaults := d.DefaultPathParams()
		if len(defaults) > 0 {
			newExtras := make([]any, 0, len(extras)+1)
			newExtras = append(newExtras, defaults)
			newExtras = append(newExtras, extras...)
			extras = newExtras
		}
		return template.URL(d.App.URL(name, extras...))
	}
}

func (app *App) renderLink(data *RenderData) template.HTML {
	// defer func() { mvphelpers.ExposeHelperPanic(recover()) }()

	classes := make([]string, 0, 8)

	href, _ := data.PopString("href")
	routeName, _ := data.PopString("route")
	body, _ := data.PopHTMLSafeString("body")
	classAttr, _ := data.PopString("class")
	inactiveClassAttr, _ := data.PopString("inactive-class")
	activeClassAttr, _ := data.PopString("active-class")
	if activeClassAttr == "" {
		activeClassAttr = "active"
	}
	sempathAttr, _ := data.PopString("sempath")
	iconAttr, _ := data.PopString("icon")
	iconClass, _ := data.PopString("icon-class")
	pathParams, _ := data.PopMapSA("path-params")

	var isActive, looksActive bool
	if href != "" {
		if sempathAttr != "" {
			looksActive = data.IsActive(sempathAttr)
			isActive = looksActive
		}
	} else if routeName != "" {
		route := app.routesByName[routeName]
		if route == nil {
			panic(fmt.Errorf("unknown route %s", routeName))
		}

		params := data.DefaultPathParams()
		for _, k := range route.pathParams {
			v, found := data.PopString(k)
			if found {
				params[k] = v
			} else if ppv, found := pathParams[k]; found {
				params[k] = fmt.Sprint(ppv)
			} else if _, found = params[k]; !found {
				panic(fmt.Errorf("route %s requires path param %s (pathParams.keys = %v, args.keys = %v)", routeName, k, maps.Keys(pathParams), maps.Keys(data.Args)))
			}
		}

		var qs url.Values
		for k := range data.Args {
			if s, ok := strings.CutPrefix(k, "query-"); ok {
				if qs == nil {
					qs = make(url.Values)
				}
				v, _ := data.PopString(k)
				s = strings.ReplaceAll(s, "-", "_")
				qs.Set(s, v)
			}
		}

		href = app.URL(routeName, params, qs)
		isActive = (data.Route != nil && data.Route.routeName == routeName)
		looksActive = isActive
		if sempathAttr != "" {
			looksActive = data.IsActive(sempathAttr)
			if !looksActive {
				// isActive is based on route name, but sempath might compare actual instances
				isActive = false
			}
		}
	}

	var iconStr template.HTML
	if iconAttr != "" {
		iconStr = app.renderIcon(data.Bind(nil, "class", JoinClasses("icon", iconClass), "src", iconAttr))
		classes = append(classes, "with-icon")
	}

	classes = append(classes, strings.Fields(classAttr)...)

	if looksActive {
		classes = AddClasses(classes, activeClassAttr)
	} else {
		classes = AddClasses(classes, inactiveClassAttr)
	}

	var extraArgs strings.Builder
	for k, v := range data.Args {
		if isPassThruArg(k) {
			extraArgs.WriteString(string(Attr(k, v)))
		} else {
			panic(fmt.Errorf("<c-link>: invalid param %s", k))
		}
	}
	if len(classes) > 0 {
		extraArgs.WriteString(string(Attr("class", JoinClassList(classes))))
	}
	if looksActive {
		extraArgs.WriteString(` aria-current="page"`)
	}

	if isActive || href == "" {
		return template.HTML(fmt.Sprintf(`<div%s>%s%s</div>`, extraArgs.String(), iconStr, body))
	} else {
		return template.HTML(fmt.Sprintf(`<a href="%s"%s>%s%s</a>`, href, extraArgs.String(), iconStr, body))
	}
}

func (app *App) renderFuncButton(data *RenderData) template.HTML {
	funcName, _ := data.PopString("func")
	route := app.routesByName[funcName]
	if route == nil {
		panic(fmt.Errorf("unknown func %q", funcName))
	}

	params := make(map[string]any)
	// log.Printf("<c-func-button>: %s: params = %v", funcName, route.BodyParamNames())
	for _, name := range route.BodyParamNames() {
		k := strings.ReplaceAll(name, "_", "-")
		v, ok := data.PopValue(k)
		if ok {
			params[name] = v
		}
	}

	formID, _ := data.PopString("form")
	if formID == "" {
		formID = app.Configuration.FallbackFormID
	}
	classAttr, _ := data.PopString("class")

	body, _ := data.PopValue("body")

	var buf strings.Builder
	buf.WriteString(`<button type="submit"`)
	if formID != "outer" && formID != "" {
		mvphelpers.AppendAttr(&buf, "form", formID)
	}
	mvphelpers.AppendAttr(&buf, "formmethod", route.Method())
	mvphelpers.AppendAttr(&buf, "formaction", route.Path())
	if classAttr != "" {
		mvphelpers.AppendAttr(&buf, "class", classAttr)
	}

	for k, v := range data.Args {
		// k = strings.ReplaceAll(k, "_", "-")
		if isPassThruArg(k) {
			mvphelpers.AppendAttrAny(&buf, k, v)
		} else {
			panic(fmt.Errorf("<c-func-button>: invalid param %s", k))
		}
	}

	if len(params) > 0 {
		bodyBytes := must(json.Marshal(params))
		mvphelpers.AppendAttr(&buf, "name", "_body")
		mvphelpers.AppendAttr(&buf, "value", string(bodyBytes))
	}
	buf.WriteString(">")
	buf.WriteString(string(mvphelpers.FuzzyHTMLAttrValue(body)))
	buf.WriteString("</button>")
	return template.HTML(buf.String())
}

func (app *App) renderIcon(data *RenderData) template.HTML {
	var src string
	var srcFound bool
	var extraArgs strings.Builder
	for k, v := range data.Args {
		if k == "src" {
			src = fmt.Sprint(v)
			srcFound = true
		} else {
			extraArgs.WriteString(string(Attr(k, v)))
		}
	}
	if !srcFound {
		panic("<c-icon>: missing src attribute")
	}
	if src == "" {
		return ""
	}

	if strings.HasSuffix(src, ".svg") {
		raw, err := data.BaseRC().ReadFile(src)
		if err != nil {
			flogger.Log(data.LC(), "WARNING: <c-icon>: %v", err)
			return ""
		}

		svgStart := svgStartRe.FindSubmatchIndex(raw)
		if !utf8.Valid(raw) || svgStart == nil {
			flogger.Log(data.LC(), "WARNING: <c-icon>: %s is not an SVG", src)
			return ""
		}
		body := string(raw)

		if extraArgs.Len() > 0 {
			i := svgStart[3]
			body = body[:i] + extraArgs.String() + body[i:]
		}
		body = strings.Replace(body, `xmlns="http://www.w3.org/2000/svg"`, ``, 1)
		return template.HTML(body)
	} else {
		url := data.BaseRC().FileURL(src, 0)
		if url == "" {
			flogger.Log(data.LC(), "WARNING: <c-icon>: cannot create URL for %q", src)
			return ""
		}

		return template.HTML(fmt.Sprintf(`<img src="%s"%s>`, url, extraArgs.String()))
	}
}

var svgStartRe = regexp.MustCompile(`(<svg)\s`)

func Attr(name string, value any) template.HTMLAttr {
	if !mvphelpers.FuzzyBool(value) {
		return ""
	}
	if value == true {
		return template.HTMLAttr(" " + name)
	}
	return template.HTMLAttr(" " + name + "=\"" + string(mvphelpers.FuzzyHTMLAttrValue(value)) + "\"")
}

func AttrsSwitch(attrs any) template.HTMLAttr {
	switch attrs := attrs.(type) {
	case map[string]string:
		return Attrs(attrs)
	case map[string]any:
		return AttrsAny(attrs)
	default:
		panic(fmt.Errorf("attrs: invalid value %T %v", attrs, attrs))
	}
}

func Attrs(attrs map[string]string) template.HTMLAttr {
	var buf strings.Builder
	for k, v := range attrs {
		mvphelpers.AppendAttr(&buf, k, v)
	}
	return template.HTMLAttr(buf.String())
}

func AttrsAny(attrs map[string]any) template.HTMLAttr {
	var buf strings.Builder
	for k, v := range attrs {
		mvphelpers.AppendAttrAny(&buf, k, v)
	}
	return template.HTMLAttr(buf.String())
}
