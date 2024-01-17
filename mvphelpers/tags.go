package mvphelpers

import (
	"html/template"
	"strings"
)

func OptionTag(htmlValue string, modelValue, currentModelValue any, label any, attrs ...any) template.HTML {
	var buf strings.Builder
	buf.WriteString("<option")
	if htmlValue != "" {
		AppendAttr(&buf, "value", htmlValue)
	}
	if currentModelValue == modelValue {
		AppendAttrAny(&buf, "selected", true)
	}
	for k, v := range Dict(attrs...) {
		AppendAttrAny(&buf, k, v)
	}
	buf.WriteString(">")
	buf.WriteString(string(FuzzyHTML(label)))
	buf.WriteString("</option>")
	return template.HTML(buf.String())
}
