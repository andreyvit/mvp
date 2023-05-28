package mvpfs

import (
	"strings"
)

const (
	StaticScheme   = "static"
	UploadedScheme = "uploaded"
)

func Split(resourceURI string) (scheme string, path string) {
	i := strings.IndexByte(resourceURI, ':')
	if i < 0 {
		return StaticScheme, resourceURI
	} else {
		return resourceURI[:i], resourceURI[i+1:]
	}
}

func StaticURI(path string) string {
	if path == "" {
		return ""
	}
	path = strings.TrimPrefix(path, "/")
	return StaticScheme + ":" + path
}

func UploadedURI(name string) string {
	if name == "" {
		return ""
	}
	return UploadedScheme + ":" + name
}
