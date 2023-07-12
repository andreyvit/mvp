package mvphelpers

import (
	"fmt"
	"html/template"
)

func FuzzyHTMLAttrValue(v any) template.HTMLAttr {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return template.HTMLAttr(template.HTMLEscapeString(v))
	case template.HTMLAttr:
		return v
	case template.HTML:
		return template.HTMLAttr(v)
	default:
		return template.HTMLAttr(template.HTMLEscapeString(fmt.Sprint(v)))
	}
}

func FuzzyHTML(v any) template.HTML {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return template.HTML(template.HTMLEscapeString(v))
	case template.HTML:
		return template.HTML(v)
	default:
		return template.HTML(template.HTMLEscapeString(fmt.Sprint(v)))
	}
}
