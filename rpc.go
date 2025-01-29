package mvp

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/andreyvit/mvp/httperrors"
	"github.com/andreyvit/mvp/mvphttp"
)

func ReadAPIRequest(r *http.Request, in any) error {
	switch r.Method {
	case http.MethodPost:
		switch DetermineMIMEType(r) {
		case "", "application/json":
			return readJSONRequest(r, in)
		default:
			return ErrAPIUnsupportedContentType
		}
	default:
		return ErrAPIInvalidMethod
	}
}

func readJSONRequest(r *http.Request, in any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1024*1024))
	if r.Header.Get("X-Ignore-Unknown-Fields") != "yes" {
		decoder.DisallowUnknownFields()
	}

	err := decoder.Decode(in)
	if err != nil {
		return ErrAPIInvalidJSON.WrapMsg(err, err.Error())
	}
	return nil
}

func (app *App) WriteAPIError(w http.ResponseWriter, err error) {
	resp := BuildAPIErrorResponse(err)
	data := must(json.Marshal(resp))

	mvphttp.ApplyCacheMode(w, mvphttp.Uncached)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(resp.HTTPCode())
	w.Write(data)
}

func WriteAPIResponse(w http.ResponseWriter, out any, indented bool) {
	mvphttp.ApplyCacheMode(w, mvphttp.Uncached)
	w.Header().Set("Content-Type", "application/json")
	var err error
	if indented {
		var data []byte
		data, err = json.MarshalIndent(out, "", "  ")
		if err == nil {
			_, err = w.Write(data)
		}
	} else {
		err = json.NewEncoder(w).Encode(out)
	}
	if err != nil {
		log.Printf("ERROR: failed to encode response %T: %v", out, err)
	}
}

func BuildAPIErrorResponse(err error) httperrors.Interface {
	resp := &defaultAPIErrorResponse{
		statusCode: HTTPStatusCode(err),
	}
	if e, ok := err.(interface{ ErrorID() string }); ok {
		resp.ErrID = e.ErrorID()
	}
	if e, ok := err.(interface{ PublicError() string }); ok {
		resp.PubMsg = e.PublicError()
	}
	if e, ok := err.(interface{ InvalidFieldPath() string }); ok {
		resp.FieldPath = e.InvalidFieldPath()
	}
	if e, ok := err.(interface{ InvalidFieldPublicError() string }); ok {
		resp.FieldError = e.InvalidFieldPublicError()
	}

	if resp.ErrID == "" {
		if resp.statusCode >= 400 && resp.statusCode <= 499 {
			resp.ErrID = "bad_request"
		} else {
			resp.ErrID = "unavail"
		}
	}
	if resp.PubMsg == "" {
		if resp.statusCode >= 400 && resp.statusCode <= 499 {
			resp.PubMsg = "Your request appears invalid. Please try again later and contact support if the problem persists."
		} else {
			resp.PubMsg = "Failed to handle your request. Please try again later and contact support if the problem persists."
		}
	}
	return resp
}

type defaultAPIErrorResponse struct {
	statusCode int    `json:"-"`
	ErrID      string `json:"error"`
	PubMsg     string `json:"message"`
	FieldPath  string `json:"field_path,omitempty"`
	FieldError string `json:"field_err,omitempty"`
}

func (e *defaultAPIErrorResponse) Error() string {
	return e.ErrID
}
func (e *defaultAPIErrorResponse) HTTPCode() int {
	return e.statusCode
}
func (e *defaultAPIErrorResponse) ErrorID() string {
	return e.ErrID
}
func (e *defaultAPIErrorResponse) PublicError() string {
	return e.PubMsg
}
func (e *defaultAPIErrorResponse) ForeachExtra(f func(k string, v interface{})) {
	if e.FieldPath != "" {
		f("field_path", e.FieldPath)
	}
	if e.FieldError != "" {
		f("field_err", e.FieldError)
	}
}
