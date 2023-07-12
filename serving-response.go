package mvp

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/hotwired"
	"github.com/andreyvit/mvp/sse"
)

type Redirect struct {
	Path       string
	StatusCode int
	Values     url.Values
}

// SameMethod uses 307 Temporary Redirect for this redirect.
// Same HTTP method will be used for the redirected request,
// unlike the default 303 See Other response which always redirects with a GET.
func (redir *Redirect) SameMethod() *Redirect {
	redir.StatusCode = http.StatusTemporaryRedirect
	return redir
}

// Permanent uses 308 Permanent Redirect for this redirect.
// Same HTTP method will be used for the redirected request.
//
// Note that there is no permanent redirection code that is guaranteed to
// always use GET. 301 Moved Permanently may or may not do that and is not
// recommended.
func (redir *Redirect) Permanent() *Redirect {
	redir.StatusCode = http.StatusPermanentRedirect
	return redir
}

type RawOutput struct {
	Data       []byte
	ContenType string
}

// DebugOutput can be returned by request handlers
type DebugOutput string

type ResponseHandled struct{}

func (rc *RC) SendTurboStream(id uint64, f func(stream *hotwired.Stream)) {
	// TODO: cache buffers
	var stream hotwired.Stream
	f(&stream)
	if stream.Buffer.Len() > 0 {
		msg := sse.Msg{
			ID:    id,
			Event: "message",
			Data:  stream.Buffer.Bytes(),
		}
		flogger.Log(rc, "Sending Turbo Stream %d: %s", msg.ID, msg.Data)
		rc.SendSSE(&msg)
	}
}

func (rc *RC) SendSSE(msg *sse.Msg) {
	rc.RespWriter.Header().Set("Content-Type", sse.ContentType)

	_, err := rc.RespWriter.Write(msg.Encode(nil))
	if err != nil {
		rc.Fail(ClientNetworkingError(err))
	}
	rc.RespWriter.(http.Flusher).Flush()
}

func (app *App) writeResponseExtras(rc *RC, w http.ResponseWriter, r *http.Request) {
	for _, cookie := range rc.SetCookies {
		http.SetCookie(w, cookie)
	}
}

func (app *App) writeResponse(rc *RC, output any, w http.ResponseWriter, r *http.Request) error {
	app.writeResponseExtras(rc, w, r)
	switch output := output.(type) {
	case *ViewData:
		if output.View == "" {
			output.View = strings.ReplaceAll(rc.Route.routeName, ".", "-")
		}
		app.fillViewData(output, rc)
		b, err := app.Render(rc, output)
		if err != nil {
			return err
		}
		if output.ContentType != "" {
			w.Header().Set("Content-Type", output.ContentType)
		}
		w.Write(b)
	case *Redirect:
		path := output.Path
		if len(output.Values) > 0 {
			path = path + "?" + output.Values.Encode()
		}
		code := output.StatusCode
		if code == 0 {
			code = http.StatusSeeOther
		}
		http.Redirect(w, r, path, code)
	case DebugOutput:
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(output))
	case ResponseHandled:
		break
	default:
		panic(fmt.Errorf("%s: invalid return value %T %v", rc.Route.desc, output, output))
	}
	return nil
}

func (app *App) fillViewData(output *ViewData, rc *RC) {
	output.RC = BaseRC.AnyFull(rc)
	output.baseRC = rc
	output.App = app
	output.Route = rc.Route
}
