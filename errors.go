package mvp

import (
	"fmt"
	"net/http"

	"github.com/andreyvit/mvp/httperrors"
)

var (
	ErrTooManyRequests = httperrors.Define(http.StatusTooManyRequests, "too_many_requests")
	ErrInvalidToken    = httperrors.Define(http.StatusUnauthorized, "invalid_token")
	ErrForbidden       = httperrors.Define(http.StatusForbidden, "forbidden")

	ErrAPIInvalidMethod          = httperrors.Define(http.StatusMethodNotAllowed, "invalid_http_method")
	ErrAPIUnsupportedContentType = httperrors.Define(http.StatusUnsupportedMediaType, "invalid_content_type")
	ErrAPIInvalidJSON            = httperrors.Define(http.StatusBadRequest, "invalid_json")
	ErrAPIUnknownMethod          = httperrors.Define(http.StatusNotFound, "unknown_method")

	ErrPanic = httperrors.Define(http.StatusInternalServerError, "internal_server_error")
)

func RecoveredError(e any) error {
	if err, ok := e.(error); ok {
		return ErrPanic.Wrap(err)
	} else {
		return ErrPanic.Wrap(fmt.Errorf("non-error panic: %#v", e))
	}
}

func HTTPStatusCode(err error) int {
	if e, ok := err.(interface{ HTTPCode() int }); ok {
		code := e.HTTPCode()
		if code != 0 {
			return code
		}
	}
	return http.StatusInternalServerError
}
