package cors

import (
	"net/http"
	"strconv"
	"time"
)

// CORS handles details for Cross-Origin Resource Sharing.
// See https://www.w3.org/TR/cors/
type CORS struct {
	CacheDuration   time.Duration
	Origins         []string
	RequestHeaders  string
	ResponseHeaders string
}

const (
	wildcardCORSOrigin      = "*"
	corsRequestMethodHeader = "Access-Control-Request-Method"
	corsOriginHeader        = "Origin"
	corsAllowOriginHeader   = "Access-Control-Allow-Origin"
)

func (c CORS) Wrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.Handle(w, r) {
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Handle returns true if the request is entirely handled (i.e. is a
// CORS options request), and false if further handling needs to
// continue.
func (c CORS) Handle(w http.ResponseWriter, r *http.Request) bool {
	// log.Printf("CORS handle with %v", r.Header)
	if r.Method == http.MethodOptions && r.Header.Get(corsRequestMethodHeader) != "" {
		c.HandlePreflight(w, r)
		return true
	}

	c.AddCORSHeaders(r, w)
	return false
}

func (c CORS) Preflight() http.Handler {
	return http.HandlerFunc(c.HandlePreflight)
}

func (c CORS) HandlePreflight(w http.ResponseWriter, r *http.Request) {
	msg, status := c.computePreflightResponse(r.Header.Get(corsOriginHeader), r.Header.Get(corsRequestMethodHeader), w.Header())
	// log.Printf("CORS preflight => %d %s", status, msg)
	if status == http.StatusOK {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, msg, status)
	}
}

func (c CORS) AddCORSHeaders(r *http.Request, w http.ResponseWriter) {
	c.AddCORSHeadersTo(r.Header, w.Header())
}

func (c CORS) AddCORSHeadersTo(reqHeaders http.Header, respHeaders http.Header) {
	c.AddCORSHeadersOrigin(reqHeaders.Get(corsOriginHeader), respHeaders)
}

func (c CORS) AddCORSHeadersOrigin(origin string, headers http.Header) {
	if c.isWildcard() {
		headers.Set(corsAllowOriginHeader, wildcardCORSOrigin)
	} else if origin != "" && c.isOriginWhitelisted(origin) {
		headers.Set(corsAllowOriginHeader, origin)
		headers.Add("Vary", corsOriginHeader)
	}
	if c.ResponseHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", c.ResponseHeaders)
	}
}

func (c CORS) computePreflightResponse(origin, method string, headers http.Header) (string, int) {
	headers.Add("Vary", corsOriginHeader)
	headers.Add("Vary", corsRequestMethodHeader)

	if origin == "" {
		return "missing Origin HTTP header", http.StatusBadRequest
	}
	if method == "" {
		return "missing Access-Control-Request-Method HTTP header", http.StatusBadRequest
	}

	if c.isWildcard() {
		origin = wildcardCORSOrigin
	} else if !c.isOriginWhitelisted(origin) {
		return "forbidden origin", http.StatusForbidden
	}

	headers.Set(corsAllowOriginHeader, origin)
	headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH")
	if c.RequestHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", c.RequestHeaders)
	}
	if c.ResponseHeaders != "" {
		headers.Set("Access-Control-Expose-Headers", c.ResponseHeaders)
	}
	if c.CacheDuration > 0 {
		headers.Set("Access-Control-Max-Age", strconv.Itoa(int(c.CacheDuration/time.Second)))
	}

	return "", http.StatusOK
}

func (c CORS) isWildcard() bool {
	return c.Origins == nil
}

func (c CORS) isOriginWhitelisted(origin string) bool {
	for _, o := range c.Origins {
		if o == origin {
			return true
		}
	}
	return false
}
