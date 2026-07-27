package forms

import (
	"net/url"
	"slices"
	"testing"
)

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
