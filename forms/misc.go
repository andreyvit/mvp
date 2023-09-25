package forms

import "net/http"

func StatusCode(isSaving bool) int {
	if isSaving {
		return http.StatusUnprocessableEntity
	} else {
		return http.StatusOK
	}
}
