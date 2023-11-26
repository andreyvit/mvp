package mvp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/httperrors"
)

func WritePlainError(w http.ResponseWriter, err error) {
	code := httperrors.HTTPCode(err)
	message := httperrors.HTTPMessage(err)
	DisableCaching(w)
	http.Error(w, message, code)
}

func WriteMethodNotAllowed(w http.ResponseWriter) {
	DisableCaching(w)
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
}

type clientNetworkingError struct {
	Cause error
}

func (err clientNetworkingError) Error() string {
	return fmt.Sprintf("client networking: %v", err.Cause)
}

func (err clientNetworkingError) Unwrap() error {
	return err
}

func ClientNetworkingError(err error) error {
	if err == nil {
		return nil
	}
	return clientNetworkingError{err}
}

func IsClientNetworkingError(err error) bool {
	var dummy *clientNetworkingError
	return errors.As(err, &dummy)
}

func (rc *RC) Fail(err error) {
	if err == nil {
		return
	}
	if rc.err == nil {
		rc.err = err
	}
	if IsClientNetworkingError(err) {
		flogger.Log(rc, "NOTICE: %v", err)
	} else {
		flogger.Log(rc, "WARNING: %v", err)
	}
}

type RenderingError struct {
	Kind     string
	Template string
	Cause    error
}

func ViewRenderingError(templ string, cause error) error {
	return &RenderingError{"view", templ, cause}
}
func PartialRenderingError(templ string, cause error) error {
	return &RenderingError{"partial", templ, cause}
}

func (err *RenderingError) Error() string {
	return fmt.Sprintf("rendering %v %s: %v", err.Kind, err.Template, err.Cause)
}

func (err *RenderingError) SkipStackOnPanic() bool {
	return true
}
