package mvp

import (
	"encoding/json"
	"net/url"
	"testing"
)

func TestMsg_signature_survives_relays(t *testing.T) {
	for _, tc := range []struct {
		name      string
		field     string
		signature string
	}{
		{"absent", "", ""},
		{"valid", `,"s":"proof"`, "proof"},
		{"empty", `,"s":""`, ""},
		{"null", `,"s":null`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			msg := DecodeMsg(`{"t":"hello","l":"/rewards","lt":"Rewards","m":"success"` + tc.field + `}`)
			if msg == nil || msg.Text != "hello" || msg.Link != "/rewards" || msg.LinkText != "Rewards" || msg.Mood != MoodSuccess || msg.Signature != tc.signature {
				t.Fatalf("decoded message = %#v", msg)
			}
			if tc.signature != "" && *msg != (Msg{Text: "hello", Link: "/rewards", LinkText: "Rewards", Mood: MoodSuccess, Signature: tc.signature}) {
				t.Fatalf("valid signature changed message equality: %#v", msg)
			}
			for _, value := range []any{msg, *msg} {
				b, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				got := DecodeMsg(string(b))
				if got == nil || *got != *msg {
					t.Fatalf("round trip = %#v, want %#v", got, msg)
				}
			}
			flash := &Flash{Msg: msg, Action: "popup", Target: "message", Commands: []*FlashCommand{{Name: "refresh"}}}
			got, err := DecodeFlashFromQuery(url.Values{"flash": {flash.JSONString()}})
			if err != nil {
				t.Fatal(err)
			}
			if got.Msg == nil || *got.Msg != *msg || got.Action != "popup" || got.Target != "message" || len(got.Commands) != 1 {
				t.Fatalf("flash round trip = %#v", got)
			}
			outer := &OuterMessage{Event: "closePopup", Data: map[string]any{"flash": flash}}
			relay, err := DecodeOuterMessageFromQuery(url.Values{"outer_message": {outer.JSONString()}})
			if err != nil {
				t.Fatal(err)
			}
			b, err := json.Marshal(relay.Data["flash"])
			if err != nil {
				t.Fatal(err)
			}
			got, err = DecodeFlashFromQuery(url.Values{"flash": {string(b)}})
			if err != nil {
				t.Fatal(err)
			}
			if got.Msg == nil || *got.Msg != *msg {
				t.Fatalf("outer round trip = %#v", got.Msg)
			}
		})
	}
}

func TestMsg_invalid_fields_do_not_leave_partial_message(t *testing.T) {
	for _, raw := range []string{
		`{"t":"hello","l":123,"s":"proof"}`,
		`{"t":"hello","s":123}`,
		`{"t":"hello","s":false}`,
		`{"t":"hello","s":[]}`,
		`{"t":"hello","s":{}}`,
		`{"t":"hello","m":"invalid","s":"proof"}`,
		`{"t":"hello","s":"proof"`,
	} {
		if msg := DecodeMsg(raw); msg != nil {
			t.Fatalf("DecodeMsg(%s) = %#v", raw, msg)
		}
	}
}

func TestMsg_unsigned_json_stays_compatible(t *testing.T) {
	msg := RawSuccessMsg("hello")
	if got := msg.Encode(); got != `{"t":"hello","m":"success"}` {
		t.Fatalf("unsigned JSON = %s", got)
	}
	if got := (&Msg{}).Encode(); got != "" {
		t.Fatalf("empty message = %s", got)
	}
}
