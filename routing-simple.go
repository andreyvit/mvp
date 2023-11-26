package mvp

import "net/http"

func IsGetOrHead(r *http.Request) bool {
	return r.Method == http.MethodGet || r.Method == http.MethodHead
}
func RequireGetOrHead(w http.ResponseWriter, r *http.Request) bool {
	if IsGetOrHead(r) {
		return true
	} else {
		WriteMethodNotAllowed(w)
		return false
	}
}
