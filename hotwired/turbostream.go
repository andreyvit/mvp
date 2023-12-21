package hotwired

import (
	"bytes"
	"html"
	"net/http"
	"strings"

	"github.com/andreyvit/mvp/mvphelpers"
)

const (
	StreamContentType = "text/vnd.turbo-stream.html"

	ActionNameAppend  = "append"
	ActionNamePrepend = "prepend"
)

func IsTurbo(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), StreamContentType)
}

func TurboFrameName(r *http.Request) string {
	return r.Header.Get("Turbo-Frame")
}

type StreamData []byte

type Stream struct {
	Buffer bytes.Buffer
}

func (stream *Stream) Append(target string, safeContent string) {
	stream.Buffer.WriteString(`<turbo-stream action="append" target="`)
	stream.Buffer.WriteString(html.EscapeString(target))
	stream.Buffer.WriteString(`"><template>`)
	stream.Buffer.WriteString(safeContent)
	stream.Buffer.WriteString(`</template></turbo-stream>`)
}

func (stream *Stream) Prepend(target string, safeContent string) {
	stream.Buffer.WriteString(`<turbo-stream action="prepend" target="`)
	stream.Buffer.WriteString(html.EscapeString(target))
	stream.Buffer.WriteString(`"><template>`)
	stream.Buffer.WriteString(safeContent)
	stream.Buffer.WriteString(`</template></turbo-stream>`)
}

func (stream *Stream) Replace(target string, safeContent string) {
	stream.Buffer.WriteString(`<turbo-stream action="replace" target="`)
	stream.Buffer.WriteString(html.EscapeString(target))
	stream.Buffer.WriteString(`"><template>`)
	stream.Buffer.WriteString(safeContent)
	stream.Buffer.WriteString(`</template></turbo-stream>`)
}

func (stream *Stream) Custom(action string, attrs map[string]any) {
	stream.startCustom(action, attrs)
	stream.Buffer.WriteString(`></turbo-stream>`)
}

func (stream *Stream) CustomContent(action string, attrs map[string]any, safeContent string) {
	stream.startCustom(action, attrs)
	stream.Buffer.WriteString(`><template>`)
	stream.Buffer.WriteString(safeContent)
	stream.Buffer.WriteString(`</template></turbo-stream>`)
}

func (stream *Stream) startCustom(action string, attrs map[string]any) {
	stream.Buffer.WriteString(`<turbo-stream action="`)
	stream.Buffer.WriteString(html.EscapeString(action))
	stream.Buffer.WriteByte('"')
	var w mvphelpers.StringByteWriter = &stream.Buffer
	for k, v := range attrs {
		mvphelpers.AppendAttrAny(w, k, v)
	}
}

func (stream *Stream) Build() StreamData {
	return stream.Buffer.Bytes()
}
