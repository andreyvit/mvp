package mvputil

import (
	"net/http"
	"strings"
)

const AcceptHeader = "Accept"

func PreferJSON(r *http.Request) bool {
	accept := r.Header.Get(AcceptHeader)
	if strings.Contains(accept, "text/html") {
		return false
	}
	if strings.Contains(accept, "application/json") {
		return true
	}
	return false
}
