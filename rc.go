package mvp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flake"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"github.com/uptrace/bunrouter"
)

// RC stands for Request Context, and holds all the things we want to associate
// with a request.
//
// RC carries:
//
//  1. Transaction
//  2. Authentication
//  3. context.Context
//  4. Logging context
//  5. Current HTTP request info
//  6. HTTP response side-channel bits (cookies at the moment)
//  7. Matched route or call info
//  8. Timing information
//  9. Anything else that middleware needs to read or write
//
// Why this instead of stuffing things into context.Context? Mostly because
// it's more explicit, typed, doesn't require O(N) lookup to have N values,
// and doesn't require an allocation per value. If anything, you'd put
// *RC into context.Context, not individual values, so you'd need this type anyway.
//
// RC doesn't just wrap HTTP requests; it is also used by jobs and one-off
// code. Similar to context.Context, it's a nice way to propagate transactions,
// authentication and other stuff throughout the routing/call handling machinery
// without every function knowing about and dealing with individual values.
//
// This type also carries any data you want to communicate between middleware
// and HTTP handlers. At the very least, it carries concrete authentication data,
// but also things like tenant ref in multi-tenant applications, etc.
// As one example, we want to be able to implement Rails-like “flash” middleware,
// so RC needs to carry input/output cookies AND any resulting flash data.
//
// As such, this type is meant to be customized for a particular application.
type RC struct {
	tx  *edb.Tx
	err error

	auth Auth

	parent context.Context
	values []any
	app    *App

	RequestID string
	Start     time.Time // ACTUAL time of request start
	Now       time.Time // wall clock time of request start
	logf      func(format string, args ...any)

	Route      *Route
	Request    bunrouter.Request
	RespWriter http.ResponseWriter

	SetCookies []*http.Cookie

	RateLimitPreset RateLimitPreset
	RateLimitKey    string

	extraLogger func(format string, args ...any)
}

func NewRC(ctx context.Context, app *App, requestID string) *RC {
	if requestID == "" {
		requestID = RandomHex(32)
	}
	rc := BaseRC.New()
	*rc = RC{
		parent:    ctx,
		values:    newValueSet(),
		app:       app,
		RequestID: requestID,
		Start:     time.Now(),
		Now:       time.Now(),
	}
	runHooksFwd2(app.Hooks.initRC, app, rc)
	return rc
}

func NewHTTPRC(app *App, w http.ResponseWriter, r bunrouter.Request) *RC {
	rc := NewRC(r.Context(), app, r.Header.Get("X-Request-ID"))
	rc.Request = r
	rc.RespWriter = w
	return rc
}

func (rc *RC) App() *App {
	return rc.app
}

func (rc *RC) BaseRC() *RC {
	return rc
}

func (rc *RC) SessionID() flake.ID {
	return rc.auth.SessionID
}

func (rc *RC) ActorRef() mvpm.Ref {
	return rc.auth.ActorRef
}

func (rc *RC) IsLoggedIn() bool {
	return !rc.auth.ActorRef.IsZero()
}

func (rc *RC) Auth() Auth {
	return rc.auth
}

func (rc *RC) BaseApp() *App {
	return rc.app
}

func (rc *RC) DBTx() *edb.Tx {
	if rc.tx == nil {
		panic("rc does not have tx open")
	}
	return rc.tx
}

func (rc *RC) Logf(format string, args ...any) {
	rc.app.logf(format, args...)
	if rc.extraLogger != nil {
		rc.extraLogger(format, args...)
	}
}

func (rc *RC) LogTo(buf *strings.Builder) {
	if buf == nil {
		rc.extraLogger = nil
	} else {
		rc.extraLogger = func(format string, args ...any) {
			s := fmt.Sprintf(format, args...)
			s = strings.TrimPrefix(s, fmt.Sprintf("[%s] ", rc.RequestID))
			buf.WriteString(s)
			if !strings.HasSuffix(s, "\n") {
				buf.WriteByte('\n')
			}
		}
	}
}

func (rc *RC) Close() {
	runHooksRev2(rc.app.Hooks.closeRC, rc.app, rc)
	// TODO: return allocated values to the pool
}

func (app *App) NewHTTPRequestRC(w http.ResponseWriter, r bunrouter.Request) *RC {
	rc := NewRC(r.Context(), app, r.Header.Get("X-Request-ID"))
	rc.Request = r
	rc.RespWriter = w
	return rc
}

func (rc *RC) Deadline() (deadline time.Time, ok bool) {
	return rc.parent.Deadline()
}
func (rc *RC) Done() <-chan struct{} {
	// TODO: arrange for Done to handle rc.err
	return rc.parent.Done()
}
func (rc *RC) Err() error {
	if rc.err != nil {
		return rc.err
	}
	return rc.parent.Err()
}
func (rc *RC) Value(key any) any {
	return rc.parent.Value(key)
}

func (rc *RC) SetCookie(cookie *http.Cookie) {
	rc.SetCookies = append(rc.SetCookies, cookie)
}

func (rc *RC) AppendLogPrefix(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "[%s] ", rc.RequestID)
}
func (rc *RC) AppendLogSuffix(buf *bytes.Buffer) {
}
