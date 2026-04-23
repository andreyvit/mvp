package forms

import (
	"testing"
	"time"
)

func TestInputTime_empty_if_zero(t *testing.T) {
	var stored time.Time
	input := &InputTime{
		Binding:     Var(&stored),
		Location:    time.UTC,
		IncludeTime: true,
	}

	if input.FormattedValue() != "" {
		t.Fatalf("expected zero time to render blank, got %q", input.FormattedValue())
	}

	input.RawFormValue = "not-a-time"
	input.RawFormValuePresent = true
	if input.FormattedValue() != "not-a-time" {
		t.Fatalf("expected raw value to win, got %q", input.FormattedValue())
	}
}
