package forms

import (
	"testing"
)

func TestJoinNames(t *testing.T) {
	tests := []struct {
		first    string
		second   string
		expected string
	}{
		{"", "", ""},
		{"", "foo", "foo"},

		{"foo", "", "foo"},
		{"foo", "bar", "foo[bar]"},
		{"foo", "[bar]", "foo[bar]"},
		{"foo", "bar[boz]", "foo[bar][boz]"},
		{"foo", "[bar][boz]", "foo[bar][boz]"},

		{"foo[bar]", "boz", "foo[bar][boz]"},
		{"foo[bar]", "[boz]", "foo[bar][boz]"},
		{"foo[bar]", "boz[fubar]", "foo[bar][boz][fubar]"},
		{"foo[bar]", "[boz][fubar]", "foo[bar][boz][fubar]"},
	}
	for _, tt := range tests {
		actual := JoinNames(tt.first, tt.second)
		if actual != tt.expected {
			t.Errorf("** JoinNames(%q, %q) == %q, expected %q", tt.first, tt.second, actual, tt.expected)
		} else {
			t.Logf("✓ JoinNames(%q, %q) == %q", tt.first, tt.second, actual)
		}
	}
}
