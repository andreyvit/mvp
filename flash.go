package mvp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
)

type Mood int

const (
	MoodNeutral = Mood(iota)
	MoodSuccess
	MoodFailure
	MoodSubtle
)

var _moodStrings = [...]string{
	MoodNeutral: "neutral",
	MoodSuccess: "success",
	MoodFailure: "failure",
	MoodSubtle:  "subtle",
}

func (v Mood) String() string {
	return _moodStrings[v]
}
func ParseMood(s string) (Mood, error) {
	if i := slices.Index(_moodStrings[:], s); i >= 0 {
		return Mood(i), nil
	} else {
		return MoodNeutral, fmt.Errorf("invalid Mood %q", s)
	}
}
func (v Mood) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
func (v *Mood) UnmarshalText(b []byte) error {
	var err error
	*v, err = ParseMood(string(b))
	return err
}
func (v Mood) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeUint(uint64(v))
}
func (v *Mood) DecodeMsgpack(dec *msgpack.Decoder) error {
	n, err := dec.DecodeUint()
	*v = Mood(n)
	return err
}

type Msg struct {
	Text     string `json:"t,omitempty"`
	Link     string `json:"l,omitempty"`
	LinkText string `json:"lt,omitempty"`
	Mood     Mood   `json:"m,omitempty"`
}

func (msg *Msg) Success() bool {
	return msg.Mood == MoodSuccess
}

func (msg *Msg) Failure() bool {
	return msg.Mood == MoodFailure
}

func (msg *Msg) Encode() string {
	if msg == nil || *msg == (Msg{}) {
		return ""
	}
	return string(must(json.Marshal(msg)))
}

func DecodeMsg(raw string) *Msg {
	if raw == "" {
		return nil
	}
	msg := new(Msg)
	_ = json.Unmarshal([]byte(raw), msg)
	if *msg == (Msg{}) {
		return nil
	}
	return msg
}

func SubtleMsg(text string) *Flash {
	return &Flash{Msg: RawSubtleMsg(text)}
}
func SuccessMsg(text string) *Flash {
	return &Flash{Msg: RawSuccessMsg(text)}
}
func FailureMsg(text string) *Flash {
	return &Flash{Msg: RawFailureMsg(text)}
}
func NeutralMsg(text string) *Flash {
	return &Flash{Msg: RawNeutralMsg(text)}
}

func RawSubtleMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodSubtle}
}
func RawSuccessMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodSuccess}
}
func RawFailureMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodFailure}
}
func RawNeutralMsg(text string) *Msg {
	return &Msg{Text: text, Mood: MoodNeutral}
}

type Flash struct {
	Msg          *Msg           `json:"m,omitempty"`
	Action       string         `json:"a,omitempty"`
	Target       string         `json:"t,omitempty"`
	ScrollTarget string         `json:"s,omitempty"`
	Extras       map[string]any `json:"e,omitempty"`
}

func NewFlash() *Flash {
	return &Flash{}
}

func (flash *Flash) WithAction(action string) *Flash {
	flash.Action = action
	return flash
}
func (flash *Flash) ScrollTo(id string) *Flash {
	flash.ScrollTarget = id
	return flash
}

func (flash *Flash) WithTarget(target string) *Flash {
	flash.Target = target
	return flash
}

func (flash *Flash) WithExtra(key string, value any) *Flash {
	if flash.Extras == nil {
		flash.Extras = make(map[string]any)
	}
	flash.Extras[key] = value
	return flash
}

func (flash *Flash) JSONBytes() []byte {
	if flash == nil {
		return nil
	}
	return must(json.Marshal(flash))
}
func (flash *Flash) JSONString() string {
	if flash == nil {
		return "null"
	}
	return string(flash.JSONBytes())
}
func (flash *Flash) SafeJSON() template.JS {
	if flash == nil {
		return "null"
	}
	return template.JS(flash.JSONString())
}

const (
	flashParam = "flash"
)

func encodeFlash(query url.Values, flash *Flash) {
	if flash == nil {
		return
	}
	query.Set(flashParam, flash.JSONString())
}

func DecodeFlashFromQuery(query url.Values) (*Flash, error) {
	if raw := query.Get("flash"); raw != "" {
		flash := new(Flash)
		err := json.Unmarshal([]byte(raw), flash)
		if err != nil {
			return nil, fmt.Errorf("flash decoding: %w", err)
		}
		return flash, nil
	} else if str := query.Get("msg"); str != "" {
		return NeutralMsg(str), nil
	} else {
		return nil, nil
	}
}

func DecodeFlashIntoRC(rc *RC) {
	var err error
	rc.Flash, err = DecodeFlashFromQuery(rc.Request.URL.Query())
	if err != nil {
		rc.Logf("ERROR: %v", err)
	}
}
