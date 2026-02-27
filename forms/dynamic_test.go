package forms

import (
	"net/url"
	"testing"
)

func TestDynamicChild_resolves_child_with_typed_field_name(t *testing.T) {
	var value string
	form := Form{}
	form.AddChild(&Item{
		Name:  "wrapper",
		Child: &DynamicChild{
			Name: "value",
			Resolver: func() (string, Child) {
				return "Text", &InputText{
					Binding: &Binding[string]{
						Getter: func() string { return value },
						Setter: func(v string) error { value = v; return nil },
					},
				}
			},
		},
	})

	ok := form.Process(&FormData{
		Action: "submit",
		Values: url.Values{
			"wrapper[value:Text]": {"hello"},
		},
	})

	if !ok {
		t.Fatal("expected Process to return true")
	}
	if value != "hello" {
		t.Fatalf("expected value %q, got %q", "hello", value)
	}
}

func TestDynamicChild_empty_type_creates_no_child(t *testing.T) {
	form := Form{}
	form.AddChild(&Item{
		Name:  "wrapper",
		Child: &DynamicChild{
			Name: "value",
			Resolver: func() (string, Child) {
				return "", nil
			},
		},
	})

	form.FinalizeForm(nil)

	if len(form.fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(form.fields))
	}
}

func TestDynamicChild_nil_resolver(t *testing.T) {
	form := Form{}
	form.AddChild(&Item{
		Name:  "wrapper",
		Child: &DynamicChild{
			Name:     "value",
			Resolver: nil,
		},
	})

	form.FinalizeForm(nil)

	if len(form.fields) != 0 {
		t.Fatalf("expected no fields, got %d", len(form.fields))
	}
}
