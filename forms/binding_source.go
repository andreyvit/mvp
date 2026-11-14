package forms

import "embed"

// BindingSource contains the native declarations for source-driven bindings.
//
//go:embed binding.go controls-file.go controls-preset-select-with-custom.go controls-select.go controls-time.go controls.go dynamic.go embeds.go forms.go list.go misc.go names.go renderer.go state.go util.go validation.go validationerrs.go validationtools.go
var BindingSource embed.FS
