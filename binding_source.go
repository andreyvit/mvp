package mvp

import "embed"

// ResponseBindingSource contains the native response declarations for host bindings.
//
//go:embed serving-response.go
var ResponseBindingSource embed.FS
