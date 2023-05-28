package hotwired

import (
	"bytes"
	"html"
)

const (
	StreamContentType = "text/vnd.turbo-stream.html"

	ActionNameAppend  = "append"
	ActionNamePrepend = "prepend"
)

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
