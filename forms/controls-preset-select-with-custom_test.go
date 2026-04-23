package forms

import (
	"net/url"
	"testing"
)

func TestPresetSelectWithCustom_stores_selected_preset(t *testing.T) {
	stored := "old"
	form, binding, _ := buildPresetSelectWithCustomForm(&stored)

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"setting[preset]": {"preset"},
			"setting[custom]": {"typed"},
		},
	})

	if !ok {
		t.Fatalf("expected Process to return true, got invalid")
	}
	if binding.Value() != "preset-value" {
		t.Fatalf("expected preset value, got %q", binding.Value())
	}
}

func TestPresetSelectWithCustom_processes_custom_child_when_selected(t *testing.T) {
	stored := "old"
	form, binding, _ := buildPresetSelectWithCustomForm(&stored)

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"setting[preset]": {"custom"},
			"setting[custom]": {"typed"},
		},
	})

	if !ok {
		t.Fatalf("expected Process to return true, got invalid")
	}
	if binding.Value() != "typed" {
		t.Fatalf("expected custom value, got %q", binding.Value())
	}
}

func TestPresetSelectWithCustom_update_skips_newly_visible_custom_child(t *testing.T) {
	stored := "old"
	form, binding, _ := buildRequiredPresetSelectWithCustomForm(&stored)

	ok := form.Process(&FormData{
		Action: "update",
		Values: url.Values{
			"setting[preset]": {"custom"},
		},
	})

	if ok {
		t.Fatalf("expected update action to return false")
	}
	if form.Invalid() {
		t.Fatalf("expected newly visible custom child to remain unvalidated")
	}
	if binding.Value() != "old" {
		t.Fatalf("expected stored value to remain unchanged, got %q", binding.Value())
	}
}

func TestPresetSelectWithCustom_submit_validates_missing_custom_child(t *testing.T) {
	stored := "old"
	form, _, _ := buildRequiredPresetSelectWithCustomForm(&stored)

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"setting[preset]": {"custom"},
		},
	})

	if ok {
		t.Fatalf("expected missing required custom value to fail")
	}
	if !form.Invalid() {
		t.Fatalf("expected missing required custom value to invalidate form")
	}
}

func TestPresetSelectWithCustom_selects_custom_for_unmatched_value(t *testing.T) {
	stored := "typed"
	form, _, control := buildPresetSelectWithCustomForm(&stored)
	form.FinalizeForm(nil)

	if !control.IsCustomSelected() {
		t.Fatalf("expected unmatched value to select custom")
	}
}

func TestPresetSelectWithCustom_ignores_custom_option_model_value(t *testing.T) {
	stored := "custom-model-value"
	form, _, control := buildPresetSelectWithCustomForm(&stored)
	form.FinalizeForm(nil)

	if !control.IsCustomSelected() {
		t.Fatalf("expected last option model value to be ignored for preset matching")
	}
}

func buildPresetSelectWithCustomForm(stored *string) (*Form, *Binding[string], *PresetSelectWithCustom[string]) {
	binding := Var(stored)
	control := &PresetSelectWithCustom[string]{
		Binding: binding,
		Options: []*Option[string]{
			{ModelValue: "preset-value", HTMLValue: "preset", Label: "Preset"},
			{ModelValue: "custom-model-value", HTMLValue: "custom", Label: "Custom"},
		},
		Custom: &InputText{
			Binding: binding,
		},
	}
	form := &Form{}
	form.AddChild(&Item{Name: "setting", Child: control})
	return form, binding, control
}

func buildRequiredPresetSelectWithCustomForm(stored *string) (*Form, *Binding[string], *PresetSelectWithCustom[string]) {
	form, binding, control := buildPresetSelectWithCustomForm(stored)
	control.Custom = &InputText{
		Binding:  binding,
		Required: true,
	}
	return form, binding, control
}
