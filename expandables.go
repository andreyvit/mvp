package mvp

import "github.com/andreyvit/mvp/expandable"

var (
	baseSchema   = expandable.NewSchema("mvp")
	BaseSettings = expandable.NewBase[Settings](baseSchema, "settings")
	BaseApp      = expandable.NewBase[App](baseSchema, "app")
	BaseRC       = expandable.NewBase[RC](baseSchema, "rc")
)
