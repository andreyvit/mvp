package mvpstatics

import (
	"io/fs"
	"net/http"

	"github.com/andreyvit/mvp/cors"
	"github.com/andreyvit/mvp/mvphttp"
	"github.com/uptrace/bunrouter"
)

type filesOnlyFS struct {
	fs.FS
}

func (f filesOnlyFS) Open(name string) (fs.File, error) {
	file, err := f.FS.Open(name)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	if info.IsDir() {
		file.Close()
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return file, nil
}

func SetupRoute(g *bunrouter.Group, urlPrefix string, f fs.FS, cm mvphttp.CacheMode, cors *cors.CORS) {
	h := http.FileServer(http.FS(filesOnlyFS{f}))
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

	if cors != nil {
		g.OPTIONS(urlPrefix+"/*path", func(w http.ResponseWriter, req bunrouter.Request) error {
			h.ServeHTTP(w, req.Request)
			return nil
		})
	}
}
