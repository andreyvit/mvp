package mvp

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/andreyvit/mvp/mvpfs"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

type FileProvider interface {
	String() string
	FileURL(path string, rc *RC, opt mvpm.URLOption) string
	ReadFile(path string) ([]byte, error)
}

func (app *App) FileProvider(resourceURI string, rc *RC) (FileProvider, string) {
	if resourceURI == "" {
		return nil, ""
	}

	// Note: accepting RC here because in the future, some schemes
	// might need RC to resolve. (Say, icons that are customizable
	// per account will require account data in RC to resolve.)

	scheme, path := mvpfs.Split(resourceURI)
	switch scheme {
	case "http", "https":
		return urlProvider{app}, resourceURI
	case mvpfs.StaticScheme, "":
		return staticProvider{app}, path
	default:
		panic(fmt.Errorf("unknown scheme in file URI %q", resourceURI))
	}
}

func (app *App) FileURL(resourceURI string, rc *RC, opt mvpm.URLOption) string {
	if resourceURI == "" {
		return ""
	}
	provider, path := app.FileProvider(resourceURI, rc)
	return provider.FileURL(path, rc, opt)
}

func (app *App) ReadFile(resourceURI string, rc *RC) ([]byte, error) {
	if resourceURI == "" {
		return nil, nil
	}
	provider, path := app.FileProvider(resourceURI, rc)
	return provider.ReadFile(path)
}

func (rc *RC) FileURL(resourceURI string, opt mvpm.URLOption) string {
	return rc.app.FileURL(resourceURI, rc, opt)
}

func (rc *RC) ReadFile(resourceURI string) ([]byte, error) {
	return rc.app.ReadFile(resourceURI, rc)
}

type urlProvider struct {
	app *App
}

func (prov urlProvider) String() string {
	return "static"
}

func (prov urlProvider) FileURL(path string, rc *RC, opt mvpm.URLOption) string {
	return path
}

func (prov urlProvider) ReadFile(path string) ([]byte, error) {
	return nil, fmt.Errorf("cannot read file data from URL")
}

type staticProvider struct {
	app *App
}

func (prov staticProvider) String() string {
	return "static"
}

func (prov staticProvider) FileURL(path string, rc *RC, opt mvpm.URLOption) string {
	urlPath := "/static/" + path
	if opt.Contains(Absolute) {
		// TODO
	}
	return urlPath
}

func (prov staticProvider) ReadFile(path string) ([]byte, error) {
	data, err := fs.ReadFile(prov.app.staticFS, path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%q does not exist under %s/", path, prov.app.Configuration.StaticSubdir)
		} else {
			return nil, fmt.Errorf("static/%s: %w", path, err)
		}
	}
	return data, nil
}
