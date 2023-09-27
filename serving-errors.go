package mvp

import (
	"errors"
	"fmt"

	"github.com/andreyvit/mvp/flogger"
)

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
