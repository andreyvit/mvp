package forms

import (
	"fmt"
	"strings"
)

type Identity struct {
	ID       string
	FullName string
}

type Field struct {
	Identity

	RawFormValue        string
	RawFormValues       []string
	RawFormValuePresent bool
}

func (field *Field) EnumFields(f func(*Field)) {
	f(field)
}

type Subst struct {
	For string
	Use string
}
type Value struct {
	Key   string
	Value any
}

type Style struct {
	Templates []Subst
	Classes   map[string]string
	// Values    []Value
}

type TemplateStyle struct {
	Classes map[string]string
}

func (ts *TemplateStyle) TemplateStylePtr() *TemplateStyle {
	return ts
}

type State struct {
	Data *FormData

	path        []string
	errSites    []ErrorSite
	fields      map[string]*Field
	templKeys   []string
	templValues []string
	// values        map[string]any
	// valueKeys     []string
	// valueValues   []any
	classes       []map[string]string
	classesCopied []bool
}

func (st *State) PushName(name string) {
	if name != "" {
		st.path = append(st.path, name)
	}
}

func (st *State) PushTemplate(forTempl, useTempl string) {
	st.templKeys = append(st.templKeys, forTempl)
	st.templValues = append(st.templValues, useTempl)
}

// func (st *State) PushValue(key string, value any) {
// 	st.valueKeys = append(st.valueKeys, key)
// 	st.valueValues = append(st.valueValues, st.values[key])
// 	st.values[key] = value
// }

func (st *State) PushTemplates(substs []Subst) {
	for _, s := range substs {
		st.PushTemplate(s.For, s.Use)
	}
}

func (st *State) ensureClassesCopied() map[string]string {
	classes := st.classes[len(st.classes)-1]
	if cn := len(st.classesCopied); !st.classesCopied[cn-1] {
		st.classesCopied[cn-1] = true
		newClasses := make(map[string]string, len(classes)+10)
		for k, v := range classes {
			newClasses[k] = v
		}
		classes = newClasses
		st.classes = append(st.classes, classes)
	}
	return classes
}

func (st *State) PushClass(key string, value string) {
	classes := st.ensureClassesCopied()
	classes[key] = joinClasses(classes[key], value)
}

func (st *State) PushClasses(newClasses map[string]string) {
	classes := st.ensureClassesCopied()
	for key, value := range newClasses {
		classes[key] = joinClasses(classes[key], value)
	}
}

func (st *State) Classes() map[string]string {
	return st.classes[len(st.classes)-1]
}

func (st *State) TemplateStyle() TemplateStyle {
	return TemplateStyle{
		Classes: st.Classes(),
	}
}

func (st *State) PushStyle(s *Style) {
	st.PushTemplates(s.Templates)
	st.PushClasses(s.Classes)
}

func (st *State) PushStyles(styles []*Style) {
	for _, s := range styles {
		st.PushStyle(s)
	}
}

func (st *State) LookupTemplate(templ string) string {
	keys := st.templKeys
	for i := len(keys) - 1; i >= 0; i-- {
		if keys[i] == templ {
			return st.templValues[i]
		}
	}
	return templ
}

func (st *State) ErrorSite() ErrorSite {
	n := len(st.errSites)
	if n == 0 {
		return nil
	}
	return st.errSites[n-1]
}

func (st *State) PushErrorSite(errs ErrorSite) {
	if errs != nil {
		errs.Init(st.ErrorSite())
		st.errSites = append(st.errSites, errs)
	}
}

func (st *State) Fin() {
	if len(st.path) != 0 {
		panic(fmt.Errorf("ended with path = %v", st.path))
	}
}

func (st *State) AddField(field *Field) {
	st.AssignIdentity(&field.Identity)
	st.fields[field.FullName] = field
}

func (st *State) AssignIdentity(ident *Identity) {
	if ident.ID == "" {
		ident.ID = strings.Join(st.path, "_")
	}
	if ident.FullName == "" {
		ident.FullName = JoinNames(st.path...)
	}
}

func (st *State) finalizeTree(c Child) {
	if c == nil {
		return
	}
	pathLen := len(st.path)
	errSitesLen := len(st.errSites)
	templLen := len(st.templKeys)
	classesLen := len(st.classes)
	st.classesCopied = append(st.classesCopied, false)
	c.Finalize(st)
	if p, ok := c.(Processor); ok {
		p.EnumFields(st.AddField)
		p.EnumBindings(func(b AnyBinding) {
			b.Init(st.ErrorSite())
		})
	}
	if t, ok := c.(interface{ TemplateStylePtr() *TemplateStyle }); ok {
		stylePtr := t.TemplateStylePtr()
		if stylePtr != nil {
			*stylePtr = st.TemplateStyle()
		}
	}
	if cont, ok := c.(Container); ok {
		cont.EnumChildren(st.finalizeTree)
	}
	st.path = st.path[:pathLen]
	st.errSites = st.errSites[:errSitesLen]
	st.templKeys = st.templKeys[:templLen]
	st.templValues = st.templValues[:templLen]
	st.classes = st.classes[:classesLen]
	st.classesCopied = st.classesCopied[:len(st.classesCopied)-1]
}

func joinClasses(prev, next string) string {
	prev = strings.TrimSpace(prev)
	next = strings.TrimSpace(next)
	if next, ok := strings.CutPrefix(next, "+ "); ok {
		next = strings.TrimSpace(next)
		if prev != "" && next != "" {
			return prev + " " + next
		} else {
			return prev + next
		}
	}
	return next
}
