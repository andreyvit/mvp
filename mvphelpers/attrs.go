package mvphelpers

import (
	"html/template"
	"io"
)

type StringByteWriter interface {
	io.StringWriter
	io.ByteWriter
}

func AppendAttr(buf StringByteWriter, k, v string) {
	buf.WriteByte(' ')
	buf.WriteString(k)
	buf.WriteString(`="`)
	buf.WriteString(template.HTMLEscapeString(v))
	buf.WriteByte('"')
}

func AppendAttrAny(buf StringByteWriter, k string, v any) {
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
