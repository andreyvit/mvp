package mvpstatics

import (
	"io/fs"
	"net/http"

	"github.com/andreyvit/mvp/cors"
	"github.com/andreyvit/mvp/mvphttp"
	"github.com/uptrace/bunrouter"
)

func SetupRoute(g *bunrouter.Group, urlPrefix string, f fs.FS, cm mvphttp.CacheMode, cors *cors.CORS) {
	h := http.FileServer(http.FS(f))
	// h = http.StripPrefix(urlPrefix, h)
	if cors != nil {
		h = cors.Wrap(h)
	}

	g.GET(urlPrefix+"/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
		mvphttp.ApplyCacheMode(w, cm)
		req.Request.URL.Path = "/" + req.Param("path")
		h.ServeHTTP(w, req.Request)
		return nil
	})
}
