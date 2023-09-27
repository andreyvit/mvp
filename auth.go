package mvp

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/httperrors"
	"github.com/andreyvit/mvp/jwt"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
	"golang.org/x/exp/slices"
)

type Auth struct {
	SessionID flake.ID
	ActorRef  mvpm.Ref
}

// func (app *App) SetAuthCookie(rc *RC, c jwt.Claims, validity time.Duration) {
// 	rc.SetCookie(app.makeAuthCookie(token, validity))
// }

func (app *App) DecodeAuthToken(rc *RC, token string) error {
	// flogger.Log(rc, "DecodeAuthToken: %q", token)
	runHooksFwd2(app.Hooks.resetAuth, app, rc)

	c := TokenDecoding{
		Now: rc.Now(),
	}
	err := c.Token.ParseString(token)
	if err != nil {
		return ErrInvalidToken.Msg(err.Error())
	}
	c.Claims = c.Token.Claims()
	c.KeyID = c.Token.KeyID()
	c.Issuer = c.Claims.Issuer()

	err = runHooksRevBltin2EUntil(app.Hooks.jwtTokenKey, builtinParseJWTToken, rc, &c, &c.keyed)
	if err != nil {
		return ErrInvalidToken.Msg(err.Error())
	}
	if !c.keyed {
		return ErrInvalidToken.Msg("unsupported JWT token")
	}
	if !c.authed {
		return ErrInvalidToken.Msg("unauthenticated JWT token")
	}

	err = runHooksFwd2E(app.Hooks.postAuth, app, rc)
	if err != nil {
		return ErrInvalidToken.WrapMsg(err, "the token is no longer valid")
	}

	return nil
}

type TokenDecoding struct {
	Now    time.Time
	Token  jwt.Token
	Claims jwt.Claims
	KeyID  string
	Issuer string
	keyed  bool
	authed bool
}

func (c *TokenDecoding) DecodeHS256(key []byte) error {
	c.keyed = true
	if len(key) < jwt.MinHS256KeyLen {
		panic(fmt.Errorf("HS256 key too short: %d bytes", len(key)))
	}
	err := c.Token.ValidateHS256(key)
	if err != nil {
		return err
	}
	err = c.Claims.ValidateTimeAt(10*time.Second, c.Now)
	if err != nil {
		return err
	}
	return nil
}

func (c *TokenDecoding) SetAuth(rc *RC, auth Auth) {
	rc.auth = auth
	c.keyed = true
	c.authed = true
}

var (
	errUnknownTokenKeyID   = errors.New("unknown key ID")
	errInvalidTokenSubject = errors.New("invalid token subject")
	errInvalidTokenID      = errors.New("invalid token ID")
)

func builtinParseJWTToken(rc *RC, c *TokenDecoding) error {
	settings := rc.app.Settings
	if !slices.Contains(settings.JWTIssuers, c.Issuer) {
		return nil
	}
	ks := settings.Configuration.AuthTokenKeys
	key := ks.Keys[c.KeyID]
	if key == nil {
		return errUnknownTokenKeyID
	}
	if err := c.DecodeHS256(key); err != nil {
		return err
	}

	subj := c.Claims.Subject()
	ref, err := mvpm.ParseRef(subj)
	if err != nil {
		return errInvalidTokenSubject
	}
	auth := Auth{
		ActorRef: ref,
	}
	if tid := c.Claims.TokenID(); tid != "" {
		sessID, err := flake.Parse(tid)
		if err != nil {
			return errInvalidTokenID
		}
		auth.SessionID = sessID
	}
	c.SetAuth(rc, auth)
	return nil
}

func (app *App) MakeAuthToken(sessID flake.ID, actorRef mvpm.Ref, validity time.Duration) string {
	var subject string
	if !actorRef.IsZero() {
		subject = actorRef.String()
	}
	c := jwt.NewAt(subject, validity, app.Now())
	c[jwt.Issuer] = app.Settings.JWTIssuers[0]
	if sessID != 0 {
		c[jwt.TokenID] = sessID.String()
	}
	ks := app.Settings.Configuration.AuthTokenKeys
	c[jwt.KeyID] = ks.ActiveKeyName
	return jwt.SignHS256String(c, nil, ks.ActiveKey())
}

func (rc *RC) SetAuthUsingCookie(auth Auth) {
	app := rc.app
	runHooksFwd2(app.Hooks.resetAuth, app, rc)
	rc.auth = auth
	err := runHooksFwd2E(app.Hooks.postAuth, app, rc)
	if err != nil {
		panic(fmt.Errorf("attempt to set auth cookie with invalid Auth: %v", err))
	}
	rc.SetAuthCookie(rc.app.MakeAuthToken(auth.SessionID, auth.ActorRef, jwt.Forever), 365*24*time.Hour)
}

func (rc *RC) SetAuthCookie(token string, validity time.Duration) {
	rc.SetCookie(rc.app.makeAuthCookie(token, validity))
}
func (rc *RC) DeleteAuthCookie() {
	rc.SetCookie(rc.app.makeAuthCookie("", 0))
}
func (app *App) makeAuthCookie(value string, maxAge time.Duration) *http.Cookie {
	c := &http.Cookie{
		Name:     app.Configuration.AuthTokenCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge / time.Second),
		Secure:   !app.Settings.AllowInsecureHttp,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if c.MaxAge == 0 {
		c.MaxAge = -1
	}
	return c
}

func (app *App) authenticateRequest(rc *RC) error {
	if authzn := rc.Request.Header.Get("Authorization"); authzn != "" {
		method, param := parseAuthorizationHeader(authzn)
		switch method {
		case "bearer":
			return app.DecodeAuthToken(rc, param)
		default:
			return httperrors.Errorf(400, "invalid_authorization", "Invalid Authorization header value")
		}
	}

	// the only possible error is ErrNoCookie
	tokenCookie, _ := rc.Request.Cookie(app.Configuration.AuthTokenCookieName)
	if tokenCookie != nil {
		err := app.DecodeAuthToken(rc, tokenCookie.Value)
		if err != nil {
			rc.DeleteAuthCookie()
		}
		return err
	}

	return nil
}

func (app *App) AuthenticateRequestMiddleware(rc *RC) (any, error) {
	return nil, app.authenticateRequest(rc)
}

func parseAuthorizationHeader(authorization string) (method string, param string) {
	if len(authorization) == 0 {
		return "", ""
	}
	authorization = strings.TrimSpace(authorization)
	if len(authorization) == 0 {
		return "", ""
	}
	method, param, ok := strings.Cut(authorization, " ")
	if !ok {
		return "", ""
	}
	return strings.ToLower(strings.TrimSpace(method)), strings.TrimSpace(param)
}
