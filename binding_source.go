package mvp

import "embed"

// ResponseBindingSource contains the native response declarations for host bindings.
//
//go:embed serving-response.go
var ResponseBindingSource embed.FS

// URLBindingSource contains the native URL option declarations for host bindings.
//
//go:embed routing-urlgen.go
var URLBindingSource embed.FS
