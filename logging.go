package mvp

import (
	"log"
	"net/http"
	"time"

	"github.com/andreyvit/mvp/httperrors"
)

func logRequest(rc *RC, r *http.Request, err error) {
	durus := time.Since(rc.Start).Microseconds()

	verb, path := r.Method, r.URL.Path

	var statusCode int
	var statusMsg string
	var prefix string
	var errorDetails string
	if err == nil {
		statusCode, statusMsg = 200, "OK"
	} else {
		statusCode = httperrors.HTTPCode(err)
		if e, ok := err.(interface{ ErrorID() string }); ok {
			statusMsg = e.ErrorID()
		} else {
			statusMsg = "error"
		}
		prefix = "NOTICE: "
		errorDetails = " — " + err.Error()
	}

	log.Printf("%sHTTP %d %s: %s %s (%06dus)%s", prefix, statusCode, statusMsg, verb, path, durus, errorDetails)
}
