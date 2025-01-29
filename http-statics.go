package mvp

import (
	"io/fs"
)

func (app *App) StaticFS() fs.FS {
	return app.staticFS
}
