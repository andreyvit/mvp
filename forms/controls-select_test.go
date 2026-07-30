package forms

import (
	"net/url"
	"slices"
	"testing"
)

func TestSelect_Finalize_loads_dynamic_options_before_selecting_value(t *testing.T) {
	stored := "never"
	control := &Select[string]{
		Binding: Var(&stored),
		OptionsFunc: func() []*Option[string] {
			return []*Option[string]{
				{ModelValue: "unconfigured", HTMLValue: "unconfigured", Label: "Default"},
				{ModelValue: "never", HTMLValue: "never", Label: "Never"},
			}
		},
	}
	form := &Form{}
	form.AddChild(&Item{Name: "expiration", Child: control})

	form.FinalizeForm(nil)

	if !control.IsHTMLValueSelected("never") {
		t.Fatalf("expected stored value to select dynamic option, got raw value %q", control.RawFormValue)
	}
}

func TestRawMultiSelect_Process(t *testing.T) {
	tests := []struct {
		name     string
		posted   url.Values
		expected []string
	}{
		{
			name:     "guard without values clears the selection",
			posted:   url.Values{"items[present]": {"1"}},
			expected: []string{},
		},
		{
			name:     "missing guard preserves the selection",
			posted:   url.Values{},
			expected: []string{"old"},
		},
		{
			name:     "guard with values replaces the selection",
			posted:   url.Values{"items[present]": {"1"}, "items": {"new-1", "new-2"}},
			expected: []string{"new-1", "new-2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stored := []string{"old"}
			form := &Form{}
			form.AddChild(&Item{
				Name: "items",
				Child: &RawMultiSelect[string]{
					Binding: Var(&stored),
					Parse: func(value string) (string, error) {
						return value, nil
					},
				},
			})

			ok := form.Process(&FormData{Action: "submit", Values: tt.posted})
			if !ok {
				t.Fatalf("expected Process to return true, got invalid")
			}
			if !slices.Equal(stored, tt.expected) {
				t.Fatalf("expected stored values %q, got %q", tt.expected, stored)
			}
		})
	}
}
