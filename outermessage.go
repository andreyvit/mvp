package mvp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
)

type OuterMessage struct {
	Event string         `json:"event"`
	Data  map[string]any `json:"data,omitempty"`
}

const outerMessageParam = "outer_message"

func (m *OuterMessage) JSONBytes() []byte {
	if m == nil {
		return nil
	}
	return must(json.Marshal(m))
}

func (m *OuterMessage) JSONString() string {
	if m == nil {
		return ""
	}
	return string(m.JSONBytes())
}

func (m *OuterMessage) SafeJSON() template.JS {
	if m == nil {
		return "null"
	}
	return template.JS(m.JSONString())
}

func encodeOuterMessage(query url.Values, msg *OuterMessage) {
	if msg == nil {
		return
	}
	query.Set(outerMessageParam, msg.JSONString())
}

func DecodeOuterMessageFromQuery(query url.Values) (*OuterMessage, error) {
	raw := query.Get(outerMessageParam)
	if raw == "" {
		return nil, nil
	}
	msg := new(OuterMessage)
	err := json.Unmarshal([]byte(raw), msg)
	if err != nil {
		return nil, fmt.Errorf("outer_message decoding: %w", err)
	}
	return msg, nil
}
