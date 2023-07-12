package mvphelpers

import (
	"html/template"
	"strings"
)

func AppendAttr(buf *strings.Builder, k, v string) {
	buf.WriteByte(' ')
	buf.WriteString(k)
	buf.WriteString(`="`)
	buf.WriteString(template.HTMLEscapeString(v))
	buf.WriteByte('"')
}

func AppendAttrAny(buf *strings.Builder, k string, v any) {
	if !FuzzyBool(v) {
		return
	}
	buf.WriteByte(' ')
	buf.WriteString(k)
	if v == true {
		return
	}
	buf.WriteString(`="`)
	buf.WriteString(string(FuzzyHTMLAttrValue(v)))
	buf.WriteByte('"')
}
