package mvputil

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func VerifyBasicAuth(w http.ResponseWriter, r *http.Request, lookupUserPasswordHash func(username string) (canonicalUsername, passwordHash string)) (username string, authenticated bool) {
	{
		u, pw, ok := r.BasicAuth()
		if !ok || u == "" {
			goto fail
		}

		u, correctPwHash := lookupUserPasswordHash(u)
		if correctPwHash == "" {
			goto fail
		}

		if err := bcrypt.CompareHashAndPassword([]byte(correctPwHash), []byte(pw)); err != nil {
			goto fail
		}

		return u, true
	}

fail:
	w.Header().Set("WWW-Authenticate", `Basic realm="Login Required", charset="utf-8"`)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return "", false
}
