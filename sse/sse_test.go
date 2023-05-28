package sse

import (
	"math"
	"testing"
)

func TestMsgEncode(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		event    string
		data     string
		expected string
	}{
		{"empty", "", "", "", "data: \n\n"},
		{"simple data", "", "", "foo", "data: foo\n\n"},
		{"multiline data", "", "", "foo\nbar\nboz", "data: foo\ndata: bar\ndata: boz\n\n"},
		{"trailing newline", "", "", "foo\n", "data: foo\ndata: \n\n"},
		{"multiline data with trailing newline", "", "", "foo\nbar\nboz\n", "data: foo\ndata: bar\ndata: boz\ndata: \n\n"},
		{"event, simple data", "", "test", "foo", "event: test\ndata: foo\n\n"},
		{"id, event, simple data", "123", "test", "foo", "id: 123\nevent: test\ndata: foo\n\n"},
		{"id, event, multiline data", "123", "test", "foo\nbar\nboz", "id: 123\nevent: test\ndata: foo\ndata: bar\ndata: boz\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Msg{tt.id, 0, tt.event, []byte(tt.data)}
			actual := msg.String()
			if actual != tt.expected {
				t.Errorf("** Encode(%q, %q) == %q, expected %q", tt.event, tt.data, actual, tt.expected)
			}
		})
	}
}

func TestMsgEncode_numeric_ids(t *testing.T) {
	tests := []struct {
		name     string
		id       uint64
		event    string
		data     string
		expected string
	}{
		{"empty", 0, "", "", "data: \n\n"},
		{"id, simple data", 424242, "", "foo", "id: 424242\ndata: foo\n\n"},
		{"id, event, simple data", 424242, "test", "foo", "id: 424242\nevent: test\ndata: foo\n\n"},
		{"max uint, simple data", math.MaxUint64, "", "foo", "id: 18446744073709551615\ndata: foo\n\n"},
		{"max uint, event, simple data", math.MaxUint64, "test", "foo", "id: 18446744073709551615\nevent: test\ndata: foo\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Msg{"", tt.id, tt.event, []byte(tt.data)}
			actual := msg.String()
			if actual != tt.expected {
				t.Errorf("** Encode(%q, %q) == %q, expected %q", tt.event, tt.data, actual, tt.expected)
			}
		})
	}
}
