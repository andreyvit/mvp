package mvp

import (
	"io/fs"
	"net/http"

	"github.com/uptrace/bunrouter"
)

func (app *App) StaticFS() fs.FS {
	return app.staticFS
}

func setupStaticServer(g *bunrouter.Group, urlPrefix string, f fs.FS) {
	h := http.FileServer(http.FS(f))
	h = http.StripPrefix(urlPrefix, h)

	g.GET(urlPrefix+"/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
		MarkPrivateMutable(w)
		h.ServeHTTP(w, req.Request)
		return nil
	})
}
