package forms

import (
	"net/url"
	"testing"
	"time"
)

func TestOptionalInputTime_zero_binding_submits_as_default_mode(t *testing.T) {
	stored := time.Date(2022, time.May, 1, 10, 0, 0, 0, time.UTC)
	form, binding := buildOptionalInputTimeForm(&stored, time.Time{})

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"default"},
			"upgrade[time]": {""},
		},
	})

	if !ok {
		t.Fatalf("expected Process to return true, got invalid")
	}
	if !binding.Value().IsZero() {
		t.Fatalf("expected binding to be zero after default submit, got %v", binding.Value())
	}
}

func TestOptionalInputTime_custom_mode_parses_valid_date(t *testing.T) {
	var stored time.Time
	form, binding := buildOptionalInputTimeForm(&stored, time.Time{})

	submitted := "2022-06-15T09:30"
	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"custom"},
			"upgrade[time]": {submitted},
		},
	})

	if !ok {
		t.Fatalf("expected Process to return true, got invalid")
	}
	expected, _ := time.ParseInLocation(DateTimeLocalInputTimeFormat, submitted, time.UTC)
	if !binding.Value().Equal(expected) {
		t.Fatalf("expected binding %v, got %v", expected, binding.Value())
	}
}

func TestOptionalInputTime_custom_mode_with_empty_time_is_invalid(t *testing.T) {
	var stored time.Time
	form, _ := buildOptionalInputTimeForm(&stored, time.Time{})

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"custom"},
			"upgrade[time]": {""},
		},
	})

	if ok {
		t.Fatalf("expected Process to return false when custom mode has empty time")
	}
}

func TestOptionalInputTime_custom_mode_with_malformed_time_is_invalid(t *testing.T) {
	var stored time.Time
	form, _ := buildOptionalInputTimeForm(&stored, time.Time{})

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"custom"},
			"upgrade[time]": {"not-a-time"},
		},
	})

	if ok {
		t.Fatalf("expected Process to return false when custom mode has malformed time")
	}
}

func TestOptionalInputTime_default_mode_ignores_time_input(t *testing.T) {
	stored := time.Date(2022, time.May, 1, 10, 0, 0, 0, time.UTC)
	form, binding := buildOptionalInputTimeForm(&stored, time.Time{})

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"default"},
			"upgrade[time]": {"2099-12-31T23:59"},
		},
	})

	if !ok {
		t.Fatalf("expected Process to return true, got invalid")
	}
	if !binding.Value().IsZero() {
		t.Fatalf("expected binding to be zero after default submit, got %v", binding.Value())
	}
}

func TestOptionalInputTime_default_option_label_includes_formatted_default(t *testing.T) {
	var stored time.Time
	defaultValue := time.Date(2022, time.February, 1, 12, 0, 0, 0, time.UTC)
	control := &OptionalInputTime{
		Binding:      Var(&stored),
		Location:     time.UTC,
		IncludeTime:  true,
		DefaultValue: defaultValue,
	}
	form := &Form{}
	form.AddChild(&Item{Name: "upgrade", Child: control})
	form.FinalizeForm(nil)

	got := control.DefaultOptionLabel()
	expected := "Default — 2022-02-01 12:00"
	if got != expected {
		t.Fatalf("expected DefaultOptionLabel %q, got %q", expected, got)
	}
}

func TestOptionalInputTime_default_option_label_without_default_value(t *testing.T) {
	var stored time.Time
	control := &OptionalInputTime{
		Binding:     Var(&stored),
		Location:    time.UTC,
		IncludeTime: true,
	}
	form := &Form{}
	form.AddChild(&Item{Name: "upgrade", Child: control})
	form.FinalizeForm(nil)

	got := control.DefaultOptionLabel()
	if got != "Default" {
		t.Fatalf("expected DefaultOptionLabel %q, got %q", "Default", got)
	}
}

func TestOptionalInputTime_initial_mode_follows_binding_value(t *testing.T) {
	zeroStored := time.Time{}
	nonZeroStored := time.Date(2022, time.May, 1, 10, 0, 0, 0, time.UTC)

	zeroControl := &OptionalInputTime{
		Binding:     Var(&zeroStored),
		Location:    time.UTC,
		IncludeTime: true,
	}
	zeroForm := &Form{}
	zeroForm.AddChild(&Item{Name: "upgrade", Child: zeroControl})
	zeroForm.FinalizeForm(nil)
	if !zeroControl.IsDefaultSelected() {
		t.Fatalf("expected zero binding to render default mode selected")
	}

	nonZeroControl := &OptionalInputTime{
		Binding:     Var(&nonZeroStored),
		Location:    time.UTC,
		IncludeTime: true,
	}
	nonZeroForm := &Form{}
	nonZeroForm.AddChild(&Item{Name: "upgrade", Child: nonZeroControl})
	nonZeroForm.FinalizeForm(nil)
	if nonZeroControl.IsDefaultSelected() {
		t.Fatalf("expected non-zero binding to render custom mode selected")
	}
}

func TestOptionalInputTime_preserves_raw_time_on_invalid_round_trip(t *testing.T) {
	var stored time.Time
	form, _ := buildOptionalInputTimeForm(&stored, time.Time{})

	form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"upgrade[mode]": {"custom"},
			"upgrade[time]": {"not-a-time"},
		},
	})

	control := findOptionalInputTime(form)
	if control.TimeFormattedValue() != "not-a-time" {
		t.Fatalf("expected raw time %q preserved on invalid round-trip, got %q", "not-a-time", control.TimeFormattedValue())
	}
	if control.IsDefaultSelected() {
		t.Fatalf("expected mode to stay custom after invalid submission")
	}
}

func buildOptionalInputTimeForm(stored *time.Time, defaultValue time.Time) (*Form, *Binding[time.Time]) {
	binding := Var(stored)
	control := &OptionalInputTime{
		Binding:      binding,
		Location:     time.UTC,
		IncludeTime:  true,
		DefaultValue: defaultValue,
	}
	form := &Form{}
	form.AddChild(&Item{Name: "upgrade", Child: control})
	return form, binding
}

func findOptionalInputTime(form *Form) *OptionalInputTime {
	var found *OptionalInputTime
	walk(&form.Group, 0, func(c Child) {
		if control, ok := c.(*OptionalInputTime); ok {
			found = control
		}
	}, nil)
	return found
}
